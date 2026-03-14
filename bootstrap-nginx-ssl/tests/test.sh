#!/bin/bash
set -e

cd "$(dirname "$0")/.."

echo "=== Building test container ==="
docker build -f tests/Dockerfile.test -t bootstrap-test .

echo ""
echo "=== Running tests in Ubuntu container ==="
docker run --rm bootstrap-test
