#!/usr/bin/env bash
# Recreates the disposable test database `panchang_test` in the native cluster
# and applies every migration. Tests never touch the live `panchang` database.
# Prints the DATABASE_URL to use.
set -euo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
PG_BIN="${PG_BIN:-$ROOT/.runtime/pgdist/usr/lib/postgresql/16/bin}"
export PGHOST="$ROOT/.runtime" PGPORT=55432 PGUSER=panchang PSQL="$PG_BIN/psql"
"$PSQL" -d postgres -qc 'DROP DATABASE IF EXISTS panchang_test WITH (FORCE)'
"$PSQL" -d postgres -qc 'CREATE DATABASE panchang_test'
export PGOPTIONS='--client-min-messages=warning'
bash "$ROOT/scripts/migrate.sh" panchang_test >/dev/null
"$PSQL" -d panchang_test -v ON_ERROR_STOP=1 -q -f "$ROOT/db/seed/festival_rules.sql"
echo "postgresql:///panchang_test?host=$ROOT/.runtime&port=55432&user=panchang&sslmode=disable"
