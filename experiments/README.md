# soka-cli experiments

End-to-end CLI experiments against real agent event streams.

Each experiment is a shell script that fetches its own fixture into `testdata/` (gitignored) and runs `soka suggest` with the relevant flags. Results are recorded inline in the script as comments.

## Running an experiment

```sh
bash experiments/<slug>/run.sh
```

No build step needed beyond `go build -o soka .` in the repo root.

## Registry

| Slug                      | Data source                              | Status   |
| ------------------------- | ---------------------------------------- | -------- |
| `harvey-deep-research`    | harveyai/deep-research-starter (Google)  | Complete — 11/11 aligned with catrace |
| `arbor-knowledge-agent`   | RUC-NLPIR/Arbor bundled demo (algotune_knn, CPU-only) | Complete — dispatch bottleneck confirmed |

## Infrastructure notes

### events.jsonl — now emitted natively (session 2)

`arbor_state.py` previously had no event log. Events were reconstructed
post-hoc from the idea tree and eval logs, which made the JSONL approximate.

`arbor_state.py` now writes `events.jsonl` live inside the session directory
(`.arbor/sessions/<run-name>/events.jsonl`) as each command runs:

| Command      | Events emitted automatically                              |
| ------------ | --------------------------------------------------------- |
| `init`       | `session.start`                                           |
| `add`        | `idea.proposed`                                           |
| `prune`      | `idea.pruned`                                             |
| `update`     | `executor.start / executor.end / idea.merged / idea.pruned` (on status change) |
| `eval`       | `eval.end`                                                |
| `worktree`   | `executor.start`                                          |
| `record`     | `executor.end`                                            |
| `merge`      | `idea.merged`                                             |

Cycle/phase boundary events (`session.end`, `cycle.start`, `cycle.end`,
`cycle.phase:*`) are emitted via the new `emit` subcommand, which the
orchestrator skill calls explicitly at each phase transition.

### Arbor fork — stephen-mcelhose/arbor (session 2)

Forked `RUC-NLPIR/Arbor` (Apache-2.0) to add `provider: process` — a BYOA
(bring your own agent) LLM provider that routes every coordinator LLM call
to an external command over stdin/stdout. No API key or OAuth required;
any agent (Claude Code, local model, etc.) can serve as the brain for a
native `arbor run` and get real `events.jsonl` via the native EventBus.

Config:
```yaml
llm:
  provider: process
  agent_cmd: /path/to/bridge-script
```

The bridge reads `{system, messages, tools, max_tokens}` JSON from stdin and
writes `{stop_reason, content}` JSON to stdout.
