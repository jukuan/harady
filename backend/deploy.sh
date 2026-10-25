#!/usr/bin/env bash
# Backend deployment: build binary, install systemd unit, restart, seed.
# Works standalone or via bin/deploy.sh.
#
# Usage:
#   bash backend/deploy.sh            # build + install (default)
#   bash backend/deploy.sh build      # build binary only (runs tests)
#   bash backend/deploy.sh install    # install + restart (assumes build done)
#   bash backend/deploy.sh seed       # add missing cities
#   bash backend/deploy.sh reseed     # wipe + re-seed
#   bash backend/deploy.sh restart
#   bash backend/deploy.sh stop
#   bash backend/deploy.sh status
#   bash backend/deploy.sh logs
#   bash backend/deploy.sh logs-snapshot
#   bash backend/deploy.sh health
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND="$ROOT/backend"
BIN="$ROOT/bin"

APP_NAME="${APP_NAME:-harady}"
INSTALL_DIR="${INSTALL_DIR:-/opt/harady}"
SERVICE_USER="${SERVICE_USER:-harady}"
LISTEN="${HARADY_ADDR:-:8101}"
PORT="${LISTEN#:}"
DB_PATH="${HARADY_DB_PATH:-$INSTALL_DIR/data/harady.db}"
LANG_TAG="${HARADY_LANGUAGE:-be}"
CORS_ORIGIN="${HARADY_CORS_ORIGINS:-https://harady.juljan.by}"

cmd="${1:-install}"

# ─────────────────────────────────────────────────────────────────────────────
# Build
# ─────────────────────────────────────────────────────────────────────────────

preflight() {
    command -v go >/dev/null || { echo "error: go not on PATH" >&2; exit 1; }
    command -v sudo >/dev/null || { echo "error: sudo not available" >&2; exit 1; }
    [ -d "$BACKEND" ] || { echo "error: $BACKEND not found" >&2; exit 1; }
}

test_gate() {
    if [ "${SKIP_CHECKS:-0}" = "1" ]; then
        echo "==> SKIP_CHECKS=1 — skipping test gate"
        return
    fi
    echo "==> Running backend tests before build"
    if [ -x "$BACKEND/bin/test.sh" ]; then
        (cd "$BACKEND" && ./bin/test.sh -short) || {
            echo "error: backend tests failed — refusing to build" >&2
            exit 1
        }
    else
        (cd "$BACKEND" && go test -short ./...) || {
            echo "error: backend tests failed — refusing to build" >&2
            exit 1
        }
    fi
}

build() {
    preflight
    test_gate
    echo "==> Building binaries into $BIN"
    mkdir -p "$BIN"
    (cd "$BACKEND" && go build -trimpath -ldflags="-s -w" -o "$BIN/harady-server" ./cmd/server)
    (cd "$BACKEND" && go build -trimpath -ldflags="-s -w" -o "$BIN/harady-cli"    ./cmd/cli)
    ls -lh "$BIN"/harady-*
}

# ─────────────────────────────────────────────────────────────────────────────
# Install
# ─────────────────────────────────────────────────────────────────────────────

write_env_if_missing() {
    if [ -f "$INSTALL_DIR/.env" ]; then
        echo "==> Existing .env preserved at $INSTALL_DIR/.env"
        if ! sudo grep -q '^HARADY_ADMIN_EMAILS=.' "$INSTALL_DIR/.env"; then
            echo "  ⚠  HARADY_ADMIN_EMAILS not set — admin panel will be unusable"
        fi
        if ! sudo grep -q '^HARADY_SMTP_HOST=.' "$INSTALL_DIR/.env"; then
            echo "  ℹ  HARADY_SMTP_HOST not set — magic links go to the log, not email"
        fi
        return
    fi

    echo "==> Writing starter .env at $INSTALL_DIR/.env"
    sudo tee "$INSTALL_DIR/.env" >/dev/null <<ENVEOF
HARADY_ADDR=$LISTEN
HARADY_DB_PATH=$DB_PATH
HARADY_CORS_ORIGINS=$CORS_ORIGIN
HARADY_LANGUAGE=$LANG_TAG
HARADY_ROOM_IDLE_TTL=2h
HARADY_LOG_LEVEL=info
HARADY_TURN_TIMEOUT=120s
HARADY_BOT_DELAY_MS=600
HARADY_ADMIN_EMAILS=
HARADY_ADMIN_BASE_URL=
HARADY_API_BASE_URL=
HARADY_ADMIN_TOKEN=
# HARADY_SMTP_HOST=
# HARADY_SMTP_PORT=587
# HARADY_SMTP_USER=
# HARADY_SMTP_PASS=
# HARADY_SMTP_FROM=Harady <noreply@localhost>
ENVEOF
    sudo chown "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR/.env"
    echo
    echo "  ⚠  Edit $INSTALL_DIR/.env to set:"
    echo "     HARADY_ADMIN_EMAILS  (required for admin panel)"
    echo "     HARADY_SMTP_*        (required for magic-link email)"
    echo
}

