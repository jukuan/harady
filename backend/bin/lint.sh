#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source bin/_env.sh

if ! command -v golangci-lint >/dev/null 2>&1; then
    echo "error: golangci-lint not installed — run bin/install-tools.sh" >&2
    exit 1
fi
FIX=${FIX:-0}
if [ "$FIX" = "1" ]; then
    golangci-lint run --fix ./...
else
    golangci-lint run ./...
fi
echo "==> lint ok"
