#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source bin/_env.sh

need() { command -v "$1" >/dev/null 2>&1 || { echo "error: $1 not installed — run bin/install-tools.sh" >&2; exit 1; }; }
need goimports
need gofumpt

bad=0
unf_imports=$(goimports -l -local github.com/jukuan/harady . || true)
if [ -n "$unf_imports" ]; then
    echo "==> goimports: files with wrong imports:"
    echo "$unf_imports"
    bad=1
fi

unf_fmt=$(gofumpt -l . || true)
if [ -n "$unf_fmt" ]; then
    echo "==> gofumpt: files not formatted:"
    echo "$unf_fmt"
    bad=1
fi

if [ "$bad" -eq 1 ]; then
    echo
    echo "Run: bin/fmt.sh"
    exit 1
fi
echo "==> formatting ok"
