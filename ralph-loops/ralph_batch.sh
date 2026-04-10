#!/bin/bash
# milestones.sh

PDRS=(
  "scripts/ralph/prds/prd-bugfix-qa.json"
  "scripts/ralph/prds/prd-qa-regression.json"
)

for prd in "${PDRS[@]}"; do
  echo "=== Iniciando: $prd ==="

  cp "$prd" scripts/ralph/prd.json
  ./scripts/ralph/ralph.sh --tool claude 50

  git checkout main
  echo "=== Concluído: $prd ==="
done