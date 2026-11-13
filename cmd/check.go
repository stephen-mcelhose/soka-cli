package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// checkResult is the output of soka check.
type checkResult struct {
	Health string   `json:"health"`
	Flags  []string `json:"flags"`
}

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [file]",
		Short: "Triage soka suggest output and emit a health verdict",
		Long: `Reads soka suggest JSON (from FILE or stdin) and applies deterministic
triage rules to produce a health verdict and list of flags.

Health levels:
  green   no issues detected — results are directional
  yellow  data quality concerns — treat results with caution
  red     structural or data problem — do not act, fix first

Exits 0 for green/yellow, 1 for red. Compose with suggest:

  soka suggest --target session.end events.jsonl | soka check`,
		Example: `  # pipe from suggest
  soka suggest --target session.end events.jsonl | soka check

  # from a saved suggest output
  soka check result.json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd, args)
		},
	}
	return cmd
}

func runCheck(cmd *cobra.Command, args []string) error {
	in, closer, err := openInput(args)
	if err != nil {
		return err
	}
	if closer != nil {
		defer closer.Close()
	}

	raw, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("parse: not valid soka suggest JSON: %w", err)
	}

	result := triage(doc)

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		return err
	}

	if result.Health == "red" {
		os.Exit(1)
	}
	return nil
}

// triage applies deterministic health rules to a soka suggest output map.
// Rules are evaluated in priority order; the first red rule wins.
// All yellow flags are collected and reported.
func triage(doc map[string]any) checkResult {
	var flags []string

	target, _ := doc["target"].(string)

	// --- red rules ---

	if _, hasKey := doc["baseline_mfpt"]; hasKey && doc["baseline_mfpt"] == nil {
		return checkResult{Health: "red", Flags: []string{"null_mfpt"}}
	}

	if classes, ok := doc["classes"].(map[string]any); ok {
		if recurrent, ok := classes["Recurrent"].([]any); ok {
			for _, rc := range recurrent {
				group, ok := rc.([]any)
				if !ok {
					continue
				}
				containsTarget := false
				for _, s := range group {
					if s == target {
						containsTarget = true
						break
					}
				}
				if !containsTarget {
					states := make([]string, 0, len(group))
					for _, s := range group {
						if str, ok := s.(string); ok {
							states = append(states, str)
						}
					}
					return checkResult{
						Health: "red",
						Flags:  []string{fmt.Sprintf("recurrent_trap:%v", states)},
					}
				}
			}
		}
	}

	events := intField(doc, "events_processed")
	if events < 20 {
		return checkResult{
			Health: "red",
			Flags:  []string{fmt.Sprintf("too_few_events:%d", events)},
		}
	}

	// --- yellow rules (all collected) ---

	if events < 50 {
		flags = append(flags, fmt.Sprintf("low_events:%d", events))
	}

	skipped := intField(doc, "lines_skipped")
	if events > 0 && float64(skipped)/float64(events) > 0.10 {
		pct := int(100 * float64(skipped) / float64(events))
		flags = append(flags, fmt.Sprintf("high_skip_rate:%d%%", pct))
	}

	if entropy, ok := doc["entropy_rate_bits"].(float64); ok && entropy > 2.0 {
		flags = append(flags, fmt.Sprintf("high_entropy:%.2f", entropy))
	}

	if len(flags) > 0 {
		return checkResult{Health: "yellow", Flags: flags}
	}

	return checkResult{Health: "green", Flags: []string{}}
}

func intField(doc map[string]any, key string) int {
	switch v := doc[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return 0
}
