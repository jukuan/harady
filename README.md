# Harady

A multiplayer word-chain game. Players take turns naming a city that starts
with the last letter of the previous one. Two misses and you're out; last
player standing wins.

Built as a small PWA: guest-only, no accounts, one shared room code.

## Stack

- **Backend:** Go + Gin + SQLite (`modernc.org/sqlite`, no cgo)
- **Realtime:** WebSocket (`gorilla/websocket`)
- **Frontend:** React 18 + TypeScript + Vite + Tailwind
- **PWA:** `vite-plugin-pwa` (workbox)
- **Monorepo:** `backend/` and `frontend/` side by side

## Quick start

### Backend

```bash
cd backend
cp .env.example .env
go mod tidy
go run ./cmd/cli seed      # load the starter city pack
go run ./cmd/server        # listens on :8101
```

## Frontend
```
cd frontend
npm ci
npm run dev                # http://localhost:5173
```

Open http://localhost:5173, create a room, add a bot, start.
How it works

    A room holds 2–10 players (humans and bots).

    The server enforces the chain: each submission must start with the required
    letter and not repeat a used city.

    ь, ъ, ы skip to the previous letter. ў → у and й → і.

    Typos are accepted within a small edit distance (first and last letters must
    match). Минск, Менск, Мінск all resolve to the canonical Мінск.

    If the same unknown city is typed three times in a row, it's added to the
    DB with source='learned' — the game self-heals.

    Bots act on a short delay; they never pass on the opening move.

    A turn left idle for HARADY_TURN_TIMEOUT (default 120s) auto-passes.

City pack

Cities live in backend/internal/store/cities*.txt (one file per volume),
embedded at build time. Format:

```
Name|Region
````
Add a new citiesNN.txt (zero-padded) and rebuild — no code changes needed.


## CLI

```
cd backend
go run ./cmd/cli list                    # all cities
go run ./cmd/cli add --name "Мінск" --region "Мінская вобласць"
go run ./cmd/cli seed                    # add missing from pack
go run ./cmd/cli reseed                  # wipe + reload from pack
go run ./cmd/cli stats summary           # submissions + games overview
go run ./cmd/cli stats top 20            # most-played cities
go run ./cmd/cli stats games 10          # longest chains ever
````

## Configuration

.env in backend/:
Var	Default	Purpose
HARADY_ADDR	:8080	HTTP listen address
HARADY_DB_PATH	./data/harady.db	SQLite file
HARADY_LOG_PATH	./data/harady.log	JSON log file (rotated)
HARADY_LOG_LEVEL	info	debug / info / warn / error
HARADY_CORS_ORIGINS	http://localhost:5173	Allowed frontend origins
HARADY_TURN_TIMEOUT	120s	Auto-pass after this idle time
HARADY_BOT_DELAY_MS	600	Bot thinking delay (base, ms)
HARADY_ADMIN_TOKEN	(empty)	Enables /api/admin/stats/*; empty = disabled

Frontend build-time env (frontend/.env.production):

```
VITE_API_BASE=https://api.example.com
VITE_WS_BASE=wss://api.example.com
VITE_SENTRY_DSN=
VITE_APP_VERSION=
```


## Tests & quality

Backend:
```bash

cd backend
./bin/check.sh       # fmt-check + vet + lint + vuln + tests (race)
```
Frontend:
```bash

cd frontend
./bin/check.sh       # prettier-check + eslint + tsc + vite build

bin/install-tools.sh (backend) installs gofumpt, goimports,
golangci-lint, and govulncheck on first use.
````

## Deploy

Everything is scripted for a Debian/Ubuntu VPS:
```
bash bin/deploy.sh install         # build + install systemd unit + restart
bash bin/deploy.sh seed            # load cities into the live DB
bash bin/deploy.sh logs            # journalctl -f
```

nginx + Let's Encrypt:
```bash
sudo bash bin/setup_nginx.sh       # installs vhosts, obtains certs
````

Services start automatically after reboot (systemctl enable).


## Layout

```
bin/                    deploy.sh, setup_nginx.sh
backend/
  cmd/server            HTTP + WebSocket server
  cmd/cli               admin CLI (cities, seed, stats)
  internal/db           schema migrations
  internal/store        city store, analytics, games
  internal/game         room, chain rules, bots
  internal/httpapi      REST routes
  internal/ws           WebSocket client
  internal/store/cities*.txt
frontend/
  src/components        game UI
  src/pages             Landing, Room
  src/store.ts          zustand store
  src/socket.ts         WebSocket with reconnect
  src/strings.ts        Belarusian UI strings
```

## MIT License

Copyright (c) 2026 Juljan.By

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
