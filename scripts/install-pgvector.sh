#!/usr/bin/env bash
# Builds a user-owned, relocated copy of the system PostgreSQL 16 install with
# the Ubuntu pgvector package added, so the dedicated native instance can load
# `vector` without root. PostgreSQL resolves share/ and lib/ relative to its own
# executable, so bin/ is copied (symlinks would resolve back to /usr) while
# lib/ and share/ are symlink farms plus the extracted pgvector files.
set -euo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
RUNTIME="$ROOT/.runtime"
SYS_BIN=/usr/lib/postgresql/16/bin
SYS_LIB=/usr/lib/postgresql/16/lib
SYS_SHARE=/usr/share/postgresql/16
DIST="$RUNTIME/pgdist"
PG_BIN_OUT="$DIST/usr/lib/postgresql/16/bin"
if [[ -f "$DIST/usr/share/postgresql/16/extension/vector.control" && -x "$PG_BIN_OUT/postgres" ]]; then
  echo "$PG_BIN_OUT"
  exit 0
fi
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT
(cd "$WORK" && apt-get download postgresql-16-pgvector >/dev/null)
dpkg-deb -x "$WORK"/postgresql-16-pgvector_*.deb "$WORK/x"
rm -rf "$DIST.tmp"
mkdir -p "$DIST.tmp/usr/lib/postgresql/16/lib" "$DIST.tmp/usr/share/postgresql/16/extension"
cp -a "$SYS_BIN" "$DIST.tmp/usr/lib/postgresql/16/bin"
for f in "$SYS_LIB"/*; do ln -s "$f" "$DIST.tmp/usr/lib/postgresql/16/lib/"; done
for f in "$SYS_SHARE"/*; do
  [[ "$(basename "$f")" == extension ]] || ln -s "$f" "$DIST.tmp/usr/share/postgresql/16/"
done
for f in "$SYS_SHARE"/extension/*; do ln -s "$f" "$DIST.tmp/usr/share/postgresql/16/extension/"; done
cp -a "$WORK"/x/usr/lib/postgresql/16/lib/vector.so "$DIST.tmp/usr/lib/postgresql/16/lib/"
cp -a "$WORK"/x/usr/share/postgresql/16/extension/vector* "$DIST.tmp/usr/share/postgresql/16/extension/"
rm -rf "$DIST"
mv "$DIST.tmp" "$DIST"
echo "$PG_BIN_OUT"
