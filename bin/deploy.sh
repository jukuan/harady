#!/usr/bin/env bash
# Harady deployment helper. Builds binaries and (optionally) installs a
# systemd unit on this VPS.
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

cmd="${1:-build}"

# ---------------------------------------------------------------------------
# Preflight
# ---------------------------------------------------------------------------
preflight() {
    command -v go >/dev/null || { echo "error: go not on PATH" >&2; exit 1; }
    command -v sudo >/dev/null || { echo "error: sudo not available" >&2; exit 1; }
    [ -d "$BACKEND" ] || { echo "error: $BACKEND not found" >&2; exit 1; }
}

# Run the local quality gate before shipping. Can be skipped with SKIP_CHECKS=1.
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

# ---------------------------------------------------------------------------
# Build
# ---------------------------------------------------------------------------
build() {
    preflight
    test_gate
    echo "==> Building binaries into $BIN"
    mkdir -p "$BIN"
    (cd "$BACKEND" && go build -trimpath -ldflags="-s -w" -o "$BIN/harady-server" ./cmd/server)
    (cd "$BACKEND" && go build -trimpath -ldflags="-s -w" -o "$BIN/harady-cli"    ./cmd/cli)
    ls -lh "$BIN"/harady-*
}

# ---------------------------------------------------------------------------
# Install
# ---------------------------------------------------------------------------
write_env_if_missing() {
    if [ -f "$INSTALL_DIR/.env" ]; then
        echo "==> Existing .env preserved at $INSTALL_DIR/.env"
        # Warn about missing admin config without overwriting.
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
# HTTP listen address
HARADY_ADDR=$LISTEN

# SQLite file
HARADY_DB_PATH=$DB_PATH

# CORS origin (the frontend URL)
HARADY_CORS_ORIGINS=$CORS_ORIGIN

# Language for the city pack
HARADY_LANGUAGE=$LANG_TAG

# Idle room GC
HARADY_ROOM_IDLE_TTL=2h

# Logging (JSON to stdout; systemd → journald)
HARADY_LOG_LEVEL=info

# Turn timer — auto-pass after N of inactivity
HARADY_TURN_TIMEOUT=120s

# Bot thinking delay (base ms; actual = base + rand[0..base])
HARADY_BOT_DELAY_MS=600

# Admin panel — comma-separated allow-list of admin emails.
# Leave empty to disable admin panel login entirely.
HARADY_ADMIN_EMAILS=

# Admin panel base URL (frontend). Leave empty to infer from request / CORS.
HARADY_ADMIN_BASE_URL=

# API base URL for magic-link generation. Leave empty to infer from request.
HARADY_API_BASE_URL=

# Legacy token for scripted access to /api/admin/* (optional).
HARADY_ADMIN_TOKEN=

# SMTP for magic-link delivery. Leave HARADY_SMTP_HOST empty to log the link
# instead of emailing it. Uncomment and fill for production email.
# HARADY_SMTP_HOST=smtp.gmail.com
# HARADY_SMTP_PORT=587
# HARADY_SMTP_USER=you@gmail.com
# HARADY_SMTP_PASS=xxxx xxxx xxxx xxxx
# HARADY_SMTP_FROM=Harady <you@gmail.com>
ENVEOF
    sudo chown "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR/.env"
    echo
    echo "  ⚠  Edit $INSTALL_DIR/.env to set at least:"
    echo "     HARADY_ADMIN_EMAILS (required for admin panel)"
    echo "     HARADY_SMTP_*      (required for email delivery)"
    echo
}

install_systemd() {
    if [ ! -x "$BIN/harady-server" ] || [ ! -x "$BIN/harady-cli" ]; then
        echo "error: binaries not built — run 'bin/deploy.sh build' first" >&2
        exit 1
    fi

    echo "==> Installing to $INSTALL_DIR (requires sudo)"
    sudo mkdir -p "$INSTALL_DIR/bin" "$INSTALL_DIR/data"

    # CRITICAL: stop the service before overwriting the binary.
    # The running process holds harady-server open; 'sudo cp' would fail with
    # "Text file busy" otherwise. 'stop' on a non-running service is a no-op.
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
}

health_check() {
    echo
    echo "==> Health check (waiting up to 10s for :${PORT}/api/health)"
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
        echo "    Recent logs:" >&2
        sudo journalctl -u "${APP_NAME}" -n 30 --no-pager >&2 || true
        exit 1
    fi
}

# ---------------------------------------------------------------------------
# Seed / reseed
# ---------------------------------------------------------------------------
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
        echo "⚠  reseed WIPES the entire cities table, including learned entries."
        echo "   All in-progress games will dead-end on the next turn."
        read -r -p "Type 'yes' to proceed: " ans
        [ "$ans" = "yes" ] || { echo "aborted"; exit 1; }
    fi
    echo "==> Reseeding cities (wipes the table) in $INSTALL_DIR"
    sudo -u "$SERVICE_USER" -H bash -c "cd '$INSTALL_DIR' && ./bin/harady-cli reseed"
}

# ---------------------------------------------------------------------------
# Usage
# ---------------------------------------------------------------------------
usage() {
    cat <<USAGE
Usage: bin/deploy.sh <command>

Commands:
  build              build harady-server + harady-cli into ./bin
                     (runs tests first unless SKIP_CHECKS=1)
  install            build + install binaries, .env, systemd unit;
                     stops the service, copies files, restarts, health-checks
  seed               add missing cities to the installed DB (idempotent)
  reseed             wipe the cities table and re-seed from embedded pack
                     (prompts unless CONFIRM=1)
  restart            restart the systemd service
  stop               stop the systemd service
  status             show service status
  logs               follow service logs (Ctrl+C to stop)
  logs-snapshot      last 50 log lines, no follow
  health             curl the /api/health endpoint
  help

Env overrides:
  APP_NAME           systemd unit name (default: harady)
  INSTALL_DIR        install path     (default: /opt/harady)
  SERVICE_USER       service account  (default: harady)
  HARADY_ADDR        listen address   (default: :8101)
  HARADY_DB_PATH     SQLite path      (default: \$INSTALL_DIR/data/harady.db)
  HARADY_CORS_ORIGINS allowed origin  (default: https://harady.juljan.by)
  HARADY_LANGUAGE    city-pack lang   (default: be)
  SKIP_CHECKS        1 = skip test gate
  CONFIRM            1 = skip reseed prompt
USAGE
}

case "$cmd" in
    build)         build ;;
    install)       build; install_systemd ;;
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
