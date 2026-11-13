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
| `arbor-knowledge-agent`   | RUC-NLPIR/Arbor session logs             | Pending  |
