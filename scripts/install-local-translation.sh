#!/usr/bin/env bash
# Installs the free local translation model for chat answers (English to
# eight Indian languages). No API key or paid service is used; the model runs
# on this CPU. IndicTrans2 is MIT-licensed, but Hugging Face serves it only to
# a signed-in account that has accepted its terms (approved automatically):
# visit https://huggingface.co/ai4bharat/indictrans2-en-indic-dist-200M,
# accept, then log in on this machine once (see docs/llm.md).
set -euo pipefail
cd "$(dirname "$0")/.."
DEPS=.runtime/translate-deps
MODELS=.runtime/translate-models
mkdir -p "$DEPS" "$MODELS"
if ! PYTHONPATH=$DEPS python3 -c 'import torch, transformers, IndicTransToolkit' 2>/dev/null; then
  python3 -m pip install --quiet --target "$DEPS" torch --index-url https://download.pytorch.org/whl/cpu
  python3 -m pip install --quiet --target "$DEPS" "transformers<4.50" sentencepiece sacremoses IndicTransToolkit
fi
# Download the model (about 1 GB) once into the cache.
PYTHONPATH=$DEPS python3 -c "
from transformers import AutoModelForSeq2SeqLM, AutoTokenizer
n='ai4bharat/indictrans2-en-indic-dist-200M'
AutoTokenizer.from_pretrained(n, trust_remote_code=True, cache_dir='$MODELS')
AutoModelForSeq2SeqLM.from_pretrained(n, trust_remote_code=True, cache_dir='$MODELS')"

mkdir -p ~/.config/systemd/user/panchang.service.d
install -m 0644 deploy/astrisk-translate.service ~/.config/systemd/user/
install -m 0644 deploy/panchang-translate.conf ~/.config/systemd/user/panchang.service.d/translate.conf
systemctl --user daemon-reload
systemctl --user enable --now astrisk-translate.service
for _ in $(seq 1 120); do curl -sf http://127.0.0.1:18093/healthz >/dev/null && break; sleep 1; done
curl -sf http://127.0.0.1:18093/healthz >/dev/null || { echo "translation server did not start; see journalctl --user -u astrisk-translate" >&2; exit 1; }
systemctl --user restart panchang.service
echo "Local translation ready."
