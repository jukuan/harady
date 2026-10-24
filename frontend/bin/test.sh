#!/usr/bin/env bash
# Frontend test runner (Vitest).
#
# Usage:
#   ./bin/test.sh                    # run once
#   ./bin/test.sh --watch            # watch mode (dev)
#   ./bin/test.sh --ui               # open Vitest UI in browser
#   ./bin/test.sh src/store.test.ts  # one file
#   ./bin/test.sh -t "playerOut"     # tests matching name
#   COVERAGE=1 ./bin/test.sh         # coverage report

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

args=()
if [ "${COVERAGE:-0}" = "1" ]; then
    args+=(--coverage)
fi
args+=("$@")

npx vitest run "${args[@]}"
