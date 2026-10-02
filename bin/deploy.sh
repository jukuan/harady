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
LISTEN="${HARADY_ADDR:-:8080}"
DB_PATH="${HARADY_DB_PATH:-$INSTALL_DIR/data/harady.db}"
LANG_TAG="${HARADY_LANGUAGE:-be}"

cmd="${1:-build}"

build() {
    echo "==> Building binaries into $BIN"
    mkdir -p "$BIN"
    (cd "$BACKEND" && go build -trimpath -ldflags="-s -w" -o "$BIN/harady-server" ./cmd/server)
    (cd "$BACKEND" && go build -trimpath -ldflags="-s -w" -o "$BIN/harady-cli"    ./cmd/cli)
    ls -lh "$BIN"/harady-*
}

install_systemd() {
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
        sudo tee "$INSTALL_DIR/.env" >/dev/null <<EOF
HARADY_ADDR=$LISTEN
HARADY_DB_PATH=$DB_PATH
HARADY_CORS_ORIGINS=https://harady.example.com
HARADY_LANGUAGE=$LANG_TAG
HARADY_ROOM_IDLE_TTL=2h
EOF
        sudo chown "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR/.env"
    fi

    sudo tee /etc/systemd/system/${APP_NAME}.service >/dev/null <<EOF
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
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable --now ${APP_NAME}
    sudo systemctl status ${APP_NAME} --no-pager || true
}

seed_db() {
    echo "==> Seeding cities"
    "$BIN/harady-cli" seed
}

usage() {
    cat <<EOF
Usage: bin/deploy.sh <command>

Commands:
  build              build harady-server + harady-cli into ./bin
  install            (sudo) install binaries, .env, and systemd unit
  seed               run harady-cli seed against the configured DB
  restart            (sudo) restart the systemd service
  logs               (sudo) journalctl -u ${APP_NAME} -f
  help
EOF
}

case "$cmd" in
    build)    build ;;
    install)  build; install_systemd ;;
    seed)     seed_db ;;
    restart)  sudo systemctl restart ${APP_NAME} ;;
    logs)     sudo journalctl -u ${APP_NAME} -f ;;
    help|*)   usage ;;
esac
