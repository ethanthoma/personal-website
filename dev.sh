#!/usr/bin/env bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PORT="${WEBSERVER_PORT:-8080}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAIN_BIN="${SCRIPT_DIR}/tmp/main"
PROXY_PATTERN="templ generate --watch --proxy=http://localhost:${PORT}"

# Free PORT by killing whatever is bound to it — covers orphans that
# survived a prior abrupt exit (closed terminal, kill -9 of dev.sh, etc.)
kill_port() {
    local pids
    pids=$(ss -ltnpH "sport = :${PORT}" 2>/dev/null | grep -oP 'pid=\K[0-9]+' | sort -u)
    if [[ -n "$pids" ]]; then
        echo "$pids" | xargs -r kill -9 2>/dev/null || true
    fi
}

kill_stragglers() {
    pkill -9 -f "${MAIN_BIN}" 2>/dev/null || true
    pkill -9 -f "${PROXY_PATTERN}" 2>/dev/null || true
    # Orphaned air watchers from a previous abrupt exit of this script
    pkill -9 -f "air.*${MAIN_BIN}" 2>/dev/null || true
    pkill -9 -f "air.*services/webserver/public" 2>/dev/null || true
    kill_port
}

kill_stragglers
sleep 0.3

cleanup() {
    echo -e "\n${YELLOW}Shutting down...${NC}"
    jobs -p | xargs -r kill 2>/dev/null || true
    sleep 0.3
    kill_stragglers
    echo -e "${GREEN}Done${NC}"
}

trap cleanup EXIT

export DEV=1
mkdir -p tmp

echo -e "${GREEN}Starting dev server...${NC}\n"

# Sync static assets once up front so the server has them on first run
rsync -a ./services/webserver/public/ ./public --exclude='*.css'

# 1. templ: watch templates, proxy to Go server
templ generate \
    --watch \
    --proxy="http://localhost:${PORT}" \
    --open-browser=false \
    --path=./services/webserver &
sleep 1

# 2. air: rebuild Go server on .go changes
DEV=1 air \
    --build.cmd "go build -o ./tmp/main ./services/webserver" \
    --build.bin "./tmp/main" \
    --build.include_ext "go" \
    --build.exclude_dir "tmp,static,public,docs,node_modules" \
    --build.send_interrupt true \
    --build.kill_delay 1s \
    &

# 3. tailwind: watch CSS
tailwindcss \
    -i ./services/webserver/public/main.css \
    -o ./public/main.css \
    --watch=always \
    --minify &

# 4. air: watch static assets, sync to ./public/ and notify templ proxy to reload.
#    Use --build.entrypoint, not --build.bin: the latter is silently ignored
#    in favour of its default `./tmp/main`, causing this watcher to race the
#    Go watcher for :8080 on every static-asset change.
air \
    --build.cmd "rsync -a ./services/webserver/public/ ./public --exclude='*.css' && curl -sf http://localhost:7331/_templ/reload > /dev/null 2>&1 || true" \
    --build.entrypoint "$(which true)" \
    --build.include_dir "services/webserver/public" \
    --build.include_ext "js,css,svg,png,jpg,ico,woff,woff2,ttf" \
    --build.exclude_dir "tmp" \
    -log.silent "true" \
    &

echo -e "\n${GREEN}✓ Dev server: ${NC}http://localhost:7331\n"

wait
