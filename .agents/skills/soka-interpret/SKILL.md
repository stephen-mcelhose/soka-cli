---
name: soka-interpret
description: "Interpret soka suggest JSON output and return a plain-English verdict with health status, key finding, and recommended action. Trigger on: 'interpret this soka output', 'what does soka say', 'should I act on this', 'what does the MFPT mean', or any time soka suggest JSON is present and needs translation into guidance."
---

# soka-interpret

Translates the output of `soka suggest` (optionally pre-triaged by `soka check`) into a structured plain-English verdict. The output is always a *hypothesis to test*, never a confident verdict — soka works with small samples and the numbers are directional, not conclusive.

---

## Input

Either:
- Full `soka suggest` JSON, or
- A `soka check` result (`{health, flags}`) combined with the raw suggest JSON

Additional context about the agent type or what the states mean is optional but improves action specificity.

---

## Output

Always produce a structured verdict with these three fields:

```
health:   green | yellow | red   (pass through from soka check if available, otherwise derive)
finding:  <one sentence — what the data actually shows>
action:   <what to try next, always ending with "pool more sessions to validate">
```

Optionally add `caveat` if there is a nuance the caller should know.

**On red:** do not recommend changes — fix the structural issue first.

---

## Triage

The mechanical health rules (null MFPT, recurrent traps, event count, skip rate, entropy) are handled by `soka check`. If a `soka check` result is available, use it directly. If not, `soka check` can be run as:

```sh
soka suggest --target <target> events.jsonl | soka check
```

---

## Reading the suggestions

Once health is known, interpret the `suggestions` array from the suggest output:

**Clustered pattern** — all top suggestions point to the same destination state with similar `mfpt_delta` values:
- Loop count is the structural driver, not any single edge
- Action: *"Test tightening the stopping condition in `<loop-control state>` so the agent exits the cycle more aggressively."*

**Single-edge pattern** — one suggestion has a much larger delta than the rest:
- One routing decision accounts for most of the MFPT
- Action: *"Test increasing the probability of `<from>` → `<to>`. When the agent is in `<from>`, it should prefer `<to>` over its current default path."*

**Flat / no improvement** — `projected_mfpt` is still very high for all suggestions:
- No single-edge perturbation will move the needle; the structure may run deeper
- Action: *"Check whether the target state is correctly defined. Consider whether the agent needs architectural changes rather than routing adjustments."*

---

## Example verdicts

**Arbor-style clustered output (yellow — 39 events):**

```
health:   yellow
finding:  MFPT is 27.9 steps (~2.3 cycles); all 96 suggestions point to session.end
          with uniform deltas (-7.5 to -8.0) — loop count is the structural driver.
action:   Test tightening the convergence criterion in cycle.phase:decide so the agent
          exits the loop more aggressively. A 10% shift projects MFPT to 19.9 (−29%).
          Pool more sessions to validate — 39 events is a small sample.
caveat:   Entropy ≈ 0, so the estimate is more reliable than the sample size alone suggests.
```

**Thrashing trap (red):**

```
health:   red
finding:  Recurrent class [cycle.phase:observe, cycle.phase:ideate] excludes the target —
          the agent can loop forever and never reach session.end.
action:   Fix the loop exit condition before optimising. The agent needs a reachable
          path from the observe–ideate cycle to session.end.
```

**Too little data (red):**

```
health:   red
finding:  Only 14 events processed — kernel is too sparse for any structural conclusions.
action:   Pool more sessions and re-run:
          cat session-*.jsonl | soka suggest --target session.end
```

---

## Reference

Full heuristic reference: https://github.com/stephen-mcelhose/soka-cli/blob/main/docs/interpreting-results.md
