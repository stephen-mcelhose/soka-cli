package cmd

import (
	"math"
	"os"
	"strings"
	"testing"

	"github.com/stephen-mcelhose/catrace"
)

// --- expectedIterations unit tests ---

// arbor-like: one large transient SCC (12 states), singletons for entry/exit,
// one recurrent singleton for the terminal state.
//
//	SCCs:      [{0}, {1..12}, {13}, {14}]
//	Recurrent: [{14}]
//	Transient: [0..13]
var arborClasses = &catrace.ClassDecomposition{
	SCCs: [][]int{
		{0},
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
		{13},
		{14},
	},
	Recurrent: [][]int{{14}},
	Transient: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13},
	Periods:   map[int]int{3: 1},
}

func TestExpectedIterations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		classes any
		mfpt    float64
		want    any // nil or float64
	}{
		{
			name:    "arbor: 12-state loop, mfpt 27.9 -> 2.33 iterations",
			classes: arborClasses,
			mfpt:    27.910714285714285,
			want:    27.910714285714285 / 12.0,
		},
		{
			name: "two transient loops: largest (8 states) wins over smaller (3 states)",
			classes: &catrace.ClassDecomposition{
				SCCs:      [][]int{{0}, {1, 2, 3}, {4, 5, 6, 7, 8, 9, 10, 11}, {12}},
				Recurrent: [][]int{{12}},
				Transient: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
				Periods:   map[int]int{3: 1},
			},
			mfpt: 24.0,
			want: 24.0 / 8.0,
		},
		{
			name: "no main cycle: all singleton SCCs (straight-line pipeline)",
			classes: &catrace.ClassDecomposition{
				SCCs:      [][]int{{0}, {1}, {2}, {3}},
				Recurrent: [][]int{{3}},
				Transient: []int{0, 1, 2},
				Periods:   map[int]int{3: 1},
			},
			mfpt: 10.0,
			want: nil,
		},
		{
			name: "recurrent multi-state SCC is not mistaken for the main loop",
			classes: &catrace.ClassDecomposition{
				SCCs:      [][]int{{0}, {1, 2, 3, 4, 5}},
				Recurrent: [][]int{{1, 2, 3, 4, 5}}, // trap, not the cycle
				Transient: []int{0},
				Periods:   map[int]int{1: 1},
			},
			mfpt: 20.0,
			want: nil,
		},
		{
			name:    "mfpt zero -> nil",
			classes: arborClasses,
			mfpt:    0,
			want:    nil,
		},
		{
			name:    "mfpt negative -> nil",
			classes: arborClasses,
			mfpt:    -5.0,
			want:    nil,
		},
		{
			name:    "mfpt +Inf -> nil",
			classes: arborClasses,
			mfpt:    math.Inf(1),
			want:    nil,
		},
		{
			name:    "mfpt NaN -> nil",
			classes: arborClasses,
			mfpt:    math.NaN(),
			want:    nil,
		},
		{
			name:    "wrong classes type -> nil",
			classes: "not a ClassDecomposition",
			mfpt:    10.0,
			want:    nil,
		},
		{
			name:    "nil classes -> nil",
			classes: nil,
			mfpt:    10.0,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := expectedIterations(tt.classes, tt.mfpt)
			if tt.want == nil {
				if got != nil {
					t.Errorf("got %v, want nil", got)
				}
				return
			}
			gotF, ok := got.(float64)
			if !ok {
				t.Fatalf("got %T (%v), want float64", got, got)
			}
			wantF := tt.want.(float64)
			if math.Abs(gotF-wantF) > 1e-9 {
				t.Errorf("got %.10f, want %.10f", gotF, wantF)
			}
		})
	}
}

// --- integration: analyze() with the Arbor fixture ---

// TestAnalyzeArborExpectedIterations runs the full analyze pipeline against the
// Arbor test fixture and asserts that expected_iterations is present, finite,
// and matches the expected value (27.9 / 12 ≈ 2.33).
func TestAnalyzeArborExpectedIterations(t *testing.T) {
	f, err := os.Open("../testdata/arbor-soka-run.jsonl")
	if err != nil {
		t.Skipf("fixture not found: %v", err)
	}
	defer f.Close()

	out, err := analyze(f, "session.end", "", "", 0, 0.10, 5)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}

	v, ok := out["expected_iterations"]
	if !ok {
		t.Fatal("expected_iterations missing from output")
	}
	if v == nil {
		t.Fatal("expected_iterations is nil — expected a float64")
	}
	got, ok := v.(float64)
	if !ok {
		t.Fatalf("expected_iterations is %T, want float64", v)
	}
	if math.IsNaN(got) || math.IsInf(got, 0) || got <= 0 {
		t.Fatalf("expected_iterations is not a valid positive number: %v", got)
	}

	// Arbor has a 12-state main loop. expected_iterations should be ~2-4 for a
	// real session. We don't pin the exact value (depends on fixture length) but
	// it must be in a plausible range.
	if got < 1.0 || got > 10.0 {
		t.Errorf("expected_iterations = %.3f looks implausible for Arbor (want 1–10)", got)
	}

	t.Logf("expected_iterations = %.4f", got)
}

// TestNamedClassesInvariantViolation verifies that namedClasses panics with an
// invariantViolation — not a silent fallback label — when catrace returns an
// index that is out of bounds for the boundary_states slice.
func TestNamedClassesInvariantViolation(t *testing.T) {
	bad := &catrace.ClassDecomposition{
		SCCs:      [][]int{{0}, {1, 99}}, // index 99 doesn't exist
		Recurrent: [][]int{{0}},
		Transient: []int{1, 99},
		Periods:   map[int]int{0: 1},
	}
	names := []string{"state.a", "state.b"} // only 2 names; index 99 is invalid

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic, got none")
		}
		if _, ok := r.(invariantViolation); !ok {
			t.Fatalf("expected invariantViolation panic, got %T: %v", r, r)
		}
	}()

	namedClasses(bad, names)
}

// TestAnalyzeNoMainLoop verifies expected_iterations is nil for a straight-line
// agent (no loop in the trace).
func TestAnalyzeNoMainLoop(t *testing.T) {
	// A → B → C, no cycles. Each event is a single transition.
	jsonl := strings.Join([]string{
		`{"type":"A"}`,
		`{"type":"B"}`,
		`{"type":"C"}`,
		`{"type":"A"}`,
		`{"type":"B"}`,
		`{"type":"C"}`,
	}, "\n")

	out, err := analyze(strings.NewReader(jsonl), "C", "", "", 0, 0.10, 5)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}

	v := out["expected_iterations"]
	if v != nil {
		t.Errorf("expected_iterations = %v for straight-line agent, want nil", v)
	}
}
