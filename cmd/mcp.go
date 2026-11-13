package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// JSON-RPC 2.0 wire types.
type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func newMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start a stdio-based MCP server exposing Markov analysis as callable tools",
		Long: `Start a Model Context Protocol server over stdin/stdout (JSON-RPC 2.0).

Exposes one tool:

  analyze   Ingest a JSONL agent event stream and return ranked MFPT-reducing
            policy knobs. Equivalent to soka suggest.

Compatible with LangGraph, LlamaIndex, Claude Agent SDK, and any MCP client
that supports stdio transport.

Usage (stdio):
  soka mcp`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMCP()
		},
	}
}

func runMCP() error {
	enc := json.NewEncoder(os.Stdout)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 8*1024*1024), 8*1024*1024) // 8 MB — large enough for embedded JSONL

	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			// Malformed JSON-RPC — skip silently per spec
			continue
		}

		// Notifications have no id; don't respond.
		if req.ID == nil {
			continue
		}

		resp := dispatch(req)
		if err := enc.Encode(resp); err != nil {
			return fmt.Errorf("mcp: write: %w", err)
		}
	}
	return scanner.Err()
}

func dispatch(req rpcRequest) rpcResponse {
	switch req.Method {
	case "initialize":
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]any{"tools": map[string]any{}},
				"serverInfo":      map[string]any{"name": "soka", "version": "0.1.0"},
			},
		}

	case "ping":
		return rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}}

	case "tools/list":
		return rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"tools": toolList()}}

	case "tools/call":
		return dispatchTool(req)

	default:
		return rpcErr(req.ID, -32601, "method not found: "+req.Method)
	}
}

// toolList returns the MCP tool manifest for soka.
func toolList() []map[string]any {
	return []map[string]any{
		{
			"name":        "analyze",
			"description": "Ingest a JSONL agent event stream and return ranked MFPT-reducing policy knobs. Pass events inline via 'events' or by file path via 'events_path'. Equivalent to soka suggest.",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"target"},
				"properties": map[string]any{
					"events": map[string]any{
						"type":        "string",
						"description": "JSONL event stream — one JSON object per line, each with a 'type' or 'event' field. Use this for small streams; prefer events_path for large files.",
					},
					"events_path": map[string]any{
						"type":        "string",
						"description": "Path to a JSONL events file on disk. Preferred for large streams. Mutually exclusive with 'events'.",
					},
					"target": map[string]any{
						"type":        "string",
						"description": "Goal state to minimize MFPT toward (e.g. 'response.completed', 'CLAIM_VERIFIED').",
					},
					"keep": map[string]any{
						"type":        "string",
						"description": "Comma-separated boundary states to retain after Trace-fold. Empty = no fold (use all observed states).",
					},
					"knobs": map[string]any{
						"type":        "string",
						"description": "Comma-separated controllable from-states to consider for perturbation. Empty = all states except target.",
					},
					"halflife": map[string]any{
						"type":        "number",
						"description": "EMA half-life in events for recency weighting (0 = infinite memory, weight all events equally).",
						"default":     0,
					},
					"eps": map[string]any{
						"type":        "number",
						"description": "Perturbation mass epsilon — how much transition mass to shift when evaluating each knob.",
						"default":     0.10,
					},
					"top_k": map[string]any{
						"type":        "integer",
						"description": "Number of top knob suggestions to return.",
						"default":     5,
					},
				},
			},
		},
	}
}

func dispatchTool(req rpcRequest) rpcResponse {
	var params struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return rpcErr(req.ID, -32602, "invalid params: "+err.Error())
	}

	switch params.Name {
	case "analyze":
		return analyzeToolCall(req.ID, params.Arguments)
	default:
		return rpcErr(req.ID, -32602, "unknown tool: "+params.Name)
	}
}

func analyzeToolCall(id any, args map[string]any) rpcResponse {
	target, _ := args["target"].(string)
	if target == "" {
		return rpcErr(id, -32602, "'target' is required")
	}

	eventsStr, hasInline := args["events"].(string)
	eventsPath, hasPath := args["events_path"].(string)

	if !hasInline && !hasPath {
		return rpcErr(id, -32602, "one of 'events' or 'events_path' is required")
	}
	if hasInline && hasPath {
		return rpcErr(id, -32602, "'events' and 'events_path' are mutually exclusive")
	}

	var r interface{ Read([]byte) (int, error) }
	if hasPath {
		f, err := os.Open(eventsPath)
		if err != nil {
			return rpcErr(id, -32603, "open events_path: "+err.Error())
		}
		defer f.Close()
		r = f
	} else {
		r = strings.NewReader(eventsStr)
	}

	keep, _ := args["keep"].(string)
	knobs, _ := args["knobs"].(string)
	halflife, _ := args["halflife"].(float64)
	eps, ok := args["eps"].(float64)
	if !ok {
		eps = 0.10
	}
	topK := 5
	if v, ok := args["top_k"].(float64); ok {
		topK = int(v)
	}

	result, err := analyze(r, target, keep, knobs, halflife, eps, topK)
	if err != nil {
		return rpcErr(id, -32603, err.Error())
	}

	b, _ := json.MarshalIndent(result, "", "  ")
	return rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": string(b)},
			},
		},
	}
}

func rpcErr(id any, code int, msg string) rpcResponse {
	return rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: msg},
	}
}
