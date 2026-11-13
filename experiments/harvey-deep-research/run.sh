#!/usr/bin/env bash
# experiment: harvey-deep-research
# data:       harveyai/deep-research-starter (Google deep-research output)
# status:     Pending
#
# Claim: soka suggest, end-to-end against a real deep-research event stream,
# surfaces the same top knobs (output_item.done → completed) and MFPT delta
# directionality validated in catrace experiment stream-deep-research-trace.
#
# Reference: catrace/experiments/stream-deep-research-trace (5/6 supported)
#
# Boundary states (S, --keep):
#   response.created, response.in_progress, response.completed,
#   response.output_item.added, response.output_item.done,
#   response.web_search_call.in_progress, response.web_search_call.searching,
#   response.web_search_call.completed, response.content_part.added,
#   response.content_part.done, response.output_text.done,
#   response.reasoning_summary_part.added, response.reasoning_summary_part.done,
#   response.reasoning_summary_text.done
#
# Target:  response.completed
# Knobs:   response.output_item.done,response.web_search_call.completed,response.reasoning_summary_text.done

set -euo pipefail

FIXTURE="testdata/harvey-deep-research.jsonl"
SOKA="${SOKA:-./soka}"

# ---------------------------------------------------------------------------
# 1. Fetch fixture
# ---------------------------------------------------------------------------
if [[ ! -f "$FIXTURE" ]]; then
  echo "Fetching fixture..."
  mkdir -p testdata
  curl -fsSL -o "$FIXTURE" \
    https://raw.githubusercontent.com/harveyai/deep-research-starter/main/output_samples/events_jsonl.jsonl
fi

# ---------------------------------------------------------------------------
# 2. Build soka if binary missing
# ---------------------------------------------------------------------------
if [[ ! -x "$SOKA" ]]; then
  echo "Building soka..."
  go build -o soka .
fi

# ---------------------------------------------------------------------------
# 3. Run: full kernel (no Trace fold)
# ---------------------------------------------------------------------------
echo ""
echo "=== A: full kernel ==="
"$SOKA" suggest \
  --target response.completed \
  --top-k 5 \
  "$FIXTURE"

# ---------------------------------------------------------------------------
# 4. Run: Trace-folded onto boundary states S
# ---------------------------------------------------------------------------
KEEP="response.created,response.in_progress,response.completed,\
response.output_item.added,response.output_item.done,\
response.web_search_call.in_progress,response.web_search_call.searching,\
response.web_search_call.completed,response.content_part.added,\
response.content_part.done,response.output_text.done,\
response.reasoning_summary_part.added,response.reasoning_summary_part.done,\
response.reasoning_summary_text.done"

KNOBS="response.output_item.done,response.web_search_call.completed,response.reasoning_summary_text.done"

echo ""
echo "=== B: Trace-folded, ε=0.05 ==="
"$SOKA" suggest \
  --target response.completed \
  --keep   "$KEEP" \
  --knobs  "$KNOBS" \
  --eps    0.05 \
  --top-k  5 \
  "$FIXTURE"

echo ""
echo "=== C: Trace-folded, ε=0.15 ==="
"$SOKA" suggest \
  --target response.completed \
  --keep   "$KEEP" \
  --knobs  "$KNOBS" \
  --eps    0.15 \
  --top-k  5 \
  "$FIXTURE"

echo ""
echo "=== D: Trace-folded, ε=0.30 ==="
"$SOKA" suggest \
  --target response.completed \
  --keep   "$KEEP" \
  --knobs  "$KNOBS" \
  --eps    0.30 \
  --top-k  5 \
  "$FIXTURE"

# ---------------------------------------------------------------------------
# Results (fill in after run)
# ---------------------------------------------------------------------------
# events_processed:
# lines_skipped:
# boundary_states (traced):
# entropy_rate (full vs traced):
# top knob (from → to):
# mfpt_delta @ ε=0.05 / 0.15 / 0.30:
# monotonic |Δ| across ε sweep: yes/no
