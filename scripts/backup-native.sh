#!/usr/bin/env bash
set -euo pipefail
umask 077
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
PG_BIN="${PG_BIN:-$ROOT/.runtime/pgdist/usr/lib/postgresql/16/bin}"
BACKUP_DIR="${BACKUP_DIR:-$ROOT/.runtime/backups}"
mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_DIR"
exec 9>"$BACKUP_DIR/.lock"
flock -n 9 || exit 0
export PGHOST="${PGHOST:-$ROOT/.runtime}" PGPORT="${PGPORT:-55432}" PGUSER="${PGUSER:-panchang}"
archive="$BACKUP_DIR/astrisk-$(date -u +%Y%m%dT%H%M%SZ).dump"
"$PG_BIN/pg_dump" --dbname="${PGDATABASE:-panchang}" --format=custom --no-owner --no-acl --file="$archive.partial"
"$PG_BIN/pg_restore" --list "$archive.partial" >/dev/null
mv "$archive.partial" "$archive"
sha256sum "$archive" >"$archive.sha256"
echo "Backup saved: $archive"
# Retain all backups by default. Never silently delete a user's recovery copies.
# These local archives need a separate protected off-device copy for disk failure.
