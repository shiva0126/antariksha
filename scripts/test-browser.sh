#!/usr/bin/env bash
# Runs the Playwright suite against a throwaway API on port 3100 backed by the
# disposable `panchang_test` database, never the live site or its members.
# Usage: scripts/test-browser.sh [playwright args…]
set -euo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
PORT="${TEST_PORT:-3100}"
if curl -fsS -m 2 "http://127.0.0.1:$PORT/healthz" >/dev/null 2>&1; then
  echo "port $PORT is already in use; stop that process or set TEST_PORT" >&2; exit 1
fi
DB_URL="$(bash "$ROOT/scripts/test-db.sh" | tail -1)"
# The book passages (Brihat Jataka, Brihat Samhita) as on the live site,
# without embeddings; sources whose scans are not on disk are skipped.
(cd "$ROOT" && env -u OPENAI_API_KEY DATABASE_URL="$DB_URL" CGO_ENABLED=1 "$HOME/.local/toolchains/go/bin/go" run ./cmd/corpus load -allow-missing-raw >/dev/null)
(cd "$ROOT" && PATH="$HOME/.local/toolchains/go/bin:$PATH" CGO_ENABLED=1 go build -buildvcs=false -o bin/panchang-api-test ./cmd/panchang-api)
# Build into dist-test: web/dist is what the live site serves.
(cd "$ROOT/web" && npx tsc -b && npx vite build --outDir dist-test --emptyOutDir >/dev/null)
log="$(mktemp)"
# A stand-in translator that tags text with its language (no model needed).
TPORT=$((PORT + 1))
python3 "$ROOT/scripts/fake-translate.py" "$TPORT" &
fake=$!
env -u OPENAI_API_KEY -u SMTP_HOST -u COOKIE_SECURE TRANSLATE_URL="http://127.0.0.1:$TPORT" RATE_LIMIT_SCALE=50 DATABASE_URL="$DB_URL" EPHE_PATH="$ROOT/ephe" WEB_DIST="$ROOT/web/dist-test" \
  HTTP_ADDR="127.0.0.1:$PORT" RAG_SEMANTIC_ENABLED=true EMBEDDING_PROVIDER=local "$ROOT/bin/panchang-api-test" >"$log" 2>&1 &
api=$!
trap 'kill $api $fake 2>/dev/null; rm -f "$log"' EXIT
for _ in $(seq 1 50); do curl -fsS "http://127.0.0.1:$PORT/healthz" >/dev/null 2>&1 && break; sleep 0.2; done
cd "$ROOT/web"
PLAYWRIGHT_BASE_URL="http://127.0.0.1:$PORT" npx playwright test "$@"
