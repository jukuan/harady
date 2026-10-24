#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source bin/_env.sh
go vet ./...
echo "==> vet ok"
