#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source bin/_env.sh

run() {
    local label="$1"; shift
    printf '\n\033[1m==> %s\033[0m\n' "$label"
    "$@"
}

run "formatting (fmt-check)"    ./bin/fmt-check.sh
run "go vet"                    ./bin/vet.sh
run "golangci-lint"             ./bin/lint.sh
run "govulncheck"               ./bin/vuln.sh
run "tests (race)"              go test -race -count=1 ./...

printf '\n\033[1;32m==> all checks passed\033[0m\n'
