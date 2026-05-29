#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DESKTOP_DIR="$(dirname "$SCRIPT_DIR")"
MOCK_PORT="${MOCK_OLLAMA_PORT:-11435}"
WAILS_URL="${WAILS_DEV_URL:-http://localhost:34115}"
MOCK_PID=""

cleanup() {
  echo ""
  echo "[e2e:run] cleaning up..."
  if [ -n "$MOCK_PID" ] && kill -0 "$MOCK_PID" 2>/dev/null; then
    kill "$MOCK_PID" 2>/dev/null || true
    echo "[e2e:run] stopped mock server (pid $MOCK_PID)"
  fi
}
trap cleanup EXIT

echo "=== NomNom Desktop E2E Test Runner ==="
echo ""

# — 1. Install e2e dependencies —
echo "[e2e:run] installing e2e dependencies..."
cd "$SCRIPT_DIR"
npm install --silent 2>&1 | tail -1

# — 2. Install Playwright browsers —
if ! npx playwright install chromium --with-deps 2>/dev/null; then
  echo "[e2e:run] playwright install skipped (already installed or no sudo)"
fi

# — 3. Start mock AI server —
echo "[e2e:run] starting mock AI server on port $MOCK_PORT..."
npx tsx mock-server.ts &
MOCK_PID=$!
sleep 1

if ! kill -0 "$MOCK_PID" 2>/dev/null; then
  echo "[e2e:run] ERROR: mock server failed to start"
  exit 1
fi
echo "[e2e:run] mock server running (pid $MOCK_PID)"

# — 4. Check for Wails dev server —
echo "[e2e:run] checking Wails dev server at $WAILS_URL..."
if curl -s -o /dev/null -w "%{http_code}" "$WAILS_URL" | grep -q "200\|302\|304"; then
  echo "[e2e:run] Wails dev server is reachable"
else
  echo "[e2e:run] Wails dev server NOT reachable at $WAILS_URL"
  echo "[e2e:run] please start it in another terminal:"
  echo ""
  echo "  cd $DESKTOP_DIR && OLLAMA_HOST=http://localhost:$MOCK_PORT wails dev"
  echo ""
  echo "[e2e:run] waiting 10s for you to start it..."
  sleep 10

  if ! curl -s -o /dev/null "$WAILS_URL"; then
    echo "[e2e:run] still not reachable — aborting."
    echo "[e2e:run] start wails dev manually and re-run: cd $SCRIPT_DIR && npx playwright test"
    exit 1
  fi
fi

# — 5. Run Playwright tests —
echo ""
echo "[e2e:run] running playwright tests..."
echo ""

cd "$SCRIPT_DIR"
npx playwright test "$@"

echo ""
echo "[e2e:run] done."
