package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/stephen-mcelhose/catrace/stream"
)

type suggestFlags struct {
	target   string
	keep     string
	knobs    string
	halfLife float64
	eps      float64
	topK     int
}

func newSuggestCmd() *cobra.Command {
	f := &suggestFlags{}
	cmd := &cobra.Command{
		Use:   "suggest [file]",
		Short: "Estimate a Markov kernel from a JSONL event stream and suggest MFPT knobs",
		Long: `Stream JSON events (JSONL) into an EMA Estimator, optionally Trace-fold
onto boundary states, then rank single-row policy perturbations that reduce
mean first-passage time to --target via AnalyzeAndSuggest.

Reads FILE if given, otherwise stdin. Malformed lines are skipped.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSuggest(cmd, f, args)
		},
	}
	cmd.Flags().StringVar(&f.target, "target", "", "goal state to minimize MFPT toward (required)")
	cmd.Flags().StringVar(&f.keep, "keep", "", "comma-separated boundary states to retain after Trace (empty = no fold)")
	cmd.Flags().StringVar(&f.knobs, "knobs", "", "comma-separated controllable from-states (empty = all except target)")
	cmd.Flags().Float64Var(&f.halfLife, "halflife", 0, "EMA half-life in events (0 = infinite memory)")
	cmd.Flags().Float64Var(&f.eps, "eps", 0.10, "perturbation mass epsilon")
	cmd.Flags().IntVar(&f.topK, "top-k", 5, "number of top interventions to return")
	cmd.Flags().IntVar(&f.topK, "top", 5, "alias for --top-k")
	_ = cmd.MarkFlagRequired("target")
	return cmd
}

func runSuggest(cmd *cobra.Command, f *suggestFlags, args []string) error {
	in, closer, err := openInput(args)
	if err != nil {
		return err
	}
	if closer != nil {
		defer closer.Close()
	}

	out, err := analyze(in, f.target, f.keep, f.knobs, f.halfLife, f.eps, f.topK)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

// analyze is the shared core used by both the CLI and the MCP server.
func analyze(r io.Reader, target, keep, knobs string, halfLife, eps float64, topK int) (map[string]any, error) {
	est := stream.NewEstimator(halfLife)
	observed, skipped, err := stream.IngestJSONL(r, est.ObserveNext)
	if err != nil {
		return nil, fmt.Errorf("ingest: %w", err)
	}
	if observed == 0 {
		return nil, fmt.Errorf("no events with a canonical state observed (%d lines skipped)", skipped)
	}

	k, err := est.Kernel()
	if err != nil {
		return nil, err
	}

	keepList := splitCSV(keep)
	if len(keepList) > 0 {
		k, err = stream.TraceByName(k, keepList, 1e-12)
		if err != nil {
			return nil, fmt.Errorf("trace: %w", err)
		}
		if err := k.Validate(1e-9); err != nil {
			return nil, fmt.Errorf("traced kernel invalid: %w", err)
		}
	}

	report, err := stream.AnalyzeAndSuggest(k, stream.SuggestConfig{
		TargetState:  target,
		Controllable: splitCSV(knobs),
		Epsilon:      eps,
		TopK:         topK,
	})
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"events_processed":  observed,
		"lines_skipped":     skipped,
		"boundary_states":   report.BoundaryStates,
		"entropy_rate_bits": report.EntropyRate,
		"classes":           report.Classes,
		"baseline_mfpt":     finiteOrNil(report.BaselineMFPT),
		"suggestions":       sanitizeSuggestions(report.Suggestions),
	}, nil
}

func finiteOrNil(v float64) any {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return nil
	}
	return v
}

func sanitizeSuggestions(in []stream.KnobSuggestion) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	for _, s := range in {
		out = append(out, map[string]any{
			"from":           s.FromState,
			"to":             s.ToState,
			"epsilon":        s.Epsilon,
			"baseline_mfpt":  finiteOrNil(s.BaselineMFPT),
			"projected_mfpt": finiteOrNil(s.ProjectedMFPT),
			"mfpt_delta":     finiteOrNil(s.MFPTDelta),
		})
	}
	return out
}

func openInput(args []string) (io.Reader, io.Closer, error) {
	if len(args) == 0 {
		return os.Stdin, nil, nil
	}
	f, err := os.Open(args[0])
	if err != nil {
		return nil, nil, err
	}
	return f, f, nil
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
