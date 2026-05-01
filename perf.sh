#!/usr/bin/env bash

# Prod-mode local Lighthouse audit. Builds versioned-URL static, starts the
# server with no DEV=1, runs lighthouse mobile against the given paths (or
# a sensible default set), prints scores.
#
# Usage:
#   ./perf.sh                       # /home, /resources, first post from rss
#   ./perf.sh /home /post/some-slug # specific paths

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

PORT="${PERF_PORT:-8090}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN="${SCRIPT_DIR}/tmp/perf-main"
OUT="${SCRIPT_DIR}/tmp/lighthouse"
LOG="${SCRIPT_DIR}/tmp/perf-server.log"

SERVER_PID=""

cleanup() {
    if [[ -n "$SERVER_PID" ]] && kill -0 "$SERVER_PID" 2>/dev/null; then
        kill "$SERVER_PID" 2>/dev/null || true
    fi
    local pids
    pids=$(ss -ltnpH "sport = :${PORT}" 2>/dev/null | grep -oP 'pid=\K[0-9]+' | sort -u)
    [[ -n "$pids" ]] && echo "$pids" | xargs -r kill -9 2>/dev/null || true
}
trap cleanup EXIT

mkdir -p "${SCRIPT_DIR}/tmp" "${SCRIPT_DIR}/public/js" "$OUT"

echo -e "${GREEN}» templ generate${NC}"
templ generate --path=./services/webserver

echo -e "${GREEN}» tailwind (minified)${NC}"
tailwindcss -i ./services/webserver/public/main.css -o ./public/main.css --minify

echo -e "${GREEN}» sync static${NC}"
rsync -a ./services/webserver/public/ ./public --exclude='*.css'

DATASTAR_VERSION="1.0.1"
DATASTAR_CACHE="${SCRIPT_DIR}/tmp/datastar-v${DATASTAR_VERSION}.js"
DATASTAR_TARGET="${SCRIPT_DIR}/public/js/datastar.js"
if [[ ! -f "$DATASTAR_CACHE" ]]; then
    echo -e "${GREEN}» fetching datastar v${DATASTAR_VERSION}${NC}"
    curl -sfL "https://raw.githubusercontent.com/starfederation/datastar/v${DATASTAR_VERSION}/bundles/datastar.js" -o "$DATASTAR_CACHE"
fi
cp "$DATASTAR_CACHE" "$DATASTAR_TARGET"

echo -e "${GREEN}» go build${NC}"
go build -o "$BIN" ./services/webserver

echo -e "${GREEN}» starting server on :${PORT} (prod mode)${NC}"
WEBSERVER_PORT="$PORT" "$BIN" > "$LOG" 2>&1 &
SERVER_PID=$!

for i in {1..50}; do
    if curl -sf "http://localhost:${PORT}/healthy" > /dev/null 2>&1; then
        break
    fi
    sleep 0.1
    if [[ $i -eq 50 ]]; then
        echo -e "${RED}server failed to start${NC}"
        cat "$LOG"
        exit 1
    fi
done

PATHS=("$@")
if [[ ${#PATHS[@]} -eq 0 ]]; then
    PATHS=("/home" "/resources")
    FIRST_POST=$(curl -s "http://localhost:${PORT}/rss.xml" | grep -oP '<link>https://www\.ethanthoma\.com/post/\K[^<]+' | head -1 || true)
    [[ -n "$FIRST_POST" ]] && PATHS+=("/post/${FIRST_POST}")
fi

CHROME="$(command -v chromium || command -v google-chrome-stable || command -v google-chrome || true)"
if [[ -z "$CHROME" ]]; then
    echo -e "${RED}no chromium/chrome on PATH${NC}"
    exit 1
fi

echo -e "${GREEN}» lighthouse (mobile)${NC}\n"
printf "  %-40s %5s %5s %5s %5s\n" "url" "perf" "lcp" "cls" "tbt"
printf "  %-40s %5s %5s %5s %5s\n" "----------------------------------------" "----" "----" "----" "----"

for path in "${PATHS[@]}"; do
    name="${path//\//_}"
    name="${name#_}"
    [[ -z "$name" ]] && name=root
    url="http://localhost:${PORT}${path}"

    lighthouse "$url" \
        --output=html --output=json \
        --output-path="${OUT}/${name}" \
        --chrome-path="$CHROME" \
        --chrome-flags="--headless=new --no-sandbox" \
        --quiet 2>/dev/null

    json="${OUT}/${name}.report.json"
    perf=$(jq -r '(.categories.performance.score * 100) | floor' "$json")
    lcp=$(jq -r '.audits["largest-contentful-paint"].displayValue' "$json")
    cls=$(jq -r '.audits["cumulative-layout-shift"].displayValue' "$json")
    tbt=$(jq -r '.audits["total-blocking-time"].displayValue' "$json")
    printf "  ${YELLOW}%-40s${NC} %5s %5s %5s %5s\n" "$path" "$perf" "$lcp" "$cls" "$tbt"
done

echo -e "\n${GREEN}✓ reports in tmp/lighthouse/${NC} (open *.report.html)"
