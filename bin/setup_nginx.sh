#!/usr/bin/env bash
# Installs nginx site configs for Harady and issues Let's Encrypt certificates.
#
# Usage:
#   sudo bash bin/setup_nginx.sh              # install HTTP configs, run certbot
#   sudo bash bin/setup_nginx.sh --http-only  # install HTTP configs only
set -euo pipefail

FRONT_DOMAIN="${FRONT_DOMAIN:-harady.juljan.by}"
API_DOMAIN="${API_DOMAIN:-api.harady.juljan.by}"
BACKEND_PORT="${BACKEND_PORT:-8070}"
REPO_DIR="${REPO_DIR:-/home/harady/harady.juljan.by}"
FRONT_ROOT="${FRONT_ROOT:-$REPO_DIR/frontend/dist}"
CERTBOT_EMAIL="${CERTBOT_EMAIL:-y.misiukevich@gmail.com}"

HTTP_ONLY=0
for arg in "$@"; do
    case "$arg" in
        --http-only) HTTP_ONLY=1 ;;
        *) echo "unknown arg: $arg" >&2; exit 1 ;;
    esac
done

if [ "$(id -u)" -ne 0 ]; then
    echo "error: run with sudo" >&2
    exit 1
fi

if ! command -v nginx >/dev/null 2>&1; then
    echo "==> nginx not installed, installing…"
    apt-get update && apt-get install -y nginx
fi

# ---------------------------------------------------------------------------
# Bootstrap HTTP-only configs first. Certbot will upgrade them to HTTPS.
# ---------------------------------------------------------------------------

echo "==> Writing /etc/nginx/sites-available/harady-front"
cat > /etc/nginx/sites-available/harady-front <<NGINX
server {
    listen 80;
    listen [::]:80;
    server_name $FRONT_DOMAIN;

    root $FRONT_ROOT;
    index index.html;

    # SPA fallback.
    location / {
        try_files \$uri \$uri/ /index.html;
    }
}
NGINX

echo "==> Writing /etc/nginx/sites-available/harady-api"
cat > /etc/nginx/sites-available/harady-api <<NGINX
server {
    listen 80;
    listen [::]:80;
    server_name $API_DOMAIN;

    location / {
        proxy_pass http://127.0.0.1:$BACKEND_PORT;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
NGINX

ln -sf /etc/nginx/sites-available/harady-front /etc/nginx/sites-enabled/harady-front
ln -sf /etc/nginx/sites-available/harady-api   /etc/nginx/sites-enabled/harady-api

# Drop the default welcome page if it's still enabled — it listens on :80 with
# no server_name, which would shadow our vhosts.
if [ -e /etc/nginx/sites-enabled/default ]; then
    echo "==> Disabling default nginx site"
    rm -f /etc/nginx/sites-enabled/default
fi

echo "==> nginx -t"
nginx -t

echo "==> reloading nginx"
systemctl reload nginx

if [ "$HTTP_ONLY" -eq 1 ]; then
    echo
    echo "HTTP-only configs installed. To get certs later:"
    echo "  sudo certbot --nginx -d $FRONT_DOMAIN -d $API_DOMAIN"
    exit 0
fi

# ---------------------------------------------------------------------------
# Certbot.
# ---------------------------------------------------------------------------

if ! command -v certbot >/dev/null 2>&1; then
    echo "==> Installing certbot"
    apt-get update
    apt-get install -y certbot python3-certbot-nginx
fi

echo "==> Requesting Let's Encrypt certificates"
echo "    front: $FRONT_DOMAIN"
echo "    api:   $API_DOMAIN"

CERTBOT_ARGS=(--nginx -d "$FRONT_DOMAIN" -d "$API_DOMAIN" --non-interactive --agree-tos --redirect)
if [ -n "$CERTBOT_EMAIL" ]; then
    CERTBOT_ARGS+=(--email "$CERTBOT_EMAIL")
else
    CERTBOT_ARGS+=(--register-unsafely-without-email)
fi

certbot "${CERTBOT_ARGS[@]}"

# ---------------------------------------------------------------------------
# Post-certbot: overwrite the certbot-generated configs with richer ones that
# include WebSocket support, caching rules, gzip, security headers. Certbot
# only writes the minimum; we want the polished setup.
# ---------------------------------------------------------------------------

echo "==> Installing final (HTTPS) frontend config"
cat > /etc/nginx/sites-available/harady-front <<NGINX
server {
    listen 80;
    listen [::]:80;
    server_name $FRONT_DOMAIN;
    return 301 https://\$host\$request_uri;
}

server {
    listen 443 ssl;
    listen [::]:443 ssl;
    http2 on;
    server_name $FRONT_DOMAIN;

    ssl_certificate     /etc/letsencrypt/live/$FRONT_DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$FRONT_DOMAIN/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    root $FRONT_ROOT;
    index index.html;

    # Security headers (safe defaults; add CSP/HSTS via Cloudflare if needed).
    add_header X-Content-Type-Options nosniff always;
    add_header X-Frame-Options SAMEORIGIN always;
    add_header Referrer-Policy strict-origin-when-cross-origin always;

    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types
        text/plain
        text/css
        text/javascript
        application/javascript
        application/json
        application/xml
        image/svg+xml;

    # Never cache the SPA shell or the service worker — PWA correctness
    # depends on the browser always revalidating them.
    location = /index.html {
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        try_files \$uri =404;
    }
    location = /sw.js {
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        try_files \$uri =404;
    }
    location = /workbox-*.js {
        add_header Cache-Control "no-cache, no-store, must-revalidate";
        try_files \$uri =404;
    }
    location = /manifest.webmanifest {
        add_header Cache-Control "no-cache";
        try_files \$uri =404;
    }
    location = /registerSW.js {
        add_header Cache-Control "no-cache";
        try_files \$uri =404;
    }

    # Vite hashes filenames under /assets/, so cache them forever.
    location /assets/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;
        try_files \$uri =404;
    }

    location / {
        try_files \$uri \$uri/ /index.html;
    }
}
NGINX

echo "==> Installing final (HTTPS) api config"
cat > /etc/nginx/sites-available/harady-api <<NGINX
server {
    listen 80;
    listen [::]:80;
    server_name $API_DOMAIN;
    return 301 https://\$host\$request_uri;
}

server {
    listen 443 ssl;
    listen [::]:443 ssl;
    http2 on;
    server_name $API_DOMAIN;

    ssl_certificate     /etc/letsencrypt/live/$FRONT_DOMAIN/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/$FRONT_DOMAIN/privkey.pem;
    include /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem;

    # Reasonable caps for a text game.
    client_max_body_size 64k;

    # ---- WebSocket endpoint ----
    # Long timeouts and no buffering; the game keeps the socket open for the
    # duration of a match. The Go side sends a ping every 50s.
    location /ws/ {
        proxy_pass http://127.0.0.1:$BACKEND_PORT;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
        proxy_buffering off;
    }

    # ---- REST API ----
    location / {
        proxy_pass http://127.0.0.1:$BACKEND_PORT;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;

        proxy_read_timeout 60s;
    }
}
NGINX

echo "==> nginx -t"
nginx -t

echo "==> reloading nginx"
systemctl reload nginx

echo
echo "==> Certificates will auto-renew via the certbot systemd timer."
echo "    Check: systemctl list-timers | grep certbot"
echo
echo "==> Done."
echo "    Front: https://$FRONT_DOMAIN"
echo "    API:   https://$API_DOMAIN"
