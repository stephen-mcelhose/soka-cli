package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// buildLoopJSONL returns a JSONL string of N full A→B→C→target cycles.
func buildLoopJSONL(n int, target string) string {
	var lines []string
	for i := 0; i < n; i++ {
		lines = append(lines, `{"type":"A"}`, `{"type":"B"}`, `{"type":"C"}`)
		lines = append(lines, `{"type":"`+target+`"}`)
	}
	return strings.Join(lines, "\n")
}

// TestStreamEmitsTickAtSessionBoundary verifies that runStream emits at least
// one tick when the target state is seen and min-events is satisfied.
func TestStreamEmitsTickAtSessionBoundary(t *testing.T) {
	// 15 sessions of A→B→C→done; 60 events total — well above default min-events 50.
	jsonl := buildLoopJSONL(15, "done")
	var buf bytes.Buffer

	err := runStream(
		strings.NewReader(jsonl),
		&buf,
		"done", "", "", 0, 0.10, 1,
		streamFlags{minEvents: 50, deltaThreshold: 0},
	)
	if err != nil {
		t.Fatalf("runStream: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) == 0 || lines[0] == "" {
		t.Fatal("expected at least one tick, got none")
	}

	// Decode the last tick and verify its shape.
	var last tick
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("failed to decode last tick: %v", err)
	}
	if last.Tick <= 0 {
		t.Errorf("tick number should be positive, got %d", last.Tick)
	}
	if last.EventsProcessed <= 0 {
		t.Errorf("events_processed should be positive, got %d", last.EventsProcessed)
	}
}

// TestStreamRespectsMinEvents verifies that no tick is emitted when
// events_processed < min-events at every session boundary.
func TestStreamRespectsMinEvents(t *testing.T) {
	// 2 sessions — only 8 events total, below min-events 50.
	jsonl := buildLoopJSONL(2, "done")
	var buf bytes.Buffer

	err := runStream(
		strings.NewReader(jsonl),
		&buf,
		"done", "", "", 0, 0.10, 1,
		streamFlags{minEvents: 50, deltaThreshold: 0},
	)
	if err != nil {
		t.Fatalf("runStream: %v", err)
	}

	// Final flush on EOF: still below min-events, so nothing should be emitted.
	if strings.TrimSpace(buf.String()) != "" {
		t.Errorf("expected no ticks below min-events, got: %s", buf.String())
	}
}

// TestStreamDeltaThresholdSuppressesTick verifies that a tick is suppressed
// when MFPT hasn't moved by more than delta-threshold since the last tick.
func TestStreamDeltaThresholdSuppressesTick(t *testing.T) {
	// Many identical sessions — MFPT should stabilise quickly, so after the
	// first tick the delta falls below a very high threshold (1000).
	jsonl := buildLoopJSONL(30, "done")
	var buf bytes.Buffer

	err := runStream(
		strings.NewReader(jsonl),
		&buf,
		"done", "", "", 0, 0.10, 1,
		streamFlags{minEvents: 10, deltaThreshold: 1000},
	)
	if err != nil {
		t.Fatalf("runStream: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	// The first tick fires when MFPT is first observed; subsequent ones are
	// suppressed. We should see very few ticks (at most a handful before MFPT
	// stabilises). Certainly not 30.
	if len(lines) >= 30 {
		t.Errorf("delta-threshold did not suppress ticks: got %d lines", len(lines))
	}
}

// TestStreamExistingBehaviourUnchanged verifies that running suggest without
// --stream still produces a single JSON object (not a tick stream).
func TestStreamExistingBehaviourUnchanged(t *testing.T) {
	jsonl := buildLoopJSONL(15, "done")
	out, err := analyze(strings.NewReader(jsonl), "done", "", "", 0, 0.10, 5)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	// Must still have all the original keys.
	for _, key := range []string{"events_processed", "baseline_mfpt", "suggestions", "total_suggestions"} {
		if _, ok := out[key]; !ok {
			t.Errorf("analyze() output missing key %q after stream changes", key)
		}
	}
}
