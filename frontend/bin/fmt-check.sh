#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
echo "==> prettier --check ."
npx prettier --check .
echo "==> formatting ok"
