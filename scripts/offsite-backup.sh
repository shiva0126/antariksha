#!/usr/bin/env bash
# Copies the newest local database backup off this PC, encrypted.
#   BACKUP_PASSPHRASE_FILE  file holding the encryption passphrase (keep a copy
#                           somewhere safe: without it the backups cannot be read)
#   RCLONE_REMOTE           an rclone destination, e.g. gdrive:astrisk-backups
#                           (run `rclone config` once to connect Google Drive,
#                           OneDrive, Dropbox, Backblaze B2 …)
#   or OFFSITE_DIR          a directory on another disk, e.g. /mnt/d/AstriskBackups
# The dump contains members' personal data, so it is never uploaded unencrypted.
set -euo pipefail
ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="${BACKUP_DIR:-$ROOT/.runtime/backups}"
latest=$(ls -1t "$BACKUP_DIR"/astrisk-*.dump 2>/dev/null | head -1 || true)
[[ -n "$latest" ]] || { echo "no local backup found; run scripts/backup-native.sh first" >&2; exit 1; }
[[ -r "${BACKUP_PASSPHRASE_FILE:-}" ]] || { echo "set BACKUP_PASSPHRASE_FILE to a readable passphrase file" >&2; exit 1; }
umask 077
enc="$latest.gpg"
gpg --batch --yes --pinentry-mode loopback --passphrase-file "$BACKUP_PASSPHRASE_FILE" --symmetric --cipher-algo AES256 -o "$enc" "$latest"
if [[ -n "${RCLONE_REMOTE:-}" ]]; then
  rclone copy "$enc" "$RCLONE_REMOTE" && echo "uploaded $(basename "$enc") to $RCLONE_REMOTE"
elif [[ -n "${OFFSITE_DIR:-}" ]]; then
  mkdir -p "$OFFSITE_DIR" && cp "$enc" "$OFFSITE_DIR/" && echo "copied $(basename "$enc") to $OFFSITE_DIR"
else
  echo "set RCLONE_REMOTE or OFFSITE_DIR" >&2; exit 1
fi
rm -f "$enc"
