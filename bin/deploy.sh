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
LISTEN="${HARADY_ADDR:-:8070}"
DB_PATH="${HARADY_DB_PATH:-$INSTALL_DIR/data/harady.db}"
LANG_TAG="${HARADY_LANGUAGE:-be}"
CORS_ORIGIN="${HARADY_CORS_ORIGINS:-https://harady.juljan.by}"

cmd="${1:-build}"

build() {
    echo "==> Building binaries into $BIN"
    mkdir -p "$BIN"
    (cd "$BACKEND" && go build -trimpath -ldflags="-s -w" -o "$BIN/harady-server" ./cmd/server)
    (cd "$BACKEND" && go build -trimpath -ldflags="-s -w" -o "$BIN/harady-cli"    ./cmd/cli)
    ls -lh "$BIN"/harady-*
}

install_systemd() {
    if [ ! -x "$BIN/harady-server" ] || [ ! -x "$BIN/harady-cli" ]; then
        echo "error: binaries not built — run 'bin/deploy.sh build' first" >&2
        exit 1
    fi

    echo "==> Installing to $INSTALL_DIR (requires sudo)"
    sudo mkdir -p "$INSTALL_DIR/bin" "$INSTALL_DIR/data"
    sudo cp "$BIN/harady-server" "$INSTALL_DIR/bin/"
    sudo cp "$BIN/harady-cli"    "$INSTALL_DIR/bin/"
    sudo chmod +x "$INSTALL_DIR/bin/"*

    if ! id -u "$SERVICE_USER" >/dev/null 2>&1; then
        sudo useradd --system --home "$INSTALL_DIR" --shell /usr/sbin/nologin "$SERVICE_USER"
    fi
    sudo chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"

    if [ ! -f "$INSTALL_DIR/.env" ]; then
        sudo tee "$INSTALL_DIR/.env" >/dev/null <<ENVEOF
HARADY_ADDR=$LISTEN
HARADY_DB_PATH=$DB_PATH
HARADY_CORS_ORIGINS=$CORS_ORIGIN
HARADY_LANGUAGE=$LANG_TAG
HARADY_ROOM_IDLE_TTL=2h
ENVEOF
        sudo chown "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR/.env"
        echo
        echo "  ⚠  Wrote a starter .env at $INSTALL_DIR/.env"
        echo "     Edit HARADY_CORS_ORIGINS if you serve from a different host."
        echo
    fi

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
    # enable is idempotent; restart always picks up new binaries.
    sudo systemctl enable ${APP_NAME}
    sudo systemctl restart ${APP_NAME}
    sudo systemctl status ${APP_NAME} --no-pager || true
}

# Run the *installed* CLI as the service user, from $INSTALL_DIR, so it picks
# up the real .env and hits the real DB. Local $BIN/harady-cli is only used
# for the build step.
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
    echo "==> Reseeding cities (wipes the table) in $INSTALL_DIR"
    sudo -u "$SERVICE_USER" -H bash -c "cd '$INSTALL_DIR' && ./bin/harady-cli reseed"
}

usage() {
    cat <<USAGE
Usage: bin/deploy.sh <command>

Commands:
  build              build harady-server + harady-cli into ./bin
  install            (sudo) build if needed, then install binaries, .env,
                     and systemd unit; enable + (re)start the service
  seed               add missing cities to the installed DB (idempotent)
  reseed             wipe the cities table and re-seed from the embedded pack
  restart            (sudo) restart the systemd service
  stop               (sudo) stop the systemd service
  status             (sudo) show service status
  logs               (sudo) journalctl -u ${APP_NAME} -f
  help

Env overrides:
  APP_NAME           systemd unit name (default: harady)
  INSTALL_DIR        install path     (default: /opt/harady)
  SERVICE_USER       service account  (default: harady)
  HARADY_ADDR        listen address   (default: :8070)
  HARADY_DB_PATH     SQLite path      (default: \$INSTALL_DIR/data/harady.db)
  HARADY_CORS_ORIGINS allowed origin  (default: https://harady.juljan.by)
  HARADY_LANGUAGE    city-pack lang   (default: be)
USAGE
}

case "$cmd" in
    build)    build ;;
    install)  build; install_systemd ;;
    seed)     seed_db ;;
    reseed)   reseed_db ;;
    restart)  sudo systemctl restart ${APP_NAME} ;;
    stop)     sudo systemctl stop ${APP_NAME} ;;
    status)   sudo systemctl status ${APP_NAME} --no-pager ;;
    logs)     sudo journalctl -u ${APP_NAME} -f ;;
    help|*)   usage ;;
esac
