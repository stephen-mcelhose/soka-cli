# Interpreting soka Results

This guide walks through the full output of `soka suggest` using a real trace from the Arbor agent. By the end you should be able to read any soka output and know exactly what each field means, what is healthy, what is a warning sign, and what to do about it.

You do not need to know the mathematics to follow this guide. Where math is involved, the operational meaning is explained first.

---

## The example trace

All examples in this guide use the Arbor experiment fixture:

```sh
soka suggest \
  --target session.end \
  experiments/arbor-knowledge-agent/testdata/arbor-live.jsonl
```

Arbor is an autonomous optimisation agent. It runs a cycle of phases — observe, ideate, select, dispatch, execute, evaluate — and repeats that cycle until it converges on a solution, then ends the session. The events file records one real Arbor session.

The output is:

```json
{
  "events_processed": 39,
  "lines_skipped": 0,
  "baseline_mfpt": 27.910714285714285,
  "entropy_rate_bits": 1.6276588500693257e-11,
  "boundary_states": [ ... ],
  "classes": { ... },
  "suggestions": [ ... ]
}
```

---

## Field reference

### `events_processed` and `lines_skipped`

```json
"events_processed": 39,
"lines_skipped": 0
```

**What it means.** `events_processed` is how many events were successfully read from the JSONL stream. `lines_skipped` is how many lines were malformed (bad JSON, missing state field) and ignored.

**Why it matters.** The Markov kernel that soka builds is estimated from these events. The more events, the more reliable the estimate. A rough guide:

| `events_processed` | Reliability |
| ------------------ | ----------- |
| < 20               | Directional only — treat as a hint, not a measurement |
| 20–100             | Reasonable for understanding structure; some variance in MFPT numbers |
| 100–1000           | Good — transitions with low counts are still uncertain |
| 1000+              | High confidence across the full chain |

39 events is usable for understanding Arbor's loop structure, but you would want to pool multiple sessions (see [issue #1](https://github.com/stephen-mcelhose/soka-cli/issues/1)) before relying on the exact MFPT number for a production decision.

**Red flag.** A high `lines_skipped` relative to `events_processed` usually means the event format does not match what soka expects. Check that your events have a `type` or `event` field containing the state name.

---

### `boundary_states`

```json
"boundary_states": [
  "session.start",
  "cycle.start",
  "cycle.phase:observe",
  "cycle.phase:ideate",
  "idea.proposed",
  "cycle.phase:select",
  "cycle.phase:dispatch",
  "executor.start",
  "eval.end",
  "executor.end",
  "cycle.phase:backprop",
  "cycle.phase:decide",
  "cycle.end",
  "convergence.reached",
  "session.end"
]
```

**What it means.** These are the distinct state names observed in the event stream, in order of first appearance. They are the nodes of the Markov chain soka built from your trace.

**Why it matters.** This list tells you whether soka saw what you expected. If you expected `CLAIM_VERIFIED` to appear but it is missing, it means that state was never reached in your trace — either the agent never got there, or the event was named differently.

For Arbor we see 15 states: one entry point (`session.start`), one main cycle (12 states from `cycle.start` through `cycle.end`), one convergence state, and one terminal state (`session.end`).

**What to check.** Make sure your `--target` state appears in this list. If it does not, the MFPT calculation cannot complete and you will get an error.

---

### `baseline_mfpt`

```json
"baseline_mfpt": 27.910714285714285
```

**What it means.** Mean First Passage Time — the average number of state transitions the agent needs to reach your target (`session.end`) from wherever it currently is, weighted by the stationary distribution over starting states.

In plain language: **on average, Arbor takes ~28 steps to finish a session.**

A "step" here is a transition between observed states — not a wall-clock second or a token count. Each transition in the Arbor chain corresponds to a phase boundary: `cycle.phase:observe → cycle.phase:ideate` is one step, `executor.start → eval.end` is one step.

**How to read the number.** MFPT is relative to your agent and your target. 27.9 is not inherently high or low — it depends on how many states the agent has and how many cycles it typically runs. For a more immediately meaningful number, see `expected_iterations` below.

---

### `expected_iterations`

```json
"expected_iterations": 2.325892857142857
```

**What it means.** The average number of times the agent passes through its main loop before reaching the target. Computed as `baseline_mfpt / size(largest transient SCC)`.

