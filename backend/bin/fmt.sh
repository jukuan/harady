#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source bin/_env.sh

need() { command -v "$1" >/dev/null 2>&1 || { echo "error: $1 not installed — run bin/install-tools.sh" >&2; exit 1; }; }
need goimports
need gofumpt

echo "==> goimports -w ."
goimports -w -local github.com/jukuan/harady .

echo "==> gofumpt -w ."
gofumpt -w .

echo "==> done"
