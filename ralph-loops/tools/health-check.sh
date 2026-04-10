#!/bin/bash
# health-check.sh
# Verifies the local dev server is running before starting a QA or E2E run.
# Usage: ./tools/health-check.sh [url] [max_retries]
#
# Examples:
#   ./tools/health-check.sh
#   ./tools/health-check.sh http://localhost:3000/health
#   ./tools/health-check.sh http://localhost:5000/health 10

URL="${1:-http://localhost:3000/health}"
MAX_RETRIES="${2:-5}"
RETRY_INTERVAL=2

echo "Checking server at $URL (max $MAX_RETRIES retries)..."

for i in $(seq 1 "$MAX_RETRIES"); do
  STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$URL" 2>/dev/null)

  if [[ "$STATUS" == "200" ]]; then
    echo "Server is up (HTTP $STATUS)"
    exit 0
  fi

  echo "  Attempt $i/$MAX_RETRIES — got HTTP $STATUS. Retrying in ${RETRY_INTERVAL}s..."
  sleep "$RETRY_INTERVAL"
done

echo "Server did not respond after $MAX_RETRIES attempts. Aborting."
exit 1
