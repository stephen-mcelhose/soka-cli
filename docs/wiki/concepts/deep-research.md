---
type: concept
title: Deep Research
description: OpenAI's agentic research capability that runs multi-step iterative web search and synthesis to produce comprehensive reports — the canonical "researcher" archetype of knowledge agents, and a primary soka case study alongside Arbor.
tags: [deep-research, openai, web-search, iterative-synthesis, knowledge-agents]
timestamp: 2026-11-13T01:58:49Z
resource: https://openai.com/index/introducing-deep-research/
---

# Deep Research

Deep Research is an agentic capability introduced by OpenAI in February 2025, powered by a version of the o3 model fine-tuned for agentic web search. Given a complex research question, it autonomously searches, reads, and synthesises hundreds of online sources into a comprehensive report — without human steering during the loop. It achieved a 72.57% pass rate on GAIA (state-of-the-art at launch) and 26.6% on Humanity's Last Exam.

Deep Research is the canonical instantiation of the **iterative search-and-synthesis** archetype of [[Knowledge Agents]]. Its loop is structurally identical to [[Arbor]]'s loop, with web queries replacing code evaluations as the unit of work: search, observe evidence, reason about contradictions, search again, synthesise.

## The loop structure

Deep Research's behavioral topology maps onto the knowledge agent signature:

| Knowledge agent phase | Deep Research equivalent |
| --------------------- | ------------------------ |
| `cycle.phase:observe` | Assess what is known; identify gaps |
| `cycle.phase:ideate` | Plan next search queries |
| `cycle.phase:dispatch` | Issue web searches in parallel |
| `eval.end` | Evaluate retrieved evidence |
| `cycle.phase:backprop` | Integrate evidence; resolve contradictions |
| `cycle.phase:decide` | Enough? → synthesise and emit report; else → continue |
| `session.end` | Final report delivered |

This structural equivalence means that an `events.jsonl` trace from a Deep Research-style system (with the same milestone event types) would yield the same kind of MFPT analysis that soka runs on Arbor traces.

## Why Deep Research is a soka proof point

1. **Cost** — each Deep Research session consumes significant compute; reducing expected loop length directly reduces cost per report.
2. **Thrashing patterns** — over-searching (too many queries before synthesising) is structurally equivalent to [[Arbor]]'s over-ideating (too many hypotheses before executing). soka surfaces both via the same MFPT analysis.
3. **Scale** — Deep Research runs at consumer scale; even a small MFPT reduction across millions of sessions has large aggregate impact.
4. **Event log availability** — o3-deep-research traces are an accessible corpus for validation experiments, complementing [[Arbor]]'s `events.jsonl`.

## Distinction from retrieval-augmented generation

Standard RAG (and [[GraphRAG]]) is a single-pass retrieval-then-generation pipeline. Deep Research is a multi-pass, iterative loop with self-directed query planning. The distinction is behavioral: RAG has no cycle; Deep Research has many. This makes Deep Research a [[Knowledge Agents|knowledge agent]] and RAG a retrieval tool.

## Relationship to other concepts

- [[Knowledge Agents]] — Deep Research is the "researcher" archetype; defines the iterative search-and-synthesis class
- [[Arbor]] — the "coder" archetype; structurally parallel loop with code evaluation instead of web search
- [[GraphRAG]] — related but single-pass; a retrieval tool that may feed into a knowledge agent loop
- [[Optimization]] — a different domain goal (performance metric vs. comprehensive report)

## Sources

- OpenAI (2025, February 2). *Introducing deep research*. https://openai.com/index/introducing-deep-research/
- OpenAI (2025). *GAIA benchmark results*. Cited in the deep research announcement.
