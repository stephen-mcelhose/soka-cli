---
type: concept
title: Hypothesis-Tree Refinement
description: A structured search strategy for autonomous agents that organises candidate solutions as a tree, propagates empirical evidence upward through backpropagation, and iteratively expands the most promising frontier nodes.
tags: [hypothesis-tree, search, arbor, iterative-refinement, autonomous-agents]
timestamp: 2026-11-13T01:58:49Z
---

# Hypothesis-Tree Refinement

Hypothesis-Tree Refinement (HTR) is the core search strategy used by [[Arbor]] and described in Jin et al. (2026). It addresses a fundamental problem in autonomous research agents: without explicit memory of *why* a direction failed, agents repeat the same mistakes and converge to local optima. HTR solves this by organising the agent's exploration history as a tree, where every node records its mechanistic hypothesis, its experimental outcome, and the insights derived from that outcome.

The key insight is that the tree is not just a log — it is a live planning artifact. At each cycle, the coordinator reads the full tree to decide which frontier to expand next, informed by all prior results. Failed branches are pruned but not discarded; their insights are propagated upward and become constraints on sibling and parent hypotheses.

## Tree structure

```
ROOT (baseline solution, score = s₀)
├── Node 1: "switch to float32 for GEMM"  → score=9.78 ✓ [merged]
│   ├── Node 1.1: "add query chunking"    → score=10.2 ✓ [merged]
│   │   └── Node 1.1.1: "drop argsort"   → score=9.13 ✗ [pruned]
│   └── Node 1.2: "reduce chunk size"    → score=9.61 ✗ [pruned]
└── Node 2: "approximate nearest-neighbour" → [pending]
```

Each node carries:
- **Hypothesis** — the mechanistic claim and expected observable
- **Branch** — an isolated git worktree for the implementation
- **Score** — the evaluator output after running
- **Insight** — what was learned, propagated upward on completion

## The six-phase cycle

HTR maps directly onto [[Arbor]]'s cycle phases:

| Phase | HTR operation |
| ----- | ------------- |
| `observe` | Read the full tree; identify best frontier nodes and accumulated insights |
| `ideate` | Propose child hypotheses at the frontier, grounded in prior insights |
| `select` | Rank candidates by expected information gain; choose top K |
| `dispatch` | Spawn executors; each runs one hypothesis in isolation |
| `backprop` | Record scores; propagate insights toward ROOT; update constraints |
| `decide` | Merge winners to trunk, prune dead branches, check convergence |

## Why a tree over a list

A flat hypothesis list (try things, pick the best) has no memory of *structure*. HTR's tree makes three things explicit:

1. **Inheritance** — a child hypothesis builds on its parent's implementation; partial wins compound
2. **Pruning signals** — a failed node signals that its entire subtree direction is likely unpromising
3. **Diversity** — the coordinator can maintain multiple independent frontier branches, avoiding premature convergence

## Relationship to other search strategies

| Strategy | Memory structure | Branching | Evidence reuse |
| -------- | ---------------- | --------- | -------------- |
| Random search | None | Unlimited | None |
| Beam search | Fixed-width list | Bounded | Partial |
| **HTR** | **Full tree** | **Adaptive** | **Backpropagated** |
| Monte Carlo Tree Search | Tree + rollouts | Adaptive | Statistical (UCB) |

HTR is closest to MCTS but substitutes statistical rollouts with real experiment evaluations and learned natural-language insights.

## Relationship to other concepts

- [[Arbor]] — the agent that implements HTR as its core search loop
- [[Knowledge Agents]] — HTR is one instantiation of the broader search–reason–verify topology
- [[Optimization]] — the domain in which HTR's hypothesis nodes correspond to concrete code changes

## Sources

- Jin et al. (2026). *Toward Generalist Autonomous Research via Hypothesis-Tree Refinement*. arXiv:2606.11926. https://arxiv.org/abs/2606.11926
- RUC-NLPIR. *Arbor*. GitHub repository. https://github.com/RUC-NLPIR/Arbor
