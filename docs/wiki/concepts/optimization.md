---
type: concept
title: Optimization
description: The domain of algorithmically improving a solution's performance metric — and the first application domain for soka, via Arbor running against the AlgoTune benchmark.
tags: [optimization, algotune, performance, benchmark, arbor]
timestamp: 2026-11-13T01:58:49Z
---

# Optimization

In the context of this project, **optimization** refers to algorithmic performance optimization: given a reference implementation of a computational task (e.g., k-NN search, linear programming, signal processing), find a faster correct implementation by any means available to a code-editing agent — algorithm substitution, numerical precision reduction, vectorization, caching.

This is the first application domain for soka because it provides a clean closed-loop signal: the evaluator runs `eval.py` and returns a numeric speedup ratio. This makes the feedback loop unambiguous, the terminal state (`convergence.reached` or `budget.exhausted`) well-defined, and the events tractable for [[Markov Chain]] analysis.

## AlgoTune as the benchmark

AlgoTune (Press et al., 2025) is a benchmark of 154 algorithmic optimization tasks drawn from SciPy, scikit-learn, and CVXPY. Each task provides:

- A `task.py` with a reference solver (the timing baseline) and an `is_solution` correctness verifier
- An editable `solution.py` for the agent to modify
- An `eval.py` that times both implementations and returns speedup = `reference_time / solution_time`

The metric is speedup ratio; higher is better. The constraint is correctness: `is_solution` must pass on every test instance.

AlgoTuner — the baseline agent in the AlgoTune paper — achieves an average 1.72× speedup. [[Arbor]] applied to `algotune_knn` (k-nearest-neighbour search, N=2000, D=16, Q=200, k=10) reached 9.78× — a task-specific regime where BLAS-based float32 computation dominates.

## Why optimization is a sharp beachhead for soka

1. **Deterministic metric** — speedup ratio is objective; no human labelling or LLM-as-judge required
2. **Clear terminal state** — convergence or budget exhaustion maps to a single event type
3. **Real cost** — each [[Hypothesis-Tree Refinement]] cycle costs API calls and wall time; MFPT reduction = dollar savings
4. **Transferable structure** — once soka has calibrated on optimization traces, the same Markov analysis applies to any [[Knowledge Agents|knowledge agent]] domain

## The optimization loop as a Markov chain

Under [[Hypothesis-Tree Refinement]], each Arbor cycle visits a fixed sequence of states. The key stochastic transitions are:

- `cycle.phase:ideate` → `cycle.phase:select` (always)
- `eval.end` → `convergence.reached` (if trunk improves) or `cycle.start` (if not)
- `cycle.phase:decide` → `session.end` (budget) or `cycle.start` (continue)

soka estimates the [[Mean First Passage Time]] from `session.start` to `session.end` and identifies which edge weight perturbation most reduces it — i.e., which phase transition, if made more likely, most shortens the total loop.

## Relationship to other concepts

- [[Arbor]] — the agent that runs optimization experiments using [[Hypothesis-Tree Refinement]]
- [[Knowledge Agents]] — optimization agents are a specific archetype within the broader knowledge agent topology
- [[Hypothesis-Tree Refinement]] — the search strategy used to explore the space of possible optimizations

## Sources

- Press et al. (2025). *AlgoTune: Can Language Models Speed Up General-Purpose Numerical Programs?* arXiv:2507.15887. https://arxiv.org/abs/2507.15887
- Jin et al. (2026). *Toward Generalist Autonomous Research via Hypothesis-Tree Refinement*. arXiv:2606.11926. https://arxiv.org/abs/2606.11926
