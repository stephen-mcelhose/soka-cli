package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"syscall"

	"github.com/stephen-mcelhose/catrace/stream"
)

// streamFlags holds the extra flags that are only meaningful in --stream mode.
type streamFlags struct {
	minEvents      int
	deltaThreshold float64
}

// tick is the compact JSON object emitted each time the kernel snapshot
// crosses a session boundary and the emission criteria are met.
type tick struct {
	Tick            int      `json:"tick"`
	EventsProcessed int      `json:"events_processed"`
	BaselineMFPT    *float64 `json:"baseline_mfpt"`
	TopKnob         *topKnob `json:"top_knob,omitempty"`
}

type topKnob struct {
	From  string  `json:"from"`
	To    string  `json:"to"`
	Delta float64 `json:"delta"`
}

// runStream reads r line-by-line, calling est.ObserveNext for each event.
// On every session boundary (target state seen) it snapshots the kernel,
// runs AnalyzeAndSuggest, and emits a compact JSON tick to w if:
//   - events_processed >= minEvents, AND
//   - |mfpt - lastMFPT| > deltaThreshold (or deltaThreshold == 0)
//
// On SIGINT (or EOF) it flushes one final tick regardless of threshold.
func runStream(r io.Reader, w io.Writer, target, keep, knobs string, halfLife, eps float64, topK int, sf streamFlags) error {
	est := stream.NewEstimator(halfLife)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	type parsedEvent struct {
		state   string
		skipped bool
	}
	eventCh := make(chan parsedEvent, 256)

	go func() {
		defer close(eventCh)
		br := bufio.NewReaderSize(r, 1<<20)
		for {
			line, readErr := br.ReadBytes('\n')
			trimmed := bytes.TrimSpace(line)
			if len(trimmed) > 0 {
				var ev stream.Event
				if err := json.Unmarshal(trimmed, &ev); err != nil {
					eventCh <- parsedEvent{skipped: true}
				} else if state := ev.CanonicalState(); state != "" {
					eventCh <- parsedEvent{state: state}
				} else {
					eventCh <- parsedEvent{skipped: true}
				}
			}
			if readErr != nil {
				return
			}
		}
	}()

	enc := json.NewEncoder(w)

	var (
		tickNum          int
		observed         int
		lastMFPT         = math.NaN()
		lastTickObserved int
	)

	emit := func(final bool) error {
		if observed < sf.minEvents {
			return nil // data sufficiency floor — applies even at EOF
		}
		if final && observed == lastTickObserved {
			return nil // no new events since last tick — suppress redundant EOF flush
		}
		k, err := est.Kernel()
		if err != nil {
			return nil // not enough data yet
		}
		report, err := stream.AnalyzeAndSuggest(k, stream.SuggestConfig{
			TargetState:  target,
			Controllable: splitCSV(knobs),
			Epsilon:      eps,
			TopK:         1, // only need top-1 for the tick
		})
		if err != nil {
			return nil
		}

		mfpt := report.BaselineMFPT
		if !final && !math.IsNaN(lastMFPT) && sf.deltaThreshold > 0 {
			if math.IsInf(mfpt, 0) || math.Abs(mfpt-lastMFPT) <= sf.deltaThreshold {
				return nil // change not significant enough
			}
		}
		lastMFPT = mfpt

		tickNum++
		lastTickObserved = observed
		t := tick{
			Tick:            tickNum,
			EventsProcessed: observed,
		}
		if !math.IsInf(mfpt, 0) && !math.IsNaN(mfpt) {
			v := mfpt
			t.BaselineMFPT = &v
		}
		if len(report.Suggestions) > 0 {
			s := report.Suggestions[0]
			t.TopKnob = &topKnob{
				From:  s.FromState,
				To:    s.ToState,
				Delta: s.MFPTDelta,
			}
		}
		return enc.Encode(t)
	}

	for {
		select {
		case <-sigCh:
			return emit(true)
		case ev, ok := <-eventCh:
			if !ok {
				// EOF — final flush
				return emit(true)
			}
			if ev.skipped {
				continue
			}
			est.ObserveNext(ev.state)
			observed++
			if ev.state == target {
				if err := emit(false); err != nil {
					return fmt.Errorf("stream: encode tick: %w", err)
				}
				est.ResetSequence()
			}
		}
	}
}
