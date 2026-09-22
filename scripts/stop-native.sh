#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
RUNTIME="$ROOT/.runtime"
PG_BIN="${PG_BIN:-/usr/lib/postgresql/16/bin}"
[[ -x "$RUNTIME/pgdist/usr/lib/postgresql/16/bin/pg_ctl" ]] && PG_BIN="$RUNTIME/pgdist/usr/lib/postgresql/16/bin"
systemctl --user stop panchang.service
if [[ -f "$RUNTIME/app.pid" ]]; then
  APP_PID="$(<"$RUNTIME/app.pid")"
  if kill -0 "$APP_PID" 2>/dev/null; then
    APP_EXE="$(readlink "/proc/$APP_PID/exe" || true)"
    if [[ "$APP_EXE" != "$ROOT/bin/panchang-api" ]]; then
      echo 'PID belongs to a different process; refusing to stop it.' >&2
      exit 1
    fi
    kill "$APP_PID"
  fi
fi
if "$PG_BIN/pg_ctl" -D "$RUNTIME/postgres" status >/dev/null 2>&1; then
  "$PG_BIN/pg_ctl" -D "$RUNTIME/postgres" -m fast -w stop
fi
echo 'Panchang stopped. Database files are preserved.'
