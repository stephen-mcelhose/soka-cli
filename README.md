# soka-cli

**Self Optimizing Knowledge Agents** — a Go CLI for Markov-based policy analysis of agent event streams.

Backed by [catrace](https://github.com/stephen-mcelhose/catrace) v0.1.1.

---

## Install

```sh
go install github.com/stephen-mcelhose/soka-cli@latest
```

Or build from source:

```sh
git clone https://github.com/stephen-mcelhose/soka-cli.git
cd soka-cli
go build -o soka .
```

---

## Commands

### `soka suggest`

Ingest a JSONL agent event stream, estimate a Markov kernel, and rank single-row policy perturbations that reduce mean first-passage time (MFPT) to a target goal state.

```sh
soka suggest --target answer:done events.jsonl
cat events.jsonl | soka suggest --target CLAIM_VERIFIED --top-k 3
```

**Flags**

| Flag          | Default | Description                                               |
| ------------- | ------- | --------------------------------------------------------- |
| `--target`    | —       | Goal state to minimize MFPT toward **(required)**         |
| `--keep`      | —       | Comma-separated boundary states to retain after Trace     |
| `--knobs`     | —       | Comma-separated controllable from-states (default: all)   |
| `--halflife`  | 0       | EMA half-life in events (0 = infinite memory)             |
| `--eps`       | 0.10    | Perturbation mass epsilon                                 |
| `--top-k`     | 5       | Number of top interventions to return                     |

**Output** (JSON)

```json
{
  "events_processed": 412,
  "lines_skipped": 3,
  "boundary_states": ["QUERY_DECOMPOSED", "CLAIM_VERIFIED", "ABSTAINED_DEAD_END"],
  "entropy_rate_bits": 1.42,
  "classes": {
    "SCCs": [["QUERY_DECOMPOSED", "CLAIM_VERIFIED"], ["ABSTAINED_DEAD_END"]],
    "Recurrent": [["ABSTAINED_DEAD_END"]],
    "Transient": ["QUERY_DECOMPOSED", "CLAIM_VERIFIED"],
    "Periods": {"ABSTAINED_DEAD_END": 1}
  },
  "baseline_mfpt": 8.3,
  "expected_iterations": 4.15,
  "total_suggestions": 12,
  "suggestions": [
    {
      "from": "CONTRADICTION_FOUND",
      "to": "WEB_FALLBACK",
      "epsilon": 0.1,
      "baseline_mfpt": 8.3,
      "projected_mfpt": 3.4,
      "mfpt_delta": -4.9
    }
  ]
}
```

For a full field-by-field walkthrough see [docs/interpreting-results.md](docs/interpreting-results.md).

### `soka mcp`

Start a stdio-based MCP server exposing Markov analysis as callable tools — compatible with LangGraph, LlamaIndex, Claude Agent SDK, and any MCP client.

> **Coming soon.**

---

## Event format

Events are JSONL, one JSON object per line. The state label is resolved from each line in priority order:

1. `type` string field
2. `event` string field (fallback)
3. If `status` is set, appended as `:<status>`
4. If `data.phase` is set, appended as `:<phase>` (Arbor-style cycle events)

```jsonl
{"type": "QUERY_DECOMPOSED", "ts": "2025-01-01T00:00:00Z"}
{"type": "response.web_search_call", "status": "in_progress"}
{"type": "cycle.phase", "data": {"phase": "observe"}}
{"event": "tool.end", "status": "success"}
```

Malformed lines and lines with no resolvable state are skipped and counted in `lines_skipped`. All other fields are ignored.

## Bring your own trace

Any JSONL stream where each line carries a `type` or `event` string works directly. For frameworks that use a different field name, a one-line `jq` remap is sufficient — no adapter library needed:

```sh
# LangGraph streaming (uses "event" natively — may work as-is)
cat langgraph_trace.jsonl | soka suggest --target graph.end

# LangSmith run export (uses "run_type")
cat langsmith_runs.jsonl | jq -c '{type: .run_type, status: .status}' \
  | soka suggest --target chain:success

# OpenTelemetry spans (uses "name")
cat otel_spans.jsonl | jq -c '{type: .name, status: .status}' \
  | soka suggest --target request.end

# PI agent (--mode json emits JSON events over stdout)
pi --mode json -p "research topic X" \
  | soka suggest --target response.done

# Any other source: map whatever field holds the state label to "type"
cat custom.jsonl | jq -c '{type: .YOUR_FIELD}' | soka suggest --target ...
```

The experiments in `experiments/` each demonstrate a validated trace source with boundary states and flags recorded inline.

---

## License

[MIT](LICENSE) © 2025 Stephen McElhose
