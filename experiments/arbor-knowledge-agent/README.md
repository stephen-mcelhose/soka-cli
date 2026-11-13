# Experiment: arbor-knowledge-agent

Runs `soka suggest` against [RUC-NLPIR/Arbor](https://github.com/RUC-NLPIR/Arbor) session traces to surface the dominant bottleneck in the Arbor hypothesis-tree cycle.

---

## Setup

### 1. Install arbor-agent

```sh
pip install arbor-agent
```

This ships a bundled demo session (117 events, `algotune_knn` CPU-only run) that `run.sh` copies automatically. No API key or login required for the demo fixture.

### 2. Install the Arbor skill suite

To run a **live experiment** (agent generates the trace rather than using the demo):

```sh
arbor install --target .agents/skills
```

This installs 11 skills into the repo's `.agents/skills/` directory. Restart your coding agent afterwards. The skills use the host agent's LLM — no separate API key needed.

### 3. Set up the benchmark target

```sh
# Clone just the kNN example
git clone --depth 1 --filter=blob:none --sparse https://github.com/RUC-NLPIR/Arbor.git /tmp/Arbor-src
cd /tmp/Arbor-src && git sparse-checkout set examples/algotune_knn
cp -r /tmp/Arbor-src/examples/algotune_knn /tmp/algotune_knn

# Install deps and verify baseline
cd /tmp/algotune_knn
pip install -r requirements.txt
python3 eval.py        # should print: score: ~1.01

# Git-init so arbor can create worktrees
git init -q && git add -A && git commit -qm "baseline"
```

---

## Running an experiment

### Option A — Demo fixture (no agent run needed)

Copies the bundled 117-event session and analyses it immediately:

```sh
bash experiments/arbor-knowledge-agent/run.sh
```

### Option B — Agent-generated trace (live run)

Ask the agent to run the Arbor skill suite against the kNN benchmark. The agent invokes:

- `arbor-research-agent` → `arbor-agent-orchestrator` → coordinator + executor skills  
- `arbor_state.py` to manage the idea tree and emit eval logs  
- Each eval cycle calls `python3 eval.py` in an isolated git worktree

A representative prompt:

```
Optimize the kNN solver in /tmp/algotune_knn/solution.py for throughput.
Eval: python3 eval.py from /tmp/algotune_knn — score: <float>, higher is better.
Baseline: ~1.01x. CPU-only NumPy only. Up to 4 cycles. Run name: live-run-1.
```

The agent calls `arbor_state.py init` at the start of the session, which creates the session
directory and writes the first event. Every subsequent `add`, `eval`, `update`, `record`, and
`emit` call appends to the same file automatically — no post-hoc reconstruction.

After the run, copy the live log to `testdata/`:

```sh
cp /tmp/algotune_knn/.arbor/sessions/live-run-1/events.jsonl \
   experiments/arbor-knowledge-agent/testdata/arbor-live.jsonl
```

Then run analysis against it:

```sh
bash experiments/arbor-knowledge-agent/run.sh \
  FIXTURE=experiments/arbor-knowledge-agent/testdata/arbor-live.jsonl
```

---

## Analysing the results

### Quick pass — full kernel

```sh
./soka suggest --target session.end --top-k 5 testdata/arbor-soka-run.jsonl
```

Expect noise states (`tool.start`, `llm.thinking_delta`) to dominate — they always do without folding.

### Trace-folded — milestone states only

```sh
KEEP="session.start,cycle.start,\
cycle.phase:observe,cycle.phase:ideate,cycle.phase:select,\
cycle.phase:dispatch,cycle.phase:backprop,cycle.phase:decide,\
cycle.end,executor.start,executor.end,\
idea.proposed,idea.completed,idea.merged,idea.pruned,\
convergence.reached,session.end"

KNOBS="idea.proposed,executor.end,cycle.phase:dispatch,cycle.phase:decide"

./soka suggest \
  --target session.end \
  --keep   "$KEEP" \
  --knobs  "$KNOBS" \
  --eps    0.15 \
  --top-k  3 \
  testdata/arbor-soka-run.jsonl
```

### Reading the output

The top knob (`from → to` with the largest negative `mfpt_delta`) tells you which transition, if made more probable, would most reduce the expected number of steps to session end.

| Top knob                              | Interpretation                                                                                           |
| ------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| `cycle.phase:dispatch → session.end`  | Agent is slow to detect convergence after dispatching — `decide` phase should be more aggressive at stopping |
| `idea.proposed → session.end`         | Over-ideation: too many proposals relative to successful merges — tighten the ideation budget or improve proposal quality |
| `executor.end → session.end`          | Merge threshold too conservative — winning ideas aren't being accepted fast enough                       |

Folding reduces baseline MFPT by roughly half — LLM token deltas and tool calls account for about 50% of expected session steps regardless of run quality.

---

## Next steps

**If the top knob is in the `decide` phase** (`dispatch → session.end`, `cycle.end → session.end`):  
The agent is running too many unnecessary cycles. Tune `coordinator.merge_threshold` down or prompt the `decide` phase to converge earlier when the improvement curve flattens.

**If the top knob is in the `ideate` phase** (`idea.proposed → session.end`):  
Reduce `max_cycles` or add a stricter ideation gate (`TreeView(format="constraints")` before proposing). Fewer, higher-quality proposals save more steps than better execution.

**If the top knob is in the `executor` phase** (`executor.end → idea.merged`):  
The merge evaluator is rejecting too aggressively. Lower `merge_threshold` in `research_config.yaml` or give the executor more turns (`executor_max_turns`).

**To increase confidence in estimates:**  
Run longer sessions (more cycles, larger benchmarks). The 93–117 event traces used here are directional. A production-scale session (500+ events) would give tighter MFPT estimates.
