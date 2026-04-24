#!/usr/bin/env bash

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PORT="${WEBSERVER_PORT:-8080}"

# Kill any leftover processes
pkill -f "tmp/main" 2>/dev/null || true
pkill -f "templ generate --watch" 2>/dev/null || true
sleep 0.5

cleanup() {
    echo -e "\n${YELLOW}Shutting down...${NC}"
    jobs -p | xargs -r kill 2>/dev/null || true
    pkill -f "tmp/main" 2>/dev/null || true
    sleep 1
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
sleep 1  # Wait for proxy to start

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

# 4. air: watch static assets, sync to ./public/ and notify templ proxy to reload
air \
    --build.cmd "rsync -a ./services/webserver/public/ ./public --exclude='*.css' && curl -s http://localhost:7331/_templ/reload > /dev/null 2>&1" \
    --build.bin "/run/current-system/sw/bin/true" \
    --build.include_dir "services/webserver/public" \
    --build.include_ext "js,css,svg,png,jpg,ico,woff,woff2,ttf" \
    --build.exclude_dir "tmp" \
    -log.silent "true" \
    &

echo -e "\n${GREEN}✓ Dev server: ${NC}http://localhost:7331\n"

wait
