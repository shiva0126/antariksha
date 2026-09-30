#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION=2.11.4
case "$(uname -sm)" in 'Linux x86_64') ;; *) echo 'This pinned installer supports Linux x86_64 only.' >&2; exit 1;; esac
STAGING=$(mktemp -d "$ROOT/.runtime/caddy-install.XXXXXX")
ASSET="caddy_${VERSION}_linux_amd64.tar.gz"
URL="https://github.com/caddyserver/caddy/releases/download/v${VERSION}"
curl --fail --location --proto '=https' --max-time 120 "$URL/$ASSET" -o "$STAGING/$ASSET"
curl --fail --location --proto '=https' --max-time 30 "$URL/caddy_${VERSION}_checksums.txt" -o "$STAGING/checksums.txt"
EXPECTED=$(awk -v asset="$ASSET" '$2==asset {print $1}' "$STAGING/checksums.txt")
if [[ "$EXPECTED" =~ ^[0-9a-fA-F]{128}$ ]]; then
 ACTUAL=$(sha512sum "$STAGING/$ASSET")
elif [[ "$EXPECTED" =~ ^[0-9a-fA-F]{64}$ ]]; then
 ACTUAL=$(sha256sum "$STAGING/$ASSET")
else
 echo 'Missing or invalid upstream checksum' >&2; exit 1
fi
[[ "${ACTUAL%% *}" == "$EXPECTED" ]] || { echo 'Checksum mismatch' >&2; exit 1; }
tar -xzf "$STAGING/$ASSET" -C "$STAGING" caddy
install -m 755 "$STAGING/caddy" "$ROOT/bin/caddy"
"$ROOT/bin/caddy" validate --config "$ROOT/deploy/Caddyfile.wsl" --adapter caddyfile
echo 'HTTPS binary installed and configuration validated. Service is NOT started.'
echo "Temporary verification files retained in $STAGING"
