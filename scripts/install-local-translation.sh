#!/usr/bin/env bash
# Installs the free local translation model for chat answers (English to
# eight Indian languages). No API key or paid service is used; the model runs
# on this CPU. IndicTrans2 is MIT-licensed; AI4Bharat's own Hugging Face repo
# needs a signed-in account, so the files come from a public copy and are
# checked byte for byte against the checksums the official repo publishes.
# Any mismatch stops the install.
set -euo pipefail
cd "$(dirname "$0")/.."
DEPS=.runtime/translate-deps
MODELS=.runtime/translate-models
mkdir -p "$DEPS" "$MODELS"
if ! PYTHONPATH=$DEPS python3 -c 'import torch, transformers, IndicTransToolkit' 2>/dev/null; then
  python3 -m pip install --quiet --target "$DEPS" torch --index-url https://download.pytorch.org/whl/cpu
  python3 -m pip install --quiet --target "$DEPS" "transformers<4.50" sentencepiece sacremoses IndicTransToolkit
fi
# Download the model (about 1 GB) once and verify it.
DIR=$MODELS/indictrans2-en-indic-dist-200M
OFFICIAL=ai4bharat/indictrans2-en-indic-dist-200M
COPY=${TRANSLATE_MODEL_COPY:-naklitechie/indictrans2-en-indic-dist-200M}
mkdir -p "$DIR"
curl -sf "https://huggingface.co/api/models/$OFFICIAL?blobs=true" >"$DIR/.official.json"
for f in LICENSE config.json configuration_indictrans.py dict.SRC.json dict.TGT.json generation_config.json model.SRC model.TGT \
  model.safetensors modeling_indictrans.py special_tokens_map.json tokenization_indictrans.py tokenizer_config.json; do
  [ -s "$DIR/$f" ] || curl -sfL -o "$DIR/$f" "https://huggingface.co/$COPY/resolve/main/$f"
done
python3 - "$DIR" <<'PY'
import hashlib, json, os, sys
d = sys.argv[1]
bad = []
for s in json.load(open(os.path.join(d, ".official.json")))["siblings"]:
    path = os.path.join(d, s["rfilename"])
    if not os.path.exists(path):
        continue
    b = open(path, "rb").read()
    if s.get("lfs"):
        ok = hashlib.sha256(b).hexdigest() == s["lfs"]["sha256"]
    else:
        ok = hashlib.sha1(b"blob %d\0" % len(b) + b).hexdigest() == s["blobId"]
    if not ok:
        bad.append(s["rfilename"])
if bad:
    sys.exit("these files differ from the official AI4Bharat release: " + ", ".join(bad))
print("model files match the official AI4Bharat release")
PY
rm -f "$DIR/.official.json"

mkdir -p ~/.config/systemd/user/panchang.service.d
install -m 0644 deploy/astrisk-translate.service ~/.config/systemd/user/
install -m 0644 deploy/panchang-translate.conf ~/.config/systemd/user/panchang.service.d/translate.conf
systemctl --user daemon-reload
systemctl --user enable --now astrisk-translate.service
for _ in $(seq 1 180); do curl -sf -m 2 http://127.0.0.1:18093/healthz >/dev/null && break; sleep 1; done
curl -sf http://127.0.0.1:18093/healthz >/dev/null || { echo "translation server did not start; see journalctl --user -u astrisk-translate" >&2; exit 1; }
systemctl --user restart panchang.service
echo "Local translation ready."
