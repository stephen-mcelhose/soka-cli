# soka-cli

**Self Optimizing Knowledge Agents** — a Go CLI for Markov-based policy analysis of agent event streams.

Backed by [catrace](https://github.com/stephen-mcelhose/catrace) v0.1.1.

---

## Install

```sh
go install github.com/stephen-mcelhose/soka-cli@latest
```

Or build from source:

```sh
git clone https://github.com/stephen-mcelhose/soka-cli.git
cd soka-cli
go build -o soka .
```

---

## Commands

### `soka suggest`

Ingest a JSONL agent event stream, estimate a Markov kernel, and rank single-row policy perturbations that reduce mean first-passage time (MFPT) to a target goal state.

```sh
soka suggest --target answer:done events.jsonl
cat events.jsonl | soka suggest --target CLAIM_VERIFIED --top-k 3
```

**Flags**

| Flag          | Default | Description                                               |
| ------------- | ------- | --------------------------------------------------------- |
| `--target`    | —       | Goal state to minimize MFPT toward **(required)**         |
| `--keep`      | —       | Comma-separated boundary states to retain after Trace     |
| `--knobs`     | —       | Comma-separated controllable from-states (default: all)   |
| `--halflife`  | 0       | EMA half-life in events (0 = infinite memory)             |
| `--eps`       | 0.10    | Perturbation mass epsilon                                 |
| `--top-k`     | 5       | Number of top interventions to return                     |

**Output** (JSON)

```json
{
  "events_processed": 412,
  "lines_skipped": 3,
  "boundary_states": ["QUERY_DECOMPOSED", "CLAIM_VERIFIED", "ABSTAINED_DEAD_END"],
  "entropy_rate_bits": 1.42,
  "classes": ["ergodic"],
  "baseline_mfpt": 8.3,
  "suggestions": [
    {
      "from": "CONTRADICTION_FOUND",
      "to": "WEB_FALLBACK",
      "epsilon": 0.1,
      "baseline_mfpt": 8.3,
      "projected_mfpt": 3.4,
      "mfpt_delta": -4.9
    }
  ]
}
```

### `soka mcp`

Start a stdio-based MCP server exposing Markov analysis as callable tools — compatible with LangGraph, LlamaIndex, Claude Agent SDK, and any MCP client.

> **Coming soon.**

---

## Event format

Events are JSONL, one JSON object per line. Each event must carry a `state` field (or `status` as a fallback) identifying the agent's current boundary state:

```jsonl
{"state": "QUERY_DECOMPOSED", "topic": "retrieval", "ts": "2025-01-01T00:00:00Z"}
{"state": "VECTOR_RETRIEVED",  "topic": "retrieval"}
{"state": "CONTRADICTION_FOUND"}
{"state": "WEB_FALLBACK"}
{"state": "CLAIM_VERIFIED"}
```

Malformed lines are skipped and counted in `lines_skipped`.

---

## License

[MIT](LICENSE) © 2025 Stephen McElhose