For Arbor: the main loop has 12 states (`cycle.start` through `cycle.end`). `27.9 / 12 = 2.33` — Arbor runs approximately **2.3 full cycles** per session on average.

This is the operationally meaningful number. MFPT of 27.9 is abstract; "2.3 iterations" is concrete and actionable.

**When it is `null`.** Returned as `null` when:
- No main cycle is detectable (agent has no multi-state transient SCC — it is a straight-line pipeline with no loop)
- `baseline_mfpt` is itself `null`
- The agent has multiple loops of similar size (ambiguous which is "the" main cycle)

For agents without a loop, `expected_iterations` is not applicable — but `baseline_mfpt` and `suggestions` together still give a clear picture: MFPT is the pipeline depth, and the top suggestion identifies the specific step most worth shortcutting.

---

**Red flag.** If `baseline_mfpt` is `null`, the target state is unreachable from the stationary distribution — the agent can never get there from some starting configurations. This almost always means a recurrent trap exists that does not include the target (see `classes.Recurrent` below).

---

### `entropy_rate_bits`

```json
"entropy_rate_bits": 1.6276588500693257e-11
```

**What it means.** Entropy rate measures how much randomness there is in the agent's routing decisions. It is measured in bits per step.

- **Near zero** (like `1.6e-11`) means the agent is nearly deterministic — it follows the same path almost every time, with almost no routing variance.
- **High entropy** (e.g. `2.0` bits for a chain with 4 possible transitions from each state) means the agent is effectively random — it has not learned a stable policy.
- **Mid-range** means the agent has meaningful variation — different inputs or contexts lead to genuinely different paths.

**For Arbor.** The scale matters here. Entropy rate is bounded above by `log2(average out-degree)` — the maximum bits per step if the agent chose each outgoing transition with equal probability. Arbor's states mostly have one or two reachable successors, so the ceiling is roughly `log2(2) = 1 bit`. `1.6e-11` against a ceiling of `~1 bit` is eleven orders of magnitude below maximum — which is what makes it effectively zero, not intuition. Arbor is running the exact same cycle sequence on every trace in this dataset. This tells us:

1. The trace is from a single session with a fixed policy — there is no exploration.
2. The MFPT estimate is reliable for this specific policy (but may not generalise to other Arbor configurations).
3. The cycle count is structural, not random variance. The agent deterministically takes 2.3 cycles every time.

**What low entropy does NOT mean.** It does not mean the agent is broken. A fully converged, well-tuned agent should be nearly deterministic. Low entropy is healthy when MFPT is also low. It is worth investigating when MFPT is high — the agent is taking the same path every time, so if MFPT is high, it is deterministically so.

---

### `classes`

```json
"classes": {
  "SCCs": [
    ["session.start"],
    ["cycle.start", "cycle.phase:observe", "cycle.phase:ideate", "idea.proposed",
     "cycle.phase:select", "cycle.phase:dispatch", "executor.start", "eval.end",
     "executor.end", "cycle.phase:backprop", "cycle.phase:decide", "cycle.end"],
    ["convergence.reached"],
    ["session.end"]
  ],
  "Recurrent": [["session.end"]],
  "Transient": [
    "session.start", "cycle.start", "cycle.phase:observe", "cycle.phase:ideate",
    "idea.proposed", "cycle.phase:select", "cycle.phase:dispatch", "executor.start",
    "eval.end", "executor.end", "cycle.phase:backprop", "cycle.phase:decide",
    "cycle.end", "convergence.reached"
  ],
  "Periods": {"session.end": 1}
}
```

This section describes the **communicating class structure** of the Markov chain. It is the most mathematically dense part of the output — but the operational meaning is straightforward once you know what to look for.

#### `SCCs` — Strongly Connected Components

A Strongly Connected Component (SCC) is a group of states where the agent can get from any state in the group to any other state in the same group.

Arbor has four SCCs:

| SCC | Description |
| --- | ----------- |
| `["session.start"]` | Entry point — visited once, never returned to |
| `["cycle.start", ..., "cycle.end"]` | The main cycle — the agent loops here |
| `["convergence.reached"]` | Transition state — visited once after the loop exits |
| `["session.end"]` | Terminal state — the target |

The large 12-state SCC is the full observation-ideation-selection-dispatch-execution-evaluation loop. All 12 cycle states are mutually reachable — the agent can traverse the entire loop repeatedly. This is exactly what you would expect from a cyclic optimisation agent.

