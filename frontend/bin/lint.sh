#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
echo "==> eslint ."
npx eslint .
echo "==> lint ok"
