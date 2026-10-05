#!/usr/bin/env bash
# Applies every db/migrations/*.up.sql not yet recorded in
# native_schema_migrations, in order, each in its own transaction.
# Usage: scripts/migrate.sh [database]   (PGHOST/PGPORT/PGUSER from the caller)
set -euo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
DB="${1:-${PGDATABASE:-panchang}}"
PSQL="${PSQL:-psql}"
"$PSQL" -d "$DB" -v ON_ERROR_STOP=1 -qc 'CREATE TABLE IF NOT EXISTS native_schema_migrations (version INTEGER PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())'
for f in "$ROOT"/db/migrations/*.up.sql; do
  v=$((10#$(basename "$f" | cut -c1-6)))
  if ! "$PSQL" -d "$DB" -Atqc "SELECT 1 FROM native_schema_migrations WHERE version=$v" | grep -q '^1$'; then
    "$PSQL" -d "$DB" -v ON_ERROR_STOP=1 -q --single-transaction -f "$f" -c "INSERT INTO native_schema_migrations(version) VALUES($v)"
    echo "applied migration $v to $DB"
  fi
done
