#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKSPACE_DIR="${RALPH_WORKSPACE_DIR:-$PWD}"
TOOL="${RALPH_BATCH_TOOL:-claude}"
MAX_ITERATIONS="${RALPH_BATCH_MAX_ITERATIONS:-50}"

PRDS=(
  "$SCRIPT_DIR/prds/prd-feature-example.json"
  "$SCRIPT_DIR/prds/prd-bugfix-example.json"
)

for prd in "${PRDS[@]}"; do
  echo "=== Starting: $prd ==="

  cp "$prd" "$SCRIPT_DIR/prd.json"
  "$SCRIPT_DIR/ralph.sh" --tool "$TOOL" "$MAX_ITERATIONS"

  git -C "$WORKSPACE_DIR" checkout main
  echo "=== Completed: $prd ==="
done
