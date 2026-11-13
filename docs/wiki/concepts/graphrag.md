---
type: concept
title: GraphRAG
description: Microsoft's graph-based retrieval-augmented generation approach that builds a community-summarised entity knowledge graph to answer global sensemaking questions — a knowledge agent archetype positioned between single-pass RAG and iterative search loops.
tags: [graphrag, microsoft, graph, rag, knowledge-graph, retrieval]
timestamp: 2026-11-13T01:58:49Z
resource: https://arxiv.org/abs/2404.16130
---

# GraphRAG

GraphRAG (Edge et al., 2024) is a graph-based approach to query-focused summarisation over private text corpora, developed at Microsoft Research. Where standard RAG fails on global sensemaking questions ("What are the main themes in the dataset?") because they require whole-corpus reasoning rather than local retrieval, GraphRAG builds a structured knowledge graph from the corpus and uses community summaries as intermediate aggregation units.

GraphRAG represents the **self-healing knowledge base** archetype of [[Knowledge Agents]] when it operates iteratively — ingesting new documents, detecting stale or conflicting entities, and updating the graph. In its basic inference mode it is closer to an enhanced retrieval tool than a full knowledge agent loop; but its graph construction pipeline is itself an agentic, multi-pass process.

## How it works

GraphRAG's pipeline has two stages:

### Indexing (graph construction)
1. LLM extracts entities and relationships from source documents → entity graph
2. Community detection algorithms (e.g., Leiden) partition the graph into hierarchical clusters
3. LLM pregenerates a summary for each community at every level of the hierarchy

### Query (inference)
1. Question arrives; each community summary generates a partial response
2. All partial responses are aggregated into a final answer

This two-stage structure separates the agentic work (graph construction, community summarisation) from the retrieval work (query-time aggregation). The agentic phase is where knowledge-agent dynamics apply.

## Relationship to standard RAG

| Dimension | Standard RAG | GraphRAG |
| --------- | ------------ | -------- |
| Structure | Flat vector index | Entity graph + community summaries |
| Query type | Local (find specific facts) | Global (sensemaking over corpus) |
| Agent loop | Single pass | Multi-pass (indexing phase) |
| Staleness handling | Re-embed on update | Re-detect communities on update |

## Why GraphRAG is relevant to soka

GraphRAG's indexing pipeline is a multi-step agentic loop: extract entities, detect communities, summarise, verify consistency, repeat for new documents. This loop emits observable events at each phase (extraction, clustering, summarisation, validation) — exactly the kind of event stream soka is designed to analyse. MFPT analysis on a GraphRAG indexing trace would surface which phase (extraction latency? community detection instability? summarisation retries?) most lengthens the time-to-consistent-graph.

## Distinction from Deep Research

[[Deep Research]] is an iterative search-and-synthesis loop with no persistent knowledge structure. GraphRAG builds and queries a persistent knowledge graph. The distinction matters for soka:

- Deep Research's events are session-scoped (one question, one loop, one answer)
- GraphRAG indexing events are corpus-scoped (continuous ingestion, long-running maintenance)

Both are addressable by soka; the target states and event schemas differ.

## Relationship to other concepts

- [[Knowledge Agents]] — GraphRAG's indexing pipeline is an instance of the self-healing knowledge base archetype
- [[Deep Research]] — iterative search-and-synthesis; complementary approach to the same global reasoning problem
- [[Optimization]] — a distinct domain; GraphRAG is about knowledge retrieval quality, not computational performance

## Sources

- Edge et al. (2024). *From Local to Global: A Graph RAG Approach to Query-Focused Summarization*. arXiv:2404.16130. https://arxiv.org/abs/2404.16130
- Microsoft GraphRAG GitHub repository. https://github.com/microsoft/graphrag
