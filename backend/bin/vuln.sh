#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source bin/_env.sh

if ! command -v govulncheck >/dev/null 2>&1; then
    echo "error: govulncheck not installed — run bin/install-tools.sh" >&2
    exit 1
fi
govulncheck ./...
echo "==> vuln check ok"
