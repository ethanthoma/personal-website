#!/usr/bin/env bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAIN_BIN="${SCRIPT_DIR}/tmp/main"

port_is_free() {
    [[ -z "$(ss -ltnH "sport = :$1" 2>/dev/null)" ]]
}

find_free_port() {
    local port=$1
    local i
    for ((i = 0; i < 100; i++)); do
        if port_is_free "$port"; then
            echo "$port"
            return 0
        fi
        port=$((port + 1))
    done
    echo "dev.sh: no free port in ${1}-$((port - 1))" >&2
    return 1
}

kill_stragglers() {
    local pid
    for pid in $(pgrep -f 'templ generate --watch|tailwindcss.*--watch|air --build|tmp/main' || true); do
        if [[ "$(readlink "/proc/${pid}/cwd" 2>/dev/null)" == "${SCRIPT_DIR}" ]]; then
            kill -9 "$pid" 2>/dev/null || true
        fi
    done
}

kill_stragglers
sleep 0.3

PORT="$(find_free_port "${WEBSERVER_PORT:-8080}")"
PROXY_PORT="$(find_free_port "${TEMPL_PROXY_PORT:-7331}")"
export WEBSERVER_PORT="${PORT}"

cleanup() {
    echo -e "\n${YELLOW}Shutting down...${NC}"
    jobs -p | xargs -r kill 2>/dev/null || true
    sleep 0.3
    kill_stragglers
    echo -e "${GREEN}Done${NC}"
}

trap cleanup EXIT

export DEV=1
mkdir -p tmp public/js

DATASTAR_VERSION="1.0.1"
DATASTAR_CACHE="${SCRIPT_DIR}/tmp/datastar-v${DATASTAR_VERSION}.js"
DATASTAR_TARGET="${SCRIPT_DIR}/public/js/datastar.js"
if [[ ! -f "$DATASTAR_CACHE" ]]; then
    echo -e "${GREEN}Fetching datastar v${DATASTAR_VERSION}...${NC}"
    curl -sfL "https://raw.githubusercontent.com/starfederation/datastar/v${DATASTAR_VERSION}/bundles/datastar.js" -o "$DATASTAR_CACHE"
fi
cp "$DATASTAR_CACHE" "$DATASTAR_TARGET"

echo -e "${GREEN}Starting dev server...${NC}\n"

rsync -a ./services/webserver/public/ ./public --exclude='*.css'

# 1. templ: watch templates, proxy to Go server
templ generate \
    --watch \
    --proxy="http://localhost:${PORT}" \
    --proxyport="${PROXY_PORT}" \
    --open-browser=false \
    --path=./services/webserver &
sleep 1

# 2. air: rebuild Go server on .go changes, and on the go:embed'd layouts/*.js
DEV=1 air \
    --build.cmd "go build -o ./tmp/main ./services/webserver" \
    --build.bin "./tmp/main" \
    --build.include_ext "go,js" \
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
    --build.cmd "rsync -a ./services/webserver/public/ ./public --exclude='*.css' && curl -sf http://localhost:${PROXY_PORT}/_templ/reload > /dev/null 2>&1 || true" \
    --build.entrypoint "$(which true)" \
    --build.include_dir "services/webserver/public" \
    --build.include_ext "js,css,svg,png,jpg,ico,woff,woff2,ttf,json" \
    --build.exclude_dir "tmp" \
    -log.silent "true" \
    &

echo -e "\n${GREEN}✓ Dev server: ${NC}http://localhost:${PROXY_PORT}\n"

wait