#### `Recurrent` — permanent traps

```json
"Recurrent": [["session.end"]]
```

A recurrent class is a group of states the agent enters and **never leaves**. Once the agent reaches a recurrent state, it stays there forever (or keeps cycling within the group forever, if the group has more than one state).

**For Arbor: only `session.end` is recurrent.** This is the ideal result. It means the only trap is the goal state — the agent will always eventually reach `session.end` and stay there. There are no pathological loops.

**The critical red flag: a recurrent class that does not contain your target.**

If `Recurrent` contained `[["cycle.phase:observe", "cycle.phase:ideate", "idea.proposed"]]` and your target was `session.end`, it would mean the agent can get permanently trapped cycling between those states and never reach the goal. This is **epistemic thrashing**: the agent loops indefinitely and can never converge.

When you see this, the MFPT to your target is mathematically infinite from the trapped states — the agent will never get there. This is a blocking diagnostic, not just a performance concern.

#### `Transient` — states the agent will eventually leave

```json
"Transient": ["session.start", "cycle.start", ..., "convergence.reached"]
```

Transient states are ones the agent will eventually leave with probability 1. For a well-structured agent, everything except the terminal state should be transient. Arbor's output is correct: 14 of 15 states are transient, only `session.end` is recurrent.

#### `Periods`

```json
"Periods": {"session.end": 1}
```

The period of an SCC is the greatest common divisor of all cycle lengths within that SCC. Period 1 means **aperiodic** — the agent can return to a state after any number of steps, not only multiples of some fixed number.

Period 1 is the common case for real agents. Period > 1 is possible — it means the agent follows a strict fixed-length cycle with no branching or self-loops — but rare in practice because real traces have branching and the EMA estimator's smoothing tends to wash it out.

**Period does not affect MFPT.** soka computes MFPT by solving a linear system, which works regardless of period. What period > 1 would break is stationary distribution estimation via power iteration — a periodic chain oscillates rather than converging. MFPT being well-defined depends on **reachability** (is the target accessible?), not aperiodicity. That is what `baseline_mfpt: null` signals.

The key is the first state name of the recurrent SCC. `"session.end": 1` is always 1 for a single absorbing state — a single-state SCC has only trivial cycles of length 1, so its period is 1 by definition.

---

### `total_suggestions`

```json
"total_suggestions": 96
```

The total number of edges that would improve MFPT if reinforced, before `--top-k` truncation. The `suggestions` array shows only the top-k (default 5).

This number matters for interpretation. When the top-k suggestions all look similar — identical deltas, same destination — you need to know whether that pattern holds across all 96 candidates or whether it breaks after entry 5. If `total_suggestions` is large and the top few all look the same, the structural bottleneck conclusion holds. If `total_suggestions` is 6 and 5 look similar, check the 6th before concluding.

To see more: `soka suggest --top-k 20 ...` or `--top-k 0` for all of them.

---

### `suggestions`

```json
"suggestions": [
  {
    "baseline_mfpt": 27.910714285714285,
    "epsilon": 0.1,
    "from": "executor.start",
    "to": "session.end",
    "mfpt_delta": -7.997448979591837,
    "projected_mfpt": 19.913265306122447
  },
  ...
]
```

The top-k ranked policy interventions — specific edges in the agent's transition graph that, if adjusted, would reduce MFPT the most.

Each suggestion has six fields:

| Field | Meaning |
| ----- | ------- |
| `from` | The state where the agent is making a routing decision |
| `to` | The destination state that should receive more probability mass |
| `epsilon` | How much mass was shifted in the perturbation (default 0.1 = 10%) |
| `baseline_mfpt` | MFPT before the perturbation (same as the top-level value) |
| `projected_mfpt` | Estimated MFPT if you shift `epsilon` mass to this transition |
| `mfpt_delta` | The difference — negative means improvement |

**How the calculation works.** For each candidate edge, soka takes the current transition matrix, shifts 10% of the probability mass that leaves `from` toward `to`, then re-solves for MFPT. The edge whose perturbation produces the largest negative delta is ranked first.

This is a local linear approximation — it is exact for the given epsilon, but does not predict what happens if you make larger changes or multiple simultaneous changes.

#### Reading the Arbor suggestions

