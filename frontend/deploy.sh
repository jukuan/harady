#!/usr/bin/env bash
# Frontend deployment: build Vite bundle, fix permissions for nginx, verify.
# Works standalone or via bin/deploy.sh.
#
# Usage:
#   bash frontend/deploy.sh            # build + chmod + health (default)
#   bash frontend/deploy.sh build      # npm ci + npm run build
#   bash frontend/deploy.sh chmod      # fix permissions (assumes build done)
#   bash frontend/deploy.sh health     # curl via Host: header
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FRONTEND="$ROOT/frontend"
DIST="${FRONTEND_DIST:-$FRONTEND/dist}"
CORS_ORIGIN="${HARADY_CORS_ORIGINS:-https://harady.juljan.by}"
HOST="$(echo "$CORS_ORIGIN" | sed -E 's|^https?://||; s|/.*$||')"

cmd="${1:-deploy}"

# ─────────────────────────────────────────────────────────────────────────────
# Node discovery (nvm fallback)
# ─────────────────────────────────────────────────────────────────────────────

ensure_node() {
    if command -v node >/dev/null 2>&1 && command -v npm >/dev/null 2>&1; then
        return 0
    fi
    for f in "$HOME/.nvm/nvm.sh" "/usr/local/nvm/nvm.sh" "/usr/local/opt/nvm/nvm.sh"; do
        if [ -s "$f" ]; then
            # shellcheck disable=SC1090
            source "$f" >/dev/null 2>&1 || true
            break
        fi
    done
    if ! command -v node >/dev/null 2>&1; then
        echo "error: node/npm not found" >&2
        echo "       Install Node.js 20+ or source nvm first:" >&2
        echo "         source ~/.nvm/nvm.sh" >&2
        exit 1
    fi
    echo "==> Node $(node --version), npm $(npm --version)"
}

# ─────────────────────────────────────────────────────────────────────────────
# Build
# ─────────────────────────────────────────────────────────────────────────────

build() {
    ensure_node
    [ -d "$FRONTEND" ] || { echo "error: $FRONTEND not found" >&2; exit 1; }
    echo "==> Building frontend in $FRONTEND"
    (
        cd "$FRONTEND"
        if [ -f package-lock.json ]; then
            echo "    npm ci"
            npm ci --no-audit --no-fund
        else
            echo "    npm install (no package-lock.json)"
            npm install --no-audit --no-fund
        fi
        echo "    npm run build"
        npm run build
    )
    if [ ! -f "$DIST/index.html" ]; then
        echo "error: $DIST/index.html not produced — build failed?" >&2
        exit 1
    fi
    echo "    build output:"
    ls -lh "$DIST" | head -10
}

# ─────────────────────────────────────────────────────────────────────────────
# Permissions
# ─────────────────────────────────────────────────────────────────────────────

chmod_web() {
    if [ ! -d "$DIST" ]; then
        echo "error: $DIST not found — run 'build' first" >&2
        exit 1
    fi
    echo "==> Fixing permissions for nginx"
    sudo chmod -R o+rX "$DIST"

    # Walk up the tree adding o+x on each parent that isn't already traversable.
    local d perms
    d="$(dirname "$DIST")"
    while [ "$d" != "/" ] && [ -n "$d" ]; do
        perms="$(stat -c '%A' "$d" 2>/dev/null || stat -f '%Sp' "$d" 2>/dev/null || echo '')"
        case "$perms" in
            *x*x*x*) ;;
            '') ;;
            *) echo "    chmod o+x $d"; sudo chmod o+x "$d" 2>/dev/null || true ;;
        esac
        d="$(dirname "$d")"
    done

    # Verify nginx (www-data) can actually read the file.
    if command -v sudo >/dev/null 2>&1 && id www-data >/dev/null 2>&1; then
        if sudo -u www-data test -r "$DIST/index.html"; then
            echo "    ✓ www-data can read $DIST/index.html"
        else
            echo "    ✗ www-data cannot read $DIST/index.html" >&2
            echo "      Diagnose: namei -l $DIST/index.html" >&2
        fi
    fi
}

# ─────────────────────────────────────────────────────────────────────────────
# Health check (via nginx with Host header)
# ─────────────────────────────────────────────────────────────────────────────

health() {
    [ -n "$HOST" ] || { echo "    (no CORS origin — skipping health)"; return 0; }
    echo "==> Frontend health check (Host: $HOST)"
    local code ok=0
    for i in 1 2 3; do
        code="$(curl -sk -o /dev/null -w '%{http_code}' -H "Host: $HOST" http://127.0.0.1/ 2>/dev/null || true)"
        case "$code" in
            200|301|308) ok=1; echo "    ✓ $HOST → $code"; break ;;
        esac
        sleep 1
    done
    if [ "$ok" != "1" ]; then
        echo "    ✗ $HOST → $code" >&2
        echo "      Try: sudo tail -30 /var/log/nginx/error.log" >&2
    fi
}

# ─────────────────────────────────────────────────────────────────────────────
# Dispatch
# ─────────────────────────────────────────────────────────────────────────────

usage() {
    cat <<USAGE
Usage: frontend/deploy.sh <command>

Commands:
  build              npm ci + npm run build into frontend/dist/
  chmod              fix permissions so nginx can read frontend/dist/
  health             curl via Host: header
  deploy             build + chmod + health (default)

Env:
  FRONTEND_DIST      build output dir (default: frontend/dist)
  HARADY_CORS_ORIGINS  used to derive Host: for health check
USAGE
}

case "$cmd" in
    build)         build ;;
    chmod)         chmod_web ;;
    health)        health ;;
    deploy|all)    build; chmod_web; health ;;
    help|*)        usage ;;
esac
