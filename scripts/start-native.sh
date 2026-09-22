#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
RUNTIME="$ROOT/.runtime"
PG_BIN="${PG_BIN:-/usr/lib/postgresql/16/bin}"
GO_BIN="${GO_BIN:-/home/shetty/.local/toolchains/go/bin/go}"
mkdir -p "$RUNTIME" "$ROOT/bin"
chmod 700 "$RUNTIME"
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
if ! psql -d postgres -Atqc "SELECT 1 FROM pg_database WHERE datname='panchang'" | rg -q '^1$'; then
  createdb panchang
fi
export PGDATABASE=panchang
psql -v ON_ERROR_STOP=1 -c 'CREATE TABLE IF NOT EXISTS native_schema_migrations (version INTEGER PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())'
if ! psql -Atqc 'SELECT 1 FROM native_schema_migrations WHERE version=1' | rg -q '^1$'; then
  psql -v ON_ERROR_STOP=1 --single-transaction -f db/migrations/000001_init.up.sql -c 'INSERT INTO native_schema_migrations(version) VALUES(1)'
fi
psql -v ON_ERROR_STOP=1 -f db/seed/festival_rules.sql
"$GO_BIN" build -buildvcs=false -o bin/panchang-api ./cmd/panchang-api
(cd web && npm run build)
export DATABASE_URL="postgresql:///panchang?host=$RUNTIME&port=55432&user=panchang&sslmode=disable"
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
