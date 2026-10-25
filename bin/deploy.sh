#!/usr/bin/env bash
# Full deployment orchestrator.
#
# One command deploys everything:
#   backend binary + systemd restart + migrations
#   frontend bundle + nginx-readable permissions
#   nginx reload (only if config is valid and harady vhosts exist)
#   health checks for both halves
#
# Usage:
#   bin/deploy.sh              # full deploy (default)
#   bin/deploy.sh all          # same
#   bin/deploy.sh backend      # backend only
#   bin/deploy.sh frontend     # frontend only
#   bin/deploy.sh seed         # add missing cities
#   bin/deploy.sh reseed       # wipe + re-seed (prompts)
#   bin/deploy.sh restart      # restart backend
#   bin/deploy.sh status       # show backend + nginx status
#   bin/deploy.sh logs         # follow backend logs
#   bin/deploy.sh logs-web     # follow nginx logs
#   bin/deploy.sh health       # both health checks
#   bin/deploy.sh help
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DEPLOY="$ROOT/backend/deploy.sh"
FRONTEND_DEPLOY="$ROOT/frontend/deploy.sh"

APP_NAME="${APP_NAME:-harady}"

# Ensure sub-scripts are executable (self-healing).
chmod +x "$BACKEND_DEPLOY" "$FRONTEND_DEPLOY" 2>/dev/null || true

cmd="${1:-all}"

# ─────────────────────────────────────────────────────────────────────────────
# Nginx
# ─────────────────────────────────────────────────────────────────────────────

nginx_reload() {
    echo "==> Nginx"
    if ! command -v nginx >/dev/null 2>&1; then
        echo "    (nginx not installed — skipping)"
        return 0
    fi
    if ! ls /etc/nginx/sites-enabled/harady-* >/dev/null 2>&1; then
        echo "    (no harady vhosts in /etc/nginx/sites-enabled — skipping)"
        return 0
    fi

    if ! sudo nginx -t >/tmp/nginx-t.log 2>&1; then
        echo "    ✗ nginx config is invalid — NOT reloading (old config stays active)" >&2
        sed 's/^/      /' /tmp/nginx-t.log >&2
        rm -f /tmp/nginx-t.log
        return 1
    fi
    rm -f /tmp/nginx-t.log
    echo "    config valid — reloading"
    sudo systemctl reload nginx
    echo "    ✓ nginx reloaded"
}

# ─────────────────────────────────────────────────────────────────────────────
# Full deploy
# ─────────────────────────────────────────────────────────────────────────────

deploy_all() {
    echo "═══════════════════════════════════════════════════════════════════"
    echo "  Harady — full deployment"
    echo "═══════════════════════════════════════════════════════════════════"
    echo

    git pull

    # 1. Build everything first. If any build fails, the site stays up
    #    on the current version — nothing is touched.
    echo "── Building backend ─────────────────────────────────────────────"
    bash "$BACKEND_DEPLOY" build

    echo
    echo "── Building frontend ────────────────────────────────────────────"
    bash "$FRONTEND_DEPLOY" build

    # 2. Install backend. This stops the service, copies the new binary,
    #    restarts (which runs migrations via db.Open), and health-checks.
    echo
    echo "── Installing backend ───────────────────────────────────────────"
    bash "$BACKEND_DEPLOY" install

    # 3. Fix frontend permissions so nginx can serve the new bundle.
    echo
    echo "── Fixing frontend permissions ──────────────────────────────────"
    bash "$FRONTEND_DEPLOY" chmod

    # 4. Reload nginx (no-op if configs aren't ours / invalid).
    echo
    nginx_reload

    # 5. Final health checks.
    echo
    echo "── Final health checks ──────────────────────────────────────────"
    bash "$BACKEND_DEPLOY" health
    bash "$FRONTEND_DEPLOY" health

    echo
    echo "═══════════════════════════════════════════════════════════════════"
    echo "  ✓ Deployed"
    echo "═══════════════════════════════════════════════════════════════════"
}

# ─────────────────────────────────────────────────────────────────────────────
# Dispatch
# ─────────────────────────────────────────────────────────────────────────────

usage() {
    cat <<USAGE
Usage: bin/deploy.sh <command>

Commands:
  all (default)      deploy backend + frontend + nginx reload + health checks
  backend            deploy backend only
  frontend           deploy frontend only
  seed               add missing cities to the installed DB
  reseed             wipe + re-seed the cities table (prompts unless CONFIRM=1)
  restart            restart the backend service
  stop               stop the backend service
  status             show backend + nginx status
  logs               follow backend logs (journalctl -u harady -f)
  logs-web           follow nginx logs
  health             run both health checks

Env overrides:
  APP_NAME, INSTALL_DIR, SERVICE_USER
  HARADY_ADDR, HARADY_DB_PATH, HARADY_CORS_ORIGINS, HARADY_LANGUAGE
  FRONTEND_DIST
  SKIP_CHECKS=1      skip backend test gate
  CONFIRM=1          skip reseed prompt

Examples:
  git pull && bin/deploy.sh           # typical full deploy
  bin/deploy.sh backend               # backend only (no npm)
  bin/deploy.sh seed                  # after adding cities*.txt
  bin/deploy.sh logs                  # tail backend logs
USAGE
}

case "$cmd" in
    all|"")        deploy_all ;;
    backend)       bash "$BACKEND_DEPLOY" all ;;
    frontend)      bash "$FRONTEND_DEPLOY" deploy ;;
    seed)          bash "$BACKEND_DEPLOY" seed ;;
    reseed)        bash "$BACKEND_DEPLOY" reseed ;;
    restart)       bash "$BACKEND_DEPLOY" restart ;;
    stop)          bash "$BACKEND_DEPLOY" stop ;;
    status)
        bash "$BACKEND_DEPLOY" status
        if command -v nginx >/dev/null 2>&1; then
            echo
            echo "── nginx ──"
            systemctl is-active nginx && systemctl status nginx --no-pager | head -5 || true
        fi
        ;;
    logs)          bash "$BACKEND_DEPLOY" logs ;;
    logs-web)      sudo tail -f /var/log/nginx/error.log /var/log/nginx/access.log ;;
    health)
        bash "$BACKEND_DEPLOY" health
        bash "$FRONTEND_DEPLOY" health
        ;;
    help|*)        usage ;;
esac
