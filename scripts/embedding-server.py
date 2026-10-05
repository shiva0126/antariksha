#!/usr/bin/env python3
"""Local, free embedding server for the Astrisk reading library.

Serves an OpenAI-compatible POST /v1/embeddings on loopback port 18091, so the
Go API and `bin/corpus` use it through EMBEDDING_PROVIDER=local without any API
key. (Port 8091 is taken by Docker Desktop through WSL mirrored networking.)

Model: BAAI bge-small-en-v1.5 (MIT licence), the quantised ONNX export that
fastembed downloads from qdrant/bge-small-en-v1.5-onnx-q. Texts longer than 400
tokens are split into 400-token windows whose vectors are averaged. The 384
dimensions are zero-padded to the 1536 of the astro_corpus.embedding column;
zero padding leaves cosine distances unchanged. The model name encodes all of
this, so changing any step must change MODEL_NAME and re-embed the corpus.
"""
import json
import os
import sys
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

import numpy as np
from fastembed import TextEmbedding

MODEL_NAME = "bge-small-en-v1.5-onnx-q-chunk400-pad1536-v1"
DIMS, PAD, CHUNK = 384, 1536, 400
MAX_INPUTS, MAX_CHARS, MAX_BODY = 64, 20000, 2 << 20
HOST = os.environ.get("EMBEDDING_HOST", "127.0.0.1")
PORT = int(os.environ.get("EMBEDDING_PORT", "18091"))
CACHE = os.environ.get("EMBEDDING_CACHE", os.path.join(os.path.dirname(__file__), "..", ".runtime", "embedding-models"))

if HOST not in ("127.0.0.1", "::1"):
    sys.exit("embedding server must bind to loopback")

model = TextEmbedding("BAAI/bge-small-en-v1.5", cache_dir=CACHE, threads=int(os.environ.get("EMBEDDING_THREADS", "2")))
tokenizer = model.model.tokenizer
# One ONNX session and one tokenizer whose truncation windows() toggles: keep
# every call serial, which also keeps memory flat.
lock = threading.Lock()


def windows(text):
    """Split text into spans of at most CHUNK tokens, by character offsets."""
    tokenizer.no_truncation()
    try:
        enc = tokenizer.encode(text, add_special_tokens=False)
    finally:
        tokenizer.enable_truncation(max_length=512)
    offs = enc.offsets
    if len(offs) <= CHUNK:
        return [text]
    return [text[offs[i][0]:offs[min(i + CHUNK, len(offs)) - 1][1]] for i in range(0, len(offs), CHUNK)]


def embed(texts):
    spans, owner = [], []
    with lock:
        for i, t in enumerate(texts):
            for s in windows(t):
                spans.append(s)
                owner.append(i)
        vecs = list(model.embed(spans, batch_size=16))
    out = []
    for i in range(len(texts)):
        parts = [v for v, o in zip(vecs, owner) if o == i]
        v = np.mean(parts, axis=0)
        v = v / (np.linalg.norm(v) or 1.0)
        out.append(np.concatenate([v, np.zeros(PAD - DIMS, dtype=v.dtype)]).astype(float).round(7).tolist())
    return out


class Handler(BaseHTTPRequestHandler):
    server_version = "astrisk-embeddings"

    def send(self, code, body):
        data = json.dumps(body).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(data)))
        self.end_headers()
        self.wfile.write(data)

    def fail(self, code, msg):
        self.send(code, {"error": {"message": msg}})

    def do_GET(self):
        if self.path in ("/healthz", "/v1/models"):
            return self.send(200, {"status": "ok", "data": [{"id": MODEL_NAME, "object": "model"}]})
        self.fail(404, "not found")

    def do_POST(self):
        if self.path != "/v1/embeddings":
            return self.fail(404, "not found")
        n = int(self.headers.get("Content-Length") or 0)
        if n <= 0 or n > MAX_BODY:
            return self.fail(413, "request body too large or empty")
        try:
            req = json.loads(self.rfile.read(n))
        except ValueError:
            return self.fail(400, "invalid JSON")
        if req.get("model") not in (None, "", MODEL_NAME):
            return self.fail(400, f"this server only provides {MODEL_NAME}")
        inp = req.get("input")
        if isinstance(inp, str):
            inp = [inp]
        if not isinstance(inp, list) or not inp or len(inp) > MAX_INPUTS or not all(isinstance(t, str) and 0 < len(t) <= MAX_CHARS for t in inp):
            return self.fail(400, f"input must be 1-{MAX_INPUTS} non-empty strings of at most {MAX_CHARS} characters")
        try:
            vecs = embed(inp)
        except Exception as e:  # noqa: BLE001 - report, never crash the server
            return self.fail(500, f"embedding failed: {type(e).__name__}")
        self.send(200, {"object": "list", "model": MODEL_NAME, "data": [{"object": "embedding", "index": i, "embedding": v} for i, v in enumerate(vecs)], "usage": {"prompt_tokens": 0, "total_tokens": 0}})

    def log_message(self, fmt, *args):  # no request text in logs
        pass


if __name__ == "__main__":
    embed(["warm-up"])
    srv = ThreadingHTTPServer((HOST, PORT), Handler)
    print(f"embedding server ready on http://{HOST}:{PORT}/v1 ({MODEL_NAME})", flush=True)
    srv.serve_forever()