install_systemd() {
    if [ ! -x "$BIN/harady-server" ] || [ ! -x "$BIN/harady-cli" ]; then
        echo "error: binaries not built — run 'backend/deploy.sh build' first" >&2
        exit 1
    fi

    echo "==> Installing to $INSTALL_DIR (requires sudo)"
    sudo mkdir -p "$INSTALL_DIR/bin" "$INSTALL_DIR/data"

    # CRITICAL: stop service before overwriting binary. The running process
    # holds harady-server open; 'sudo cp' fails with "Text file busy".
    echo "==> Stopping ${APP_NAME} (if running)"
    sudo systemctl stop "${APP_NAME}" 2>/dev/null || true

    sudo cp "$BIN/harady-server" "$INSTALL_DIR/bin/"
    sudo cp "$BIN/harady-cli"    "$INSTALL_DIR/bin/"
    sudo chmod +x "$INSTALL_DIR/bin/"*

    if ! id -u "$SERVICE_USER" >/dev/null 2>&1; then
        sudo useradd --system --home "$INSTALL_DIR" --shell /usr/sbin/nologin "$SERVICE_USER"
    fi
    sudo chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"

    write_env_if_missing

    sudo tee /etc/systemd/system/${APP_NAME}.service >/dev/null <<UNITEOF
[Unit]
Description=Harady game server
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
EnvironmentFile=$INSTALL_DIR/.env
ExecStart=$INSTALL_DIR/bin/harady-server
Restart=on-failure
RestartSec=3

[Install]
WantedBy=multi-user.target
UNITEOF

    sudo systemctl daemon-reload
    sudo systemctl enable "${APP_NAME}"
    sudo systemctl start "${APP_NAME}"
    sudo systemctl status "${APP_NAME}" --no-pager || true

    health_check
    migrations_report
}

health_check() {
    echo
    echo "==> Backend health check (waiting up to 10s for :${PORT}/api/health)"
    local ok=0
    for i in 1 2 3 4 5 6 7 8 9 10; do
        if curl -sf "http://127.0.0.1:${PORT}/api/health" >/dev/null 2>&1; then
            ok=1
            break
        fi
        sleep 1
    done
    if [ "$ok" = "1" ]; then
        echo "    ✓ service is healthy on :${PORT}"
    else
        echo "    ✗ service did not become healthy in 10s" >&2
        sudo journalctl -u "${APP_NAME}" -n 30 --no-pager >&2 || true
        exit 1
    fi
}

migrations_report() {
    echo
    echo "==> Migration check"
    local applied
    applied="$(sudo journalctl -u "${APP_NAME}" --since "60 seconds ago" --no-pager 2>/dev/null \
        | grep -oE 'applied migration [0-9]+ \([^)]+\)' || true)"
    if [ -n "$applied" ]; then
        echo "    New migrations applied:"
        echo "$applied" | sed 's/^/      /'
    else
        echo "    No new migrations (schema up to date)"
    fi
}

# ─────────────────────────────────────────────────────────────────────────────
# Seed
# ─────────────────────────────────────────────────────────────────────────────

seed_db() {
    if [ ! -x "$INSTALL_DIR/bin/harady-cli" ]; then
        echo "error: $INSTALL_DIR/bin/harady-cli not found — run 'install' first" >&2
        exit 1
    fi
    echo "==> Seeding cities in $INSTALL_DIR"
    sudo -u "$SERVICE_USER" -H bash -c "cd '$INSTALL_DIR' && ./bin/harady-cli seed"
}

reseed_db() {
    if [ ! -x "$INSTALL_DIR/bin/harady-cli" ]; then
        echo "error: $INSTALL_DIR/bin/harady-cli not found — run 'install' first" >&2
        exit 1
    fi
    if [ "${CONFIRM:-0}" != "1" ]; then
        echo "⚠  reseed WIPES the cities table, including learned entries."
        read -r -p "Type 'yes' to proceed: " ans
        [ "$ans" = "yes" ] || { echo "aborted"; exit 1; }
    fi
    echo "==> Reseeding cities in $INSTALL_DIR"
    sudo -u "$SERVICE_USER" -H bash -c "cd '$INSTALL_DIR' && ./bin/harady-cli reseed"
}

# ─────────────────────────────────────────────────────────────────────────────
# Dispatch
# ─────────────────────────────────────────────────────────────────────────────

usage() {
    cat <<USAGE
Usage: backend/deploy.sh <command>

Commands:
  build              build harady-server + harady-cli (runs tests)
  install            build if needed, install binaries + systemd unit, restart,
                     run migrations, health check
  all                alias for 'install' (build + install) — default
  seed               add missing cities to the installed DB (idempotent)
  reseed             wipe the cities table and re-seed (prompts unless CONFIRM=1)
  restart            restart the systemd service
  stop               stop the systemd service
  status             show service status
  logs               follow backend logs
  logs-snapshot      last 50 log lines
  health             curl /api/health

Env:
  APP_NAME, INSTALL_DIR, SERVICE_USER
  HARADY_ADDR, HARADY_DB_PATH, HARADY_CORS_ORIGINS, HARADY_LANGUAGE
  SKIP_CHECKS=1      skip test gate
  CONFIRM=1          skip reseed prompt
USAGE
}

case "$cmd" in
    build)         build ;;
    install)       install_systemd ;;
    all)           build; install_systemd ;;
    seed)          seed_db ;;
    reseed)        reseed_db ;;
    restart)       sudo systemctl restart "${APP_NAME}" ;;
    stop)          sudo systemctl stop "${APP_NAME}" ;;
    status)        sudo systemctl status "${APP_NAME}" --no-pager ;;
    logs)          sudo journalctl -u "${APP_NAME}" -f ;;
    logs-snapshot) sudo journalctl -u "${APP_NAME}" -n 50 --no-pager ;;
    health)        curl -sf "http://127.0.0.1:${PORT}/api/health" && echo ;;
    help|*)        usage ;;
esac
