# Results: harvey-deep-research

**Run date:** 2025  
**Fixture:** `harveyai/deep-research-starter` — `output_samples/events_jsonl.jsonl`  
**Reference:** `catrace/experiments/stream-deep-research-trace` (5/6 supported)

---

## Metrics

| Metric                          | catrace (recorded)              | soka-cli (this run)                   | Match? |
| ------------------------------- | ------------------------------- | ------------------------------------- | ------ |
| `events_processed`              | 8055                            | 8055                                  | ✅     |
| `lines_skipped`                 | 3228                            | 3228                                  | ✅     |
| `boundary_states` count (traced)| 14                              | 14                                    | ✅     |
| `entropy_rate_bits`             | ≈ 0 (near-deterministic sink)   | 0                                     | ✅     |
| Recurrent class                 | 1 (`response.completed`)        | 1 (`response.completed`)              | ✅     |
| Transient count                 | 13                              | 13                                    | ✅     |
| Top knob (`from → to`)          | `output_item.done → completed`  | `output_item.done → completed`        | ✅     |
| `mfpt_delta` @ ε=0.05           | −614                            | −614.25                               | ✅     |
| `mfpt_delta` @ ε=0.15           | −664                            | −663.98                               | ✅     |
| `mfpt_delta` @ ε=0.30           | −678                            | −677.70                               | ✅     |
| Monotonic `|Δ|` across ε sweep  | yes                             | yes (−614 → −664 → −678)              | ✅     |

**11/11 metrics aligned.**

---

## Verdict

Claim supported. `soka suggest` end-to-end reproduces the catrace library results exactly. The top policy knob (`response.output_item.done → response.completed`) and ε-monotonicity are consistent to within floating point rounding.

## Out of scope

`DriftScore` (Monitor fast/slow clock) is not surfaced by `soka suggest` — would require a future `soka drift` command. This accounts for the one catrace metric not covered here.
