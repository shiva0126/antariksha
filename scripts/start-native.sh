#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
RUNTIME="$ROOT/.runtime"
GO_BIN="${GO_BIN:-/home/shetty/.local/toolchains/go/bin/go}"
mkdir -p "$RUNTIME" "$ROOT/bin"
chmod 700 "$RUNTIME"
# pgvector is not installed system-wide; run the dedicated instance from a
# user-owned relocated PostgreSQL tree that includes it (see install-pgvector.sh).
PG_BIN="${PG_BIN:-$(bash "$ROOT/scripts/install-pgvector.sh" || echo /usr/lib/postgresql/16/bin)}"
if systemctl --user is-active --quiet panchang.service; then
  echo "Panchang is already running at http://localhost:3000"
  exit 0
fi
if [[ ! -f "$RUNTIME/postgres/PG_VERSION" ]]; then
  "$PG_BIN/initdb" -D "$RUNTIME/postgres" -U panchang --auth-local=trust --auth-host=reject --encoding=UTF8 --locale=C
fi
if ! "$PG_BIN/pg_ctl" -D "$RUNTIME/postgres" status >/dev/null 2>&1; then
  "$PG_BIN/pg_ctl" -D "$RUNTIME/postgres" -l "$RUNTIME/postgres.log" -o "-k $RUNTIME -p 55432 -c listen_addresses=''" -w start
fi
export PGHOST="$RUNTIME" PGPORT=55432 PGUSER=panchang
if ! psql -d postgres -Atqc "SELECT 1 FROM pg_database WHERE datname='panchang'" | grep -q '^1$'; then
  createdb panchang
fi
export PGDATABASE=panchang
psql -v ON_ERROR_STOP=1 -c 'CREATE TABLE IF NOT EXISTS native_schema_migrations (version INTEGER PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())'
if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=1' | grep -q '^1$'; then
  psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000001_init.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(1)'
fi
psql -v ON_ERROR_STOP=1 -f db/seed/festival_rules.sql
if psql -Atqc "SELECT 1 FROM pg_available_extensions WHERE name='vector'" | grep -q '^1$'; then
  if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=2' | grep -q '^1$'; then
    psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000002_reading.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(2)'
  fi
  if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=3' | grep -q '^1$'; then
    psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000003_corpus_provenance.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(3)'
  fi
  if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=4' | grep -q '^1$'; then
    psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000004_chat.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(4)'
  fi
  CORPUS_DB=1
else
  echo 'pgvector extension unavailable; reading cache/RAG tables were not applied. Facts and fallback readings remain available.'
fi
if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=5' | grep -q '^1$'; then
  psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000005_accounts.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(5)'
fi
"$GO_BIN" build -buildvcs=false -o bin/panchang-api ./cmd/panchang-api
if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=6' | grep -q '^1$'; then
  psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000006_community.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(6)'
fi
if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=7' | grep -q '^1$'; then
  psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000007_chat_access.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(7)'
fi
if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=8' | grep -q '^1$'; then
  psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000008_email_onboarding.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(8)'
fi
"$GO_BIN" build -buildvcs=false -o bin/corpus ./cmd/corpus
if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=9' | grep -q '^1$'; then
  psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000009_notifications.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(9)'
fi
(cd web && npm run build)
if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=10' | grep -q '^1$'; then
  psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000010_family_assistance.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(10)'
fi
if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=11' | grep -q '^1$'; then
  psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000011_chart_storage_email.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(11)'
fi
export DATABASE_URL="postgresql:///panchang?host=$RUNTIME&port=55432&user=panchang&sslmode=disable"
if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=12' | grep -q '^1$'; then
  psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000012_matrimony_biodata.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(12)'
fi
if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=13' | grep -q '^1$'; then
  psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000013_character.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(13)'
fi
if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=14' | grep -q '^1$'; then
  psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000014_admin.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(14)'
fi
if [[ "${CORPUS_DB:-}" == 1 ]]; then
  # Classical sources are fetched once and sha256-verified; offline starts
  # still load the full self-authored corpus.
  bin/corpus acquire || echo 'corpus acquire failed; loading without public-domain passages'
  EMBED_FLAG=()
  [[ -n "${OPENAI_API_KEY:-}" ]] && EMBED_FLAG=(-embed)
  bin/corpus load -allow-missing-raw "${EMBED_FLAG[@]}"
fi
export EPHE_PATH="$ROOT/ephe" WEB_DIST="$ROOT/web/dist" HTTP_ADDR="0.0.0.0:3000"
systemctl --user daemon-reload
systemctl --user start panchang.service
for attempt in {1..30}; do
  if curl -fsS http://127.0.0.1:3000/healthz >/dev/null; then
    echo 'Panchang is running at http://localhost:3000'
    exit 0
  fi
  if ! systemctl --user is-active --quiet panchang.service; then
    echo 'Application exited; run journalctl --user -u panchang.service' >&2
    exit 1
  fi
  sleep 1
done
echo "Startup check timed out; inspect $RUNTIME/app.log" >&2
exit 1