```
#1  executor.start → session.end     delta: -8.0   projected: 19.9
#2  eval.end       → session.end     delta: -7.8   projected: 20.1
#3  executor.start → convergence.reached  delta: -7.7   projected: 20.2
#4  executor.end   → session.end     delta: -7.6   projected: 20.3
#5  idea.proposed  → session.end     delta: -7.6   projected: 20.3
```

**The pattern.** Every single suggestion is "shift mass from somewhere in the execution loop directly to `session.end`." And the deltas are almost identical (-7.5 to -8.0 across all five). This is not a coincidence — and `total_suggestions: 96` confirms it is not an artefact of the top-5 cutoff. All 96 improvable edges follow the same pattern.

When suggestions cluster like this — many edges, similar deltas, all pointing to the same destination — no single edge accounts for the difference. It is a structural property of the whole loop — the agent passes through the cycle multiple times before reaching the target.

The 10% perturbation on `executor.start → session.end` cuts MFPT from 27.9 to 19.9 — a 29% reduction. That is a large sensitivity from a small change, which confirms the loop is the primary contributor to MFPT.

**What "shift mass" means operationally.** If the top knob is `executor.start → session.end`, it means: when an executor finishes a task, it currently almost always loops back into the cycle. The agent would benefit from exiting more often after execution completes.

For Arbor, the actionable change is in the **convergence criterion** — the logic in `cycle.phase:decide` that decides whether to run another cycle or call it done. Right now it almost always runs another cycle. If it were more aggressive about converging, MFPT would drop significantly.

---

## Reading the whole picture

Good interpretation combines all four sections, not just the suggestions.

For Arbor:

1. **39 events, 0 skipped** — small but clean sample. Results are directional.
2. **MFPT 27.9** — ~2.3 cycles per session.
3. **Entropy ≈ 0** — deterministic policy. The 2.3 cycles is not noise; Arbor reliably does this every time. That makes the MFPT estimate more trustworthy despite the small sample.
4. **Recurrent: [session.end] only** — no thrashing traps. The agent will always converge. This is the green light on structure — if you want to improve MFPT, the agent will reach the goal, so the question is performance, not correctness.
5. **Suggestions: uniform cluster** — loop count is the structural driver, not any specific edge. The convergence criterion in `cycle.phase:decide` is the highest-leverage place to experiment.

The synthesis: **Arbor is structurally healthy. If lower MFPT is the goal, the convergence check in `cycle.phase:decide` is where to start.**

---

## Common warning signs

| Observation | What it likely means |
| ----------- | -------------------- |
| `baseline_mfpt: null` | Target is unreachable from some starting states — check for recurrent traps |
| Recurrent class that excludes your target | Epistemic thrashing trap — agent can loop forever without reaching goal |
| `events_processed < 20` | Too few events — results are illustrative only, do not act on them |
| High `lines_skipped` relative to processed | Event format mismatch — check field names |
| `entropy_rate_bits` very high | Agent policy is nearly random — has not learned stable routing |
| `entropy_rate_bits ≈ 0` with high MFPT | Deterministically high MFPT — structural, not random variance |
| All suggestions point to the same destination | Loop count is the structural driver — tuning the convergence criterion is higher leverage than any single edge |
| Suggestions have wildly different deltas | One specific edge accounts for most of the MFPT delta — directly tunable |
| `projected_mfpt` is still very high | No single-edge perturbation will move the needle — the structure runs deeper |

---

## Next steps after interpreting results

**If you have a clear top knob** (one suggestion with a much larger delta than the rest): that edge is the highest-leverage place to experiment. Reinforce that transition — increase the probability of taking that path — and re-run to see if MFPT drops as projected.

**If suggestions cluster uniformly** (like Arbor): loop count is the structural driver. The loop control logic — how many iterations the agent runs, what triggers convergence, what the stopping criterion is — is the highest-leverage place to experiment.

**If `baseline_mfpt` is null or recurrent traps exist**: fix correctness before optimising for speed. A trapped agent never reaches the goal regardless of how fast it moves.

**If `events_processed` is low**: pool more runs first (see [issue #1](https://github.com/stephen-mcelhose/soka-cli/issues/1)). Running `soka suggest` on 10 sessions concatenated will give a much more reliable kernel than running it on one.

**To verify an improvement**: run the agent with the adjusted policy, collect a new trace, run `soka suggest` again, and compare `baseline_mfpt` before and after. A real improvement shows as a lower MFPT in the new trace.
