#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

run() {
    local label="$1"; shift
    printf '\n\033[1m==> %s\033[0m\n' "$label"
    "$@"
}

run "formatting (fmt-check)"    ./bin/fmt-check.sh
run "eslint"                    ./bin/lint.sh
run "typescript (tsc)"          ./bin/typecheck.sh
run "vite build"                npm run build

printf '\n\033[1;32m==> all checks passed\033[0m\n'
