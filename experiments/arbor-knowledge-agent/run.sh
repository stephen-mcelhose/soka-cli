#!/usr/bin/env bash
# experiment: arbor-knowledge-agent
# data:       RUC-NLPIR/Arbor bundled demo session (algotune_knn, CPU-only)
# status:     Pending
#
# Claim: soka suggest on an Arbor research-agent session surfaces the dominant
# bottleneck in the hypothesis-tree cycle (the edge whose MFPT delta is largest)
# and that edge is domain-interpretable: it corresponds to a decision transition
# inside the Arbor Cycle (observe→ideate→select→dispatch→backprop→decide).
#
# Fixture source: bundled in pip install arbor-agent at:
#   arbor/cli/assets/demo_session/events.jsonl  (117 events, algotune_knn run)
#
# State vocabulary
# ----------------
# B (fold — noise):   llm.thinking_delta, llm.call, tool.start, tool.end
# S (keep — milestones):
#   session.start, cycle.start, cycle.phase, cycle.end,
#   executor.start, executor.end,
#   idea.proposed, idea.completed, idea.merged, idea.pruned,
#   convergence.reached, session.end
# Target: idea.merged
# Knobs:  idea.proposed, executor.end, cycle.phase
#
# Note: 117 events is sparse — MFPT estimates carry higher uncertainty than
# the harvey run (8055 events). Results are directional, not authoritative.

set -euo pipefail

FIXTURE="${FIXTURE:-testdata/arbor-demo.jsonl}"
SOKA="${SOKA:-./soka}"

KEEP="session.start,cycle.start,cycle.phase:observe,cycle.phase:ideate,cycle.phase:select,cycle.phase:dispatch,cycle.phase:backprop,cycle.phase:decide,cycle.end,executor.start,executor.end,idea.proposed,idea.completed,idea.merged,idea.pruned,convergence.reached,session.end"
KNOBS="idea.proposed,executor.end,cycle.phase:dispatch,cycle.phase:decide"

# ---------------------------------------------------------------------------
# 1. Source fixture from bundled arbor-agent install
# ---------------------------------------------------------------------------
if [[ ! -f "$FIXTURE" ]]; then
  echo "Locating bundled arbor demo session..."
  mkdir -p testdata
  BUNDLED=$(python3 -c "
import pathlib, arbor.cli.replay as r
print(r.DEMO_DIR / 'events.jsonl')
" 2>/dev/null) || {
    echo "ERROR: arbor-agent not installed. Run: pip install arbor-agent"
    exit 1
  }
  if [[ ! -f "$BUNDLED" ]]; then
    echo "ERROR: bundled demo events.jsonl not found at $BUNDLED"
    exit 1
  fi
  cp "$BUNDLED" "$FIXTURE"
  echo "Copied $(wc -l < "$FIXTURE") events from $BUNDLED"
fi

# ---------------------------------------------------------------------------
# 2. Build soka if binary missing
# ---------------------------------------------------------------------------
if [[ ! -x "$SOKA" ]]; then
  echo "Building soka..."
  go build -o soka .
fi

# ---------------------------------------------------------------------------
# 3. Inspect raw event-type vocabulary
# ---------------------------------------------------------------------------
echo ""
echo "=== Event type distribution ==="
python3 -c "
import sys, json, pathlib
counts = {}
for line in pathlib.Path('$FIXTURE').read_text().splitlines():
    line = line.strip()
    if not line: continue
    t = json.loads(line).get('type', '')
    counts[t] = counts.get(t, 0) + 1
for k, v in sorted(counts.items()):
    print(f'{v:4d}  {k}')
print(f'---')
print(f'{sum(counts.values()):4d}  total')
"

# ---------------------------------------------------------------------------
# 4. Run: full kernel (no Trace fold)
#    Note: idea.merged has null MFPT (paths escape to session.end first).
#    Use session.end as target throughout.
# ---------------------------------------------------------------------------
echo ""
echo "=== A: full kernel, target=session.end ==="
"$SOKA" suggest \
  --target session.end \
  --top-k 5 \
  "$FIXTURE"

# ---------------------------------------------------------------------------
# 5. Run: Trace-folded onto boundary states S, ε sweep
# ---------------------------------------------------------------------------
for eps in 0.05 0.15 0.30; do
  echo ""
  echo "=== B: Trace-folded, target=session.end, ε=$eps ==="
  "$SOKA" suggest \
    --target session.end \
    --keep   "$KEEP" \
    --knobs  "$KNOBS" \
    --eps    "$eps" \
    --top-k  3 \
    "$FIXTURE"
done

# ---------------------------------------------------------------------------
# Results (fill in after run → results.md)
# ---------------------------------------------------------------------------
