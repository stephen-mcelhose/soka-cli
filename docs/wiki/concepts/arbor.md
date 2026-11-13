---
type: concept
title: Arbor
description: "Optimize anything" — an autonomous agent from RUC/Microsoft Research that treats scientific research methodology (hypothesis-tree refinement) as an optimization strategy, targeting any benchmarkable metric. soka's primary beachhead for event-trace analysis.
tags: [arbor, autonomous-agents, optimization, events-jsonl, beachhead, hypothesis-tree]
timestamp: 2026-11-13T01:58:49Z
---

# Arbor

Arbor's tagline is **"Optimize anything."** It is an autonomous agent that positions scientific research methodology as the most robust optimization strategy for complex, benchmarkable tasks. Where a standard optimizer mutates parameters and measures outcomes, Arbor adds the full apparatus of scientific inquiry — literature checks, structured hypothesis trees, backpropagation of insights — to avoid the local minima and redundant paths that simpler optimizers hit. Research is the cognitive method; optimization is the concrete objective.

Developed jointly by Renmin University of China (Gaoling School of AI) and Microsoft Research, Arbor can target any task with a numeric metric: code speedup (AlgoTune), ML engineering (MLE-Bench), math reasoning, web research accuracy. The scorer is the only task-specific component; the rest of the loop is general.

Its significance for soka is two-fold. First, Arbor's cycle maps cleanly onto the [[Knowledge Agents|knowledge agent]] behavioral topology. Second, it emits a well-structured `events.jsonl` log — a timestamped stream of milestone events — making it the sharpest available source of real agent execution traces for [[catrace]] analysis.

## Architecture

Arbor uses a two-agent design:

- **Coordinator (Research Director):** Maintains the Idea Tree, drives planning cycles, and dispatches tasks to executors.
- **Executor (Research Engineer):** Writes code changes, runs experiments inside isolated git worktrees, evaluates outcomes, and reports feedback upward.

The coordinator reads prior results at every `observe` phase, pruning dead branches and expanding promising ones — this is [[Hypothesis-Tree Refinement]] in action.

## The Arbor Cycle

Each research cycle follows a fixed six-phase sequence:

```
cycle.start
  → cycle.phase:observe      (read tree state, prior results, constraints)
  → cycle.phase:ideate       (propose child hypotheses)
  → cycle.phase:select       (choose which to pursue this cycle)
  → cycle.phase:dispatch     (spawn executor agents in isolated worktrees)
  → executor.start / eval.end / executor.end  (per hypothesis)
  → cycle.phase:backprop     (propagate scores and insights up the tree)
  → cycle.phase:decide       (merge to trunk, prune, or continue)
cycle.end
```

This repeating structure produces a Markov chain with a small, interpretable state space — the ideal input for `soka suggest`.

## events.jsonl as the beachhead

Arbor is soka's primary beachhead because:

1. **Clean schema** — events are typed, timestamped, and milestone-level. No free-text parsing needed.
2. **Open-source** — the event format is not proprietary; `arbor_state.py` emits it natively.
3. **Expensive loops** — each cycle costs real API calls and wall-clock time. MFPT analysis directly translates to dollar savings.
4. **Interpretable states** — `cycle.phase:ideate`, `eval.end`, `convergence.reached` are semantically meaningful; soka's knob recommendations are actionable.

## AlgoTune as the benchmark domain

In the soka experiments, Arbor runs against AlgoTune tasks — a benchmark of 154 algorithmic optimization problems sourced from SciPy, scikit-learn, and CVXPY (Press et al., 2025). The scorer is `eval.py`; the metric is speedup over a reference implementation. Arbor's cycle is well-matched to AlgoTune because each hypothesis corresponds to a concrete, verifiable code change.

## Relationship to other concepts

- [[Knowledge Agents]] — Arbor is a concrete instance of the knowledge agent topology
- [[Hypothesis-Tree Refinement]] — the search strategy Arbor uses to explore the solution space
- [[Optimization]] — the domain Arbor operates in; the scorer defines the objective function
- [[Deep Research]] — a parallel archetype: web search instead of code evaluation, same loop shape

## Sources

- RUC-NLPIR. *Arbor*. GitHub repository. https://github.com/RUC-NLPIR/Arbor
- Jin et al. (2026). *Toward Generalist Autonomous Research via Hypothesis-Tree Refinement*. arXiv:2606.11926. https://arxiv.org/abs/2606.11926
- Press et al. (2025). *AlgoTune: Can Language Models Speed Up General-Purpose Numerical Programs?* arXiv:2507.15887. https://arxiv.org/abs/2507.15887
