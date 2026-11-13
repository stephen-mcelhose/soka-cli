package cmd

import (
	"testing"
)

func doc(overrides map[string]any) map[string]any {
	base := map[string]any{
		"target":            "session.end",
		"events_processed":  float64(100),
		"lines_skipped":     float64(0),
		"entropy_rate_bits": float64(0.01),
		"baseline_mfpt":     float64(27.9),
		"classes": map[string]any{
			"Recurrent": []any{
				[]any{"session.end"},
			},
		},
	}
	for k, v := range overrides {
		base[k] = v
	}
	return base
}

func TestTriageGreen(t *testing.T) {
	r := triage(doc(nil))
	if r.Health != "green" {
		t.Errorf("want green, got %s (flags: %v)", r.Health, r.Flags)
	}
	if len(r.Flags) != 0 {
		t.Errorf("want no flags, got %v", r.Flags)
	}
}

func TestTriageRedNullMFPT(t *testing.T) {
	r := triage(doc(map[string]any{"baseline_mfpt": nil}))
	if r.Health != "red" {
		t.Errorf("want red, got %s", r.Health)
	}
	if len(r.Flags) == 0 || r.Flags[0] != "null_mfpt" {
		t.Errorf("want null_mfpt flag, got %v", r.Flags)
	}
}

func TestTriageRedRecurrentTrap(t *testing.T) {
	r := triage(doc(map[string]any{
		"classes": map[string]any{
			"Recurrent": []any{
				[]any{"cycle.phase:observe", "cycle.phase:ideate"}, // excludes target
			},
		},
	}))
	if r.Health != "red" {
		t.Errorf("want red, got %s", r.Health)
	}
	if len(r.Flags) == 0 {
		t.Error("want recurrent_trap flag, got none")
	}
}

func TestTriageRedTooFewEvents(t *testing.T) {
	r := triage(doc(map[string]any{"events_processed": float64(14)}))
	if r.Health != "red" {
		t.Errorf("want red, got %s", r.Health)
	}
	if len(r.Flags) == 0 || r.Flags[0] != "too_few_events:14" {
		t.Errorf("want too_few_events:14, got %v", r.Flags)
	}
}

func TestTriageYellowLowEvents(t *testing.T) {
	r := triage(doc(map[string]any{"events_processed": float64(35)}))
	if r.Health != "yellow" {
		t.Errorf("want yellow, got %s", r.Health)
	}
	if len(r.Flags) == 0 || r.Flags[0] != "low_events:35" {
		t.Errorf("want low_events:35, got %v", r.Flags)
	}
}

func TestTriageYellowHighSkipRate(t *testing.T) {
	r := triage(doc(map[string]any{
		"events_processed": float64(100),
		"lines_skipped":    float64(15), // 15% > 10%
	}))
	if r.Health != "yellow" {
		t.Errorf("want yellow, got %s", r.Health)
	}
	found := false
	for _, f := range r.Flags {
		if f == "high_skip_rate:15%" {
			found = true
		}
	}
	if !found {
		t.Errorf("want high_skip_rate:15%% flag, got %v", r.Flags)
	}
}

func TestTriageYellowHighEntropy(t *testing.T) {
	r := triage(doc(map[string]any{"entropy_rate_bits": float64(2.5)}))
	if r.Health != "yellow" {
		t.Errorf("want yellow, got %s", r.Health)
	}
	found := false
	for _, f := range r.Flags {
		if len(f) >= 12 && f[:12] == "high_entropy" {
			found = true
		}
	}
	if !found {
		t.Errorf("want high_entropy flag, got %v", r.Flags)
	}
}

func TestTriageYellowMultipleFlags(t *testing.T) {
	// Low events + high entropy — both yellow flags should appear.
	r := triage(doc(map[string]any{
		"events_processed":  float64(30),
		"entropy_rate_bits": float64(3.0),
	}))
	if r.Health != "yellow" {
		t.Errorf("want yellow, got %s", r.Health)
	}
	if len(r.Flags) < 2 {
		t.Errorf("want at least 2 yellow flags, got %v", r.Flags)
	}
}

func TestTriageRedBeatsYellow(t *testing.T) {
	// Low events (yellow) + null MFPT (red) — red wins.
	r := triage(doc(map[string]any{
		"events_processed": float64(14),
		"baseline_mfpt":    nil,
	}))
	if r.Health != "red" {
		t.Errorf("want red, got %s", r.Health)
	}
}
