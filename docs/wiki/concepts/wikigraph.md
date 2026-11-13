---
type: concept
title: wikigraph
description: A Go CLI that models a Markdown wiki as a Markov chain over its [[wikilink]] graph — computing stationary distributions, MFPT-based learning paths, and communicating-class health reports. The first practical application of catrace.
tags: [wikigraph, wiki, markov-chain, mfpt, knowledge-graph, go, catrace]
timestamp: 2026-11-13T01:58:49Z
resource: https://github.com/stephen-mcelhose/wikigraph
---

# wikigraph

wikigraph is a Go CLI that treats a Markdown wiki — a directory of `.md` files connected by `[[slug]]` wikilinks — as a finite-state Markov chain and applies [[catrace]]'s mathematical machinery to it. It is the first practical application of catrace: where catrace models autonomous agent loops, wikigraph models a wiki reader as an autonomous agent whose "actions" are clicking links and whose "world states" are pages.

The resulting analysis goes far beyond a simple link graph: stationary distributions reveal which pages are genuinely authoritative (not just frequently linked); communicating classes reveal structural silos; MFPT surfaces the fastest paths from any page to a designated goal.

## The Markov model

wikigraph represents a wiki reader as a random walker:

- **States**: wiki pages (one per `.md` file)
- **Transitions**: following a wikilink from the current page, chosen uniformly at random
- **Sink handling**: pages with no outbound links (sinks) teleport the walker uniformly to any page, preventing absorption

Under this model, all of [[catrace]]'s analytical machinery applies directly:

| catrace primitive | wikigraph interpretation |
| ----------------- | ------------------------ |
| Stationary distribution $\pi_i$ | Long-run probability of being on page $i$ — true "authority" |
| Communicating classes | Groups of pages that are mutually reachable — structural silos visible at a glance |
| MFPT$(i \to j)$ | Expected clicks from page $i$ to reach page $j$ — the basis for learning paths |
| Entropy rate | How predictable the reader's path is — high entropy = wiki is a maze |

## Core subcommands

### `wikigraph graph`
Generates a self-contained interactive HTML force-directed graph. Node size = stationary distribution (authority); node colour = communicating class (silo); edge width = transition probability. No server required.

### `wikigraph goal <target-page>`
Computes MFPT from every page to `<target-page>` and renders a ranked list: the shortest expected paths to the goal. Use this to generate curated learning paths through a knowledge base, or to identify which pages are "far" from a canonical destination.

### `wikigraph analyze`
Wiki health report:
- **Orphaned pages** — no inbound links; invisible to random walkers
- **Sinks** — no outbound links; break the random walk
- **Central pages** — highest stationary distribution (most authoritative)
- **Missing link suggestions** — pages that co-occur in links from common parents but don't link to each other

### `wikigraph export`
Exports the computed graph metrics to JSON, CSV, or DOT for external network analysis.

## Why wikigraph is relevant to this project

Our `docs/wiki/` uses `[[wiki-links]]` and follows the doc-manager schema. wikigraph can run directly against it:

```bash
wikigraph analyze --recursive docs/wiki/
wikigraph goal "Knowledge Agents" docs/wiki/concepts/
```

This surfaces:
- Which concept pages are orphaned (no concept links to them yet)
- Which pages are genuinely central vs. merely self-proclaimed hubs
- The expected number of wikilink hops from any entry point to the core concepts

As the wiki grows, running `wikigraph analyze` as a lint step ensures the link graph stays connected and authoritative.

## The vectorize roadmap

wikigraph has an active ADR (`docs/adr-001-embedding-layer.md`) for a `wikigraph vectorize` command that adds a semantic embedding layer. This would transition wikigraph from purely structural analysis (topology of links) to hybrid semantic-and-topological analysis — finding pages that *should* be linked based on content similarity, even if no explicit link exists. This is the direction toward a true self-healing knowledge base.

## Relationship to other concepts

- [[catrace]] — the mathematical engine wikigraph applies to wiki link graphs; MFPT, stationary distribution, communicating classes all come from catrace
- [[Knowledge Agents]] — wikigraph models wiki readers as knowledge agents; `wikigraph goal` is a knowledge-agent routing advisor
- [[GraphRAG]] — complementary: GraphRAG builds entity graphs from document content; wikigraph analyses the explicit link structure of a human-curated wiki; the two approaches can compose
- [[Optimization]] — wikigraph's `analyze` and `goal` commands are optimization tools for wiki structure, not just visualization

## Sources

- McElhose, S. *wikigraph — Markov-chain analysis of Markdown wikis*. GitHub. https://github.com/stephen-mcelhose/wikigraph
- McElhose, S. *catrace*. GitHub. https://github.com/stephen-mcelhose/catrace
