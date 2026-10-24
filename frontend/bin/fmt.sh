#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
echo "==> prettier --write ."
npx prettier --write .
echo "==> done"
