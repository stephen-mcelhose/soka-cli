#!/usr/bin/env bash
# experiment: arbor-knowledge-agent
# data:       RUC-NLPIR/Arbor session logs (.arbor/sessions/<run>/events.jsonl)
# status:     Pending
#
# Claim: soka suggest on an Arbor knowledge-agent session surfaces high-MFPT
# bottleneck edges that correspond to the agent's known retrieval failure modes
# (e.g. IDEATE→SEARCH loops when KB coverage is thin), and the top knobs are
# domain-interpretable interventions.
#
# How to fetch fixture:
#   Clone RUC-NLPIR/Arbor, run a session via:
#     python examples/algotune_knn/run.py
#   Then copy the output:
#     cp .arbor/sessions/<run>/events.jsonl testdata/arbor-session.jsonl
#
# State vocabulary: TBD — inspect fixture once available.
#   Candidate boundary states (S): IDEATE, SEARCH, RETRIEVE, SYNTHESIZE, MERGED, ABSTAIN
#   Target: MERGED
#   Noise (B): token deltas, intermediate chunk events

set -euo pipefail

FIXTURE="testdata/arbor-session.jsonl"
SOKA="${SOKA:-./soka}"

# ---------------------------------------------------------------------------
# Preflight
# ---------------------------------------------------------------------------
if [[ ! -f "$FIXTURE" ]]; then
  echo "ERROR: fixture not found at $FIXTURE"
  echo ""
  echo "To generate it:"
  echo "  1. Clone https://github.com/RUC-NLPIR/Arbor"
  echo "  2. Run a session: python examples/algotune_knn/run.py"
  echo "  3. Copy output:   cp .arbor/sessions/<run>/events.jsonl testdata/arbor-session.jsonl"
  exit 1
fi

if [[ ! -x "$SOKA" ]]; then
  echo "Building soka..."
  go build -o soka .
fi

# ---------------------------------------------------------------------------
# TODO: fill in --keep and --knobs once state vocabulary is known from fixture
# ---------------------------------------------------------------------------
echo "=== Arbor: full kernel ==="
"$SOKA" suggest \
  --target MERGED \
  --top-k 5 \
  "$FIXTURE"

# ---------------------------------------------------------------------------
# Results (fill in after run)
# ---------------------------------------------------------------------------
# events_processed:
# lines_skipped:
# boundary_states:
# baseline_mfpt:
# top knob (from → to):
# mfpt_delta:
# domain interpretation:
