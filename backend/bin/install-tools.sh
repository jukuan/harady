#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
source bin/_env.sh

GOBIN="$(go env GOPATH)/bin"
echo "==> Installing Go tooling to $GOBIN"
mkdir -p "$GOBIN"

echo "==> gofumpt"
go install mvdan.cc/gofumpt@latest

echo "==> goimports"
go install golang.org/x/tools/cmd/goimports@latest

echo "==> govulncheck"
go install golang.org/x/vuln/cmd/govulncheck@latest

echo "==> golangci-lint v2.14.0"
curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b "$GOBIN" v2.14.0

# Verify by absolute path — never trusts $PATH lookup.
echo
echo "==> Installed:"
for tool in gofumpt goimports govulncheck golangci-lint; do
    path="$GOBIN/$tool"
    if [ -x "$path" ]; then
        case "$tool" in
            goimports)
                printf '  ✓ %-16s %s (no version flag)\n' "$tool" "$path"
                ;;
            *)
                v=$("$path" --version 2>&1 | head -1)
                printf '  ✓ %-16s %s\n' "$tool" "$v"
                ;;
        esac
    else
        printf '  ✗ %-16s NOT FOUND at %s\n' "$tool" "$path"
    fi
done

# Shadow detection — tell the user if their interactive shell will find a
# *different* golangci-lint first. Scripts are safe either way (they use
# $GOBIN prepended in _env.sh), but a mismatched `golangci-lint` in another
# terminal could still produce different results.
echo
on_path="$(command -v golangci-lint 2>/dev/null || true)"
if [ -n "$on_path" ] && [ "$on_path" != "$GOBIN/golangci-lint" ]; then
    echo "!! Your interactive shell will resolve 'golangci-lint' to:"
    echo "     $on_path"
    echo "   Scripts here use $GOBIN/golangci-lint (v2.14.0) instead."
    echo
    echo "   To make your terminal agree, add this to ~/.zshrc:"
    echo "     export PATH=\"$GOBIN:\$PATH\"   # FIRST, not appended"
    echo "   Or remove the other copy:"
    echo "     rm $on_path"
fi
