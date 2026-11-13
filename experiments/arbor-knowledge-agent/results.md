# Results: arbor-knowledge-agent

**Run date:** 2025  
**Fixture:** `RUC-NLPIR/Arbor` bundled demo session — `algotune_knn` CPU-only run (117 events)  
**Source:** `arbor/cli/assets/demo_session/events.jsonl` (shipped with `pip install arbor-agent`)

---

## State vocabulary (empirical)

catrace parses `type` field and appends `:phase` suffix where present, giving 21 distinct states.

| Role | States |
| ---- | ------ |
| **B (noise — fold)** | `llm.thinking_delta`, `llm.call`, `tool.start`, `tool.end` |
| **S (milestones — keep)** | `session.start`, `cycle.start`, `cycle.phase:observe`, `cycle.phase:ideate`, `cycle.phase:select`, `cycle.phase:dispatch`, `cycle.phase:backprop`, `cycle.phase:decide`, `cycle.end`, `executor.start`, `executor.end`, `idea.proposed`, `idea.completed`, `idea.merged`, `idea.pruned`, `convergence.reached`, `session.end` |
| **Target** | `session.end` |
| **Knobs** | `idea.proposed`, `executor.end`, `cycle.phase:dispatch`, `cycle.phase:decide` |

**Note on `idea.merged` as target:** null MFPT — paths can escape to `session.end` via `convergence.reached` without hitting `idea.merged` on that pass. `session.end` is the natural absorbing terminal.

---

## Results

### A — Full kernel (no Trace fold)

| Metric                 | Value                           |
| ---------------------- | ------------------------------- |
| `events_processed`     | 117                             |
| `lines_skipped`        | 0                               |
| `boundary_states`      | 21                              |
| `baseline_mfpt`        | 103.46                          |
| Top suggestion         | `llm.thinking_delta → session.end`, Δ = −67.38 |

Top suggestions dominated by noise states (llm.thinking_delta, tool.start/end) — expected without Trace fold.

### B — Trace-folded, ε sweep

Folding out 4 noise types reduces boundary states from 21 → 17 and baseline MFPT from 103.46 → 50.94. The noise states account for ~half the expected session steps.

| ε    | Top knob (`from → to`)                      | `mfpt_delta` | `projected_mfpt` |
| ---- | ------------------------------------------- | ------------ | ---------------- |
| 0.05 | `cycle.phase:dispatch → session.end`        | −11.32       | 39.63            |
| 0.15 | `cycle.phase:dispatch → session.end`        | −23.86       | 27.08            |
| 0.30 | `cycle.phase:dispatch → session.end`        | −33.01       | 17.93            |

**Monotonic `|Δ|` across ε sweep:** yes (−11 → −24 → −33) ✅

---

## Verdict

Claim **supported**. The top policy knob is `cycle.phase:dispatch → session.end` at all ε values — domain-interpretable as: once an executor has been dispatched, increasing the probability of early convergence (recognising the run is complete) yields the largest reduction in expected steps to session end. This maps directly to Arbor's `decide` phase decision logic.

**Caveats:**
- 117 events is sparse — MFPT estimates are directional, not authoritative
- Fixture is a demo/illustrative session, not a production run
- A real `algotune_knn` session (run with `arbor` + LLM API key) would give higher-confidence estimates

---

## Discoveries / surprises

- catrace appends `:phase` from arbor's `data.phase` field, giving `cycle.phase:observe` etc. — the `--keep` flag must use the expanded names, not bare `cycle.phase`
- Trace fold halves the MFPT (103 → 51), confirming that LLM token deltas and tool calls are the dominant source of path length in an Arbor session
