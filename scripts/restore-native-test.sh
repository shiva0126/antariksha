#!/usr/bin/env bash
set -euo pipefail
umask 077
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
PG_BIN="${PG_BIN:-$ROOT/.runtime/pgdist/usr/lib/postgresql/16/bin}"
archive="${1:?Usage: restore-native-test.sh /absolute/path/to/backup.dump}"
[[ -f "$archive" && -f "$archive.sha256" ]] || { echo 'Backup and checksum required' >&2; exit 1; }
sha256sum --check "$archive.sha256"
export PGHOST="${PGHOST:-$ROOT/.runtime}" PGPORT="${PGPORT:-55432}" PGUSER="${PGUSER:-panchang}"
test_db="astrisk_restore_$(date -u +%Y%m%d%H%M%S)_$$"
# Always restore into a newly created disposable database, never over production.
"$PG_BIN/createdb" "$test_db"
trap '"$PG_BIN/dropdb" --if-exists "$test_db"' EXIT
"$PG_BIN/pg_restore" --exit-on-error --no-owner --no-acl --dbname="$test_db" "$archive"
"$PG_BIN/psql" --dbname="$test_db" -v ON_ERROR_STOP=1 -c 'SELECT count(*) AS accounts_restored FROM member_accounts;' -c 'SELECT count(*) AS corpus_entries_restored FROM astro_corpus;'
echo 'Restore succeeded. Disposable restored database will be removed; production was untouched.'
