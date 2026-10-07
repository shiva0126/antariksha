#!/usr/bin/env python3
"""Local, free translation server for Astrisk chat answers.

Serves POST /translate {"lang": "hi", "texts": [...]} -> {"texts": [...]} on
loopback port 18093. English to Hindi, Marathi, Kannada, Tamil, Telugu,
Malayalam, Gujarati and Bengali with AI4Bharat IndicTrans2 (the distilled
200M English-to-Indic model, MIT licence), quantised to int8 for the CPU.
The Go API calls it through TRANSLATE_URL and falls back to English when it
is missing or slow.
"""
import gc
import json
import os
import sys
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

import torch
from IndicTransToolkit import IndicProcessor
from transformers import AutoModelForSeq2SeqLM, AutoTokenizer

# A local directory holding the official AI4Bharat files (see
# scripts/install-local-translation.sh, which verifies them), or a Hugging
# Face model id.
MODEL = os.environ.get("TRANSLATE_MODEL", os.path.join(os.path.dirname(__file__), "..", ".runtime", "translate-models", "indictrans2-en-indic-dist-200M"))
CACHE = os.environ.get("TRANSLATE_CACHE", os.path.join(os.path.dirname(__file__), "..", ".runtime", "translate-models"))
HOST = os.environ.get("TRANSLATE_HOST", "127.0.0.1")
PORT = int(os.environ.get("TRANSLATE_PORT", "18093"))
BEAMS = int(os.environ.get("TRANSLATE_BEAMS", "3"))
MAX_TEXTS, MAX_CHARS, MAX_BODY, BATCH = 128, 1200, 1 << 20, 16
LANGS = {"hi": "hin_Deva", "mr": "mar_Deva", "kn": "kan_Knda", "ta": "tam_Taml", "te": "tel_Telu",
         "ml": "mal_Mlym", "gu": "guj_Gujr", "bn": "ben_Beng"}

if HOST not in ("127.0.0.1", "::1"):
    sys.exit("translation server must bind to loopback")

torch.set_num_threads(int(os.environ.get("TRANSLATE_THREADS", "2")))
tokenizer = AutoTokenizer.from_pretrained(MODEL, trust_remote_code=True, cache_dir=CACHE)
model = AutoModelForSeq2SeqLM.from_pretrained(MODEL, trust_remote_code=True, cache_dir=CACHE).eval()
model = torch.quantization.quantize_dynamic(model, {torch.nn.Linear}, dtype=torch.qint8)
gc.collect()
processor = IndicProcessor(inference=True)
lock = threading.Lock()  # one generation at a time keeps memory flat


def translate(texts, lang):
    # The model renders ":" as a visarga, so each text is split at ": " and
    # the pieces are translated separately and joined with a real colon.
    pieces, counts = [], []
    for t in texts:
        parts = t.split(": ")
        pieces.extend(parts)
        counts.append(len(parts))
    done = translate_pieces(pieces, lang)
    out, i = [], 0
    for n in counts:
        # The model ends each piece as a sentence; only the last keeps it.
        parts = [p.rstrip().rstrip(".।") for p in done[i:i + n - 1]] + done[i + n - 1:i + n]
        out.append(": ".join(parts))
        i += n
    return out


def translate_pieces(texts, lang):
    tag = LANGS[lang]
    out = []
    for i in range(0, len(texts), BATCH):
        chunk = texts[i:i + BATCH]
        batch = processor.preprocess_batch(chunk, src_lang="eng_Latn", tgt_lang=tag)
        enc = tokenizer(batch, truncation=True, padding="longest", return_tensors="pt", max_length=256)
        with torch.inference_mode():
            gen = model.generate(**enc, use_cache=True, min_length=0, max_length=256, num_beams=BEAMS, num_return_sequences=1)
        dec = tokenizer.batch_decode(gen, skip_special_tokens=True, clean_up_tokenization_spaces=True)
        out.extend(processor.postprocess_batch(dec, lang=tag))
    return out


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def reply(self, code, obj):
        body = json.dumps(obj, ensure_ascii=False).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        if self.path == "/healthz":
            return self.reply(200, {"status": "ok", "model": MODEL, "languages": sorted(LANGS)})
        self.reply(404, {"error": "not found"})

    def do_POST(self):
        if self.path != "/translate":
            return self.reply(404, {"error": "not found"})
        n = int(self.headers.get("Content-Length") or 0)
        if n <= 0 or n > MAX_BODY:
            return self.reply(413, {"error": "body too large"})
        try:
            req = json.loads(self.rfile.read(n))
            lang, texts = req["lang"], req["texts"]
        except (ValueError, KeyError, TypeError):
            return self.reply(400, {"error": "expected {lang, texts}"})
        if lang not in LANGS:
            return self.reply(400, {"error": "unsupported language"})
        if not isinstance(texts, list) or len(texts) > MAX_TEXTS or not all(isinstance(t, str) and len(t) <= MAX_CHARS for t in texts):
            return self.reply(400, {"error": "texts must be at most %d strings of %d characters" % (MAX_TEXTS, MAX_CHARS)})
        with lock:
            out = translate(texts, lang)
        self.reply(200, {"texts": out})


if __name__ == "__main__":
    ThreadingHTTPServer((HOST, PORT), Handler).serve_forever()
