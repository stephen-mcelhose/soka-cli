---
type: concept
title: catrace
description: A Go library for modeling autonomous agents and agent networks as finite-state Markov systems — the mathematical engine underlying soka-cli, providing trace-chain folding, MFPT computation, and communicating-class analysis.
tags: [catrace, markov-chain, mfpt, trace, engine, go]
timestamp: 2026-11-13T01:58:49Z
resource: https://github.com/stephen-mcelhose/catrace
---

# catrace

catrace is a Go library that treats the classic agent loop — **perceive → decide → act** — as the object of mathematical study. It models agent behaviour as a finite-state Markov system and provides exact analytical tools for understanding that system: stationary distributions, entropy rates, communicating classes, mean first passage times, and commute times. soka-cli is built on top of catrace; `soka suggest` is catrace's trace-chain machinery exposed as a CLI.

The library extracts the stochastic and Markov machinery from Hoffman, Prakash & Chattopadhyay's 2024 paper on agent loop formalisation, explicitly excluding any philosophical claims and retaining only the operational mathematics.

## The core abstraction: trace-chain folding

The key insight is the **trace equation**. A real agent loop emits hundreds of low-level events (token deltas, embedding calls, chunk re-rankings, SQL retries) for every meaningful epistemic milestone. These hidden background states ($B$) swamp the observable boundary states ($S$) that actually matter for routing decisions.

The trace equation collapses $B$ out exactly:

$$\text{Tr}(L) = L_{SS} + L_{SB}(I - L_{BB})^{-1}L_{BS}$$

where $L$ is the full transition matrix over all states $S \cup B$. The result is an exact transition kernel over just the observable states $S$ — capturing how the hidden dynamics shape the transitions between milestones, without approximation.

This is what `soka suggest` does: it folds a raw `events.jsonl` stream (which contains both boundary milestones and noisy micro-steps) into a compact Markov chain over boundary states only, then runs MFPT analysis on that chain.

## Three cyclic kernels (the agent loop geometry)

catrace models an agent as three interacting maps across three spaces:

| Map | Symbol | Spaces |
| --- | ------ | ------ |
| Perception | $P$ | World $W$ → Experience $X$ |
| Decision | $D$ | Experience $X$ → Strategy $G$ |
| Action | $A$ | Strategy $G$ → World $W$ |

Composing these three maps in different orders produces three cyclic kernels — three equivalent views of the same agent loop:

| Kernel | Formula | Perspective |
| ------ | ------- | ----------- |
| Qualia kernel $Q$ | $DAP$ | Loop viewed through Experience |
| Strategy kernel $S$ | $APD$ | Loop viewed through Action |
| World kernel $W$   | $PDA$ | Loop viewed through World |

For soka's purposes, the full three-kernel geometry is used when modelling compound agent networks (e.g., a validator-repairer pair). For single-loop knowledge agents, the trace equation is sufficient.

## Core API

```go
// Fold a noisy transition matrix to observable states only
tr, err := fullKernel.Trace(observedStateIndices, 1e-12)

// Compute MFPT from every state to every other state
mfpt := tr.MeanFirstPassage()

// Find communicating classes (recurrent traps, transient states)
classes := tr.Classes()

// Compute entropy rate (loop predictability)
H, err := tr.EntropyRate(2)

// Stationary distribution (long-run visit frequency)
pi, err := tr.Stationary(1e-12, 5000)
```

## How soka uses catrace

The `soka suggest` pipeline:

1. **Parse `events.jsonl`** → emit observed event types as a sequence
2. **Estimate transition matrix** (EMA-decayed counts over the event sequence)
3. **Identify $S$ vs $B$**: boundary states (user-specified via `--keep`) are $S$; all others are $B$
4. **Call `Tr(L)`**: fold the full matrix to $S$-only
5. **Call `MeanFirstPassage()`**: compute MFPT from each state to `--target`
6. **`AnalyzeAndSuggest`**: for each edge $(i \to j)$ in $S$, perturb the transition matrix by $\varepsilon$ and re-solve MFPT; rank by MFPT delta
7. **Emit ranked knobs**: the edge whose perturbation most reduces MFPT is the top recommendation

## Relationship to other concepts

- [[Knowledge Agents]] — catrace was designed to model the perceive-decide-act loop that all knowledge agents share; the trace equation makes it applicable to any agent that emits observable milestones
- [[wikigraph]] — the first practical application of catrace: applies MFPT and communicating-class analysis to Markdown wiki link graphs
- [[Arbor]] — the primary event source for soka; catrace's trace equation folds Arbor's noisy phase events into a clean cycle-level chain
- [[Deep Research]] — the harvey-deep-research experiment validated catrace against a real Deep Research event trace (8055 events, 5/6 metrics matched exactly)

## Sources

- McElhose, S. *catrace — Finite-state Markov models of autonomous-agent networks*. GitHub. https://github.com/stephen-mcelhose/catrace
- Hoffman, D., Prakash, C. & Chattopadhyay, S. (2024). *Traces of Consciousness*. Preprints. https://www.preprints.org/manuscript/202410.1305/v1
