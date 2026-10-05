#!/usr/bin/env bash
# Installs the free local embedding model for the reading library and embeds
# the corpus. No API key or paid service is used; the model runs on this CPU.
set -euo pipefail
cd "$(dirname "$0")/.."
DEPS=.runtime/embedding-deps
MODELS=.runtime/embedding-models
mkdir -p "$DEPS" "$MODELS"
if ! PYTHONPATH=$DEPS python3 -c 'import fastembed' 2>/dev/null; then
  python3 -m pip install --quiet --target "$DEPS" -r scripts/embedding-requirements.txt
fi
# Download bge-small-en-v1.5 (MIT licence, about 65 MB) once into the cache.
PYTHONPATH=$DEPS python3 -c "from fastembed import TextEmbedding; TextEmbedding('BAAI/bge-small-en-v1.5', cache_dir='$MODELS')"

mkdir -p ~/.config/systemd/user/panchang.service.d
install -m 0644 deploy/astrisk-embedding.service ~/.config/systemd/user/
install -m 0644 deploy/panchang-embedding.conf ~/.config/systemd/user/panchang.service.d/embedding.conf
systemctl --user daemon-reload
systemctl --user enable --now astrisk-embedding.service
for _ in $(seq 1 60); do curl -sf http://127.0.0.1:18091/healthz >/dev/null && break; sleep 1; done
curl -sf http://127.0.0.1:18091/healthz >/dev/null || { echo "embedding server did not start; see journalctl --user -u astrisk-embedding" >&2; exit 1; }

# Embed every corpus row (only rows whose text or model changed are re-sent).
DATABASE_URL=${DATABASE_URL:-"postgresql:///panchang?host=$PWD/.runtime&port=55432&user=panchang&sslmode=disable"} \
  EMBEDDING_PROVIDER=local ./bin/corpus load -allow-missing-raw -embed
systemctl --user restart panchang.service
echo "Local embeddings ready."
