#!/usr/bin/env bash
# Checks Astrisk every few minutes (astrisk-health.timer) and restarts what is
# down: the API, the local embedding model and the Cloudflare tunnel. With
# ALERT_URL set (for example a free https://ntfy.sh/<your-topic>), it also
# posts a one-line alert when something had to be restarted or stays down.
set -uo pipefail
PUBLIC_URL="${PUBLIC_URL:-https://astrisk.space/healthz}"
alert() {
  echo "$1"
  if [[ -n "${ALERT_URL:-}" ]]; then curl -fsS -m 10 -d "Astrisk: $1" "$ALERT_URL" >/dev/null || true; fi
}
ok() { curl -fsS -m 10 -o /dev/null "$1"; }
restart() { systemctl --user restart "$1"; sleep "${2:-5}"; }

if ! ok http://127.0.0.1:3000/healthz; then
  restart panchang.service 8
  if ok http://127.0.0.1:3000/healthz; then alert "API was down and has been restarted."; else alert "API is down and did not recover after a restart."; exit 1; fi
fi
if systemctl --user is-enabled --quiet astrisk-embedding.service && ! ok http://127.0.0.1:18091/healthz; then
  restart astrisk-embedding.service 20
  ok http://127.0.0.1:18091/healthz && alert "Embedding model was down and has been restarted." || alert "Embedding model is down; chat still works without library search."
fi
if systemctl --user is-enabled --quiet astrisk-cloudflare.service && ! ok "$PUBLIC_URL"; then
  sleep 20
  if ! ok "$PUBLIC_URL"; then
    restart astrisk-cloudflare.service 20
    ok "$PUBLIC_URL" && alert "Public site was unreachable; the tunnel was restarted." || alert "Public site is unreachable (internet or Cloudflare). The app is running locally."
  fi
fi
exit 0
