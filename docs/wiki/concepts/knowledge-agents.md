---
type: concept
title: Knowledge Agents
description: AI systems defined by a shared behavioral topology — iterative search, reasoning, verification, and looping until a terminal goal state — regardless of framework or domain.
tags: [knowledge-agents, ai-agents, iterative-loops, react, soka]
timestamp: 2026-11-13T01:58:49Z
---

# Knowledge Agents

A **knowledge agent** is any AI system whose execution has the shape:

> **Search → Reason → Verify → Loop → Terminal**

The defining characteristic is not what framework powers the system, nor what domain it operates in — it is the *behavioral topology*: a closed loop that iterates until it reaches a terminal goal state. The ReAct framework (Yao et al., 2023) first formalised this loop as an interleaving of Thought, Action, and Observation steps that continues until a goal is reached or a budget is exhausted. Modern instantiations scale this pattern across web browsing, code execution, and graph traversal.

This framing is what makes soka's [[Markov Chain]] analysis broadly applicable: given a JSONL event log from any such system, you can fold its noisy execution trace into a compact state machine, estimate the [[Mean First Passage Time]] to the terminal state, and rank the routing tweaks with the highest leverage — without knowing anything about the agent's internals.

## Three archetypes

### Iterative Search & Synthesis
Systems that repeatedly query sources, evaluate evidence, resolve contradictions, and draft an answer. The loop terminates when a confidence threshold is met or a budget is exhausted. [[Deep Research]] (OpenAI, February 2025) is the canonical example, running multi-step web research and synthesising hundreds of sources into a final report. [[GraphRAG]] pipelines also fit here when they run multi-hop retrieval with self-critique.

### Self-Healing Knowledge Bases
Agents that ingest new documents, detect stale or conflicting facts, and update a canonical knowledge graph or wiki. The loop runs on a schedule or on document arrival; the terminal state is a consistent, up-to-date knowledge base.

### Multi-Turn Coding & Investigation Agents
Coding assistants or SRE incident-response agents that hypothesize, run tests or queries, observe errors, and retry until reaching `TESTS_PASS` or `INCIDENT_RESOLVED`. [[Arbor]] is a research-focused instance of this archetype applied to algorithmic [[Optimization]].

## Why behavioral topology over framework

Defining knowledge agents by topology — not by product, API, or event schema — means:

1. **soka works on any JSONL stream** with identifiable milestones (`session.start`, `cycle.end`, etc.) and a reachable target state. The agent does not need to know about soka.
2. **Insights transfer across domains.** A thrashing pattern in a [[Deep Research]] loop (over-searching before synthesising) and a thrashing pattern in an [[Arbor]] loop (over-ideating before executing) are structurally identical. The same MFPT analysis surfaces both.
3. **The beachhead generalises.** Autonomous research agents like [[Arbor]] are the sharpest initial use case — they emit clean `events.jsonl` and suffer from expensive multi-step loops — but every archetype above is addressable.

## The behavioral signature in event logs

A knowledge agent's event log almost always contains:

| Event class           | Examples                                                 |
| --------------------- | -------------------------------------------------------- |
| **Session boundaries** | `session.start`, `session.end`                          |
| **Cycle markers**      | `cycle.start`, `cycle.end`, `turn.start`                |
| **Phase transitions**  | `phase:search`, `phase:reason`, `phase:verify`          |
| **Work units**         | `idea.proposed`, `query.issued`, `patch.applied`        |
| **Outcomes**           | `eval.end`, `answer.accepted`, `test.passed`            |
| **Terminal states**    | `convergence.reached`, `budget.exhausted`, `goal.achieved` |

soka reads this stream and constructs a [[Markov Chain]] over the milestone-level states, then computes [[Mean First Passage Time]] from each state to the target.

## Relationship to other concepts

- [[Arbor]] — concrete knowledge agent for algorithmic optimization; the primary beachhead and event source
- [[Deep Research]] — iterative web search and synthesis; the canonical "researcher" archetype
- [[GraphRAG]] — graph-based retrieval with multi-hop reasoning; the "librarian" archetype
- [[Hypothesis-Tree Refinement]] — the search strategy used inside Arbor and other exploration agents
- [[Optimization]] — the domain Arbor operates in; also an example of how domain knowledge shapes loop structure

## Sources

- Yao et al. (2023). *ReAct: Synergizing Reasoning and Acting in Language Models*. ICLR 2023. https://arxiv.org/abs/2210.03629
- Weng, L. (2023). *LLM Powered Autonomous Agents*. Lil'Log. https://lilianweng.github.io/posts/2023-06-23-agent/
