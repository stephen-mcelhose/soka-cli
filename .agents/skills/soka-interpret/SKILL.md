---
name: soka-interpret
description: "Interpret soka suggest JSON output and return a plain-English verdict with health status, key finding, and recommended action. Trigger on: 'interpret this soka output', 'what does soka say', 'should I act on this', 'what does the MFPT mean', or any time soka suggest JSON is present and needs translation into guidance."
---

# soka-interpret

Translates the JSON output of `soka suggest` (or a `--stream` tick) into a structured plain-English verdict. Applies the heuristics from `docs/interpreting-results.md` to produce a health assessment, key finding, and concrete recommended action.

---

## Input

The agent passes soka JSON — either the full `soka suggest` output or a single stream tick. Additional context about the agent type or what the states mean is optional but improves action specificity.

---

## Output

Always produce a structured verdict with these four fields:

```
health:   green | yellow | red
finding:  <one sentence — what the data actually shows>
action:   <plain English — what to try next>
confidence: low | medium | high
```

Optionally add `caveat` if there is a nuance the caller should know.

---

## Interpretation rules

Apply these in order. The first triggered rule sets `health`.

### 1. Red flags (health: red — fix before optimising)

- `baseline_mfpt` is `null` → target is unreachable from some states. Check `classes.Recurrent` for traps that exclude the target. **Do not act on suggestions — fix structure first.**
- Any entry in `classes.Recurrent` whose states do not include the target → epistemic thrashing trap. The agent can loop forever and never reach the goal. This is a correctness failure, not a performance one.
- `events_processed < 20` → too few events for a reliable kernel. Results are illustrative only. **Do not act — pool more runs first.**

### 2. Yellow flags (health: yellow — act with caution)

- `events_processed` between 20 and 50 → low sample size. Results are directional but noisy. Pool more sessions before committing to a change.
- `lines_skipped` > 10% of `events_processed` → event format mismatch. Check field names; the kernel may be missing significant state transitions.
- `entropy_rate_bits` very high (> 2.0) → agent policy is nearly random. MFPT estimate is noisier; structural conclusions are less reliable.

### 3. Green (health: green — results are actionable)

All red and yellow conditions are absent.

---

## Reading the suggestions

Once health is determined, interpret the suggestions:

**Clustered pattern** — all top suggestions point to the same destination state, with similar `mfpt_delta` values:
- Loop count is the structural driver, not any single edge
- The convergence criterion (the logic that decides when to stop looping) is the highest-leverage place to experiment
- Action: *"Adjust the stopping condition in `<loop-control state>` to exit the cycle more aggressively."*

**Single-edge pattern** — one suggestion has a much larger delta than the rest:
- One specific routing decision accounts for most of the MFPT
- Directly tunable: reinforce that transition
- Action: *"Increase the probability of `<from>` → `<to>`. When the agent is in `<from>`, it should prefer `<to>` over its current default path."*

**Flat / no improvement** — `projected_mfpt` is still very high for all suggestions:
- No single-edge perturbation will move the needle
- The structure runs deeper; consider whether the target state is correctly defined or whether the agent needs architectural changes

---

## Confidence calibration

| `events_processed` | confidence |
| ------------------ | ---------- |
| < 20               | — (red, don't act) |
| 20 – 50            | low        |
| 50 – 200           | medium     |
| > 200              | high       |

Low entropy (`entropy_rate_bits ≈ 0`) with medium/high event count upgrades confidence by one level — a deterministic policy makes MFPT estimates more trustworthy despite smaller samples.

---

## Example verdicts

**Arbor-style clustered output:**

```
health:     green
finding:    MFPT is 27.9 steps (~2.3 cycles); all 96 suggestions point to session.end
            with uniform deltas (-7.5 to -8.0), confirming loop count is the structural driver.
action:     Tighten the convergence criterion in cycle.phase:decide — make the agent exit
            the loop more aggressively after execution. A 10% shift projects MFPT to 19.9 (−29%).
confidence: medium  (39 events, entropy ≈ 0 — deterministic policy, so estimate is reliable)
```

**Thrashing trap:**

```
health:     red
finding:    Recurrent class [[cycle.phase:observe, cycle.phase:ideate]] excludes the target.
            MFPT is infinite from trapped states — agent can never reach session.end.
action:     Fix the loop exit condition before optimising. The agent needs a path out of
            the observe–ideate cycle that leads toward session.end.
confidence: —
```

**Low data:**

```
health:     yellow
finding:    Only 14 events processed — kernel is too sparse for reliable conclusions.
action:     Pool at least 3–5 more sessions and re-run. Use
            `cat session-*.jsonl | soka suggest --target session.end`.
confidence: low
```

---

## Reference

Full heuristic reference: `docs/interpreting-results.md`
