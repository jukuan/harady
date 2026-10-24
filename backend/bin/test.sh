#!/usr/bin/env bash
# Runs the backend test suite.
#
# Default behaviour: -race -count=1 over ./...
# Extra args are passed straight to `go test`, so any Go test flag works:
#   ./bin/test.sh                          # everything, race, no cache
#   ./bin/test.sh ./internal/game          # one package
#   ./bin/test.sh -run TestSubmitCity      # one test by name
#   ./bin/test.sh -short                   # skip slow tests (bot-driven ones)
#   ./bin/test.sh ./... -run TestWS_       # all WS smoke tests
#
# Env toggles:
#   COVERAGE=1 ./bin/test.sh               # coverage + html report (coverage.html)
#   VERBOSE=1  ./bin/test.sh               # -v (test names + t.Log)
#   TIMEOUT=10m ./bin/test.sh              # custom -timeout (default 5m)

set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

# ---- default args, appended with any user args ----------------------------
args=(-race -count=1 -timeout "${TIMEOUT:-5m}")

if [ "${VERBOSE:-0}" = "1" ]; then
    args+=(-v)
fi

if [ "${COVERAGE:-0}" = "1" ]; then
    args+=(-coverprofile=coverage.out -covermode=atomic)
fi

# If the caller didn't provide any packages / patterns, default to ./...
# `go test` with only flags and no args implies ".", which would miss
# subpackages — we want the whole tree by default.
if [ "$#" -eq 0 ]; then
    args+=(./...)
else
    args+=("$@")
fi

echo "==> go test ${args[*]}"
go test "${args[@]}"

# ---- optional coverage report --------------------------------------------
if [ "${COVERAGE:-0}" = "1" ] && [ -f coverage.out ]; then
    echo
    echo "==> go tool cover -func=coverage.out"
    go tool cover -func=coverage.out | tail -n +1
    go tool cover -html=coverage.out -o coverage.html
    echo
    echo "==> coverage report: $(pwd)/coverage.html"
fi

echo
echo "==> all tests passed"
