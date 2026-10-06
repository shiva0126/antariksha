# The Astrisk language model

Goal: the most accurate and most trustworthy astrology assistant. "Best" means three measurable things,
in this order:

1. **Never wrong about the chart.** Positions, houses, dashas and yogas come from the Swiss Ephemeris
   engine; the model only explains them. Every answer is checked against the engine's facts.
2. **Grounded in the tradition.** Interpretations come from the corpus (classical texts and our own
   entries), cited, never invented.
3. **Clear and humane.** Plain language, in the user's language, with no fatalistic predictions.

A general model can be fluent but gets charts wrong; a model trained only on old texts repeats harmful
predictions. Astrisk's advantage is the engine plus the corpus, so the model is a narrator over both.

## How it runs today

| Piece | Where |
|---|---|
| Chart facts | `engine/` (Swiss Ephemeris, cgo) |
| Retrieval | `reading/corpus.go`, pgvector + local bge-small embeddings (`astrisk-embedding.service`) |
| Grounded answer (always available) | `reading/insight.go`, `reading/chat.go` `compose` |
| Language model | `astrisk-llm.service` (`~/.config/systemd/user`): llama.cpp on 127.0.0.1:18092, Qwen3-0.6B now, Qwen3-1.7B once the GPU works (both Apache-2.0) |
| Wiring | `LLM_BASE_URL` (+ `LLM_MODEL`, `LLM_TIMEOUT`, default 75 s) in the panchang service |

With `LLM_BASE_URL` set, prompts use the compact digest (`reading/digest.go`): the same facts as
canonical sentences, about a third of the size, because prompt length sets CPU latency. If the model
is slow, down or returns something that fails validation, the grounded answer is served instead, so
the model can only improve an answer, never break one.

### Measured on this host (2026-10-06)

WSL is limited by `%USERPROFILE%\.wslconfig` to 4 threads (2 physical cores) and 6 GB, out of
4 cores / 8 threads / 15.8 GB on the laptop (i7-10510U).

| Model (llama.cpp CPU) | prompt tok/s | output tok/s | one chat answer | eval (chat) |
|---|---|---|---|---|
| Qwen3-1.7B Q8_0 | 19–25 | 4.4 | 75–150 s | not finished |
| Qwen3-0.6B Q8_0 | 61 | 9 | 29–194 s, median 104 s | 4/6 pass (2 missing disclaimer); some garbled facts |

That is too slow for live use (Cloudflare drops requests after 100 s), so production does **not** set
`LLM_BASE_URL` yet and users keep getting the instant grounded answer. `astrisk-llm.service` is
installed but stopped.

To make it fast enough:

1. **GPU:** the GeForce MX330 (2 GB, Pascal, sm_61) is usable from WSL once Windows has an NVIDIA
   driver from the R580 branch (the last branch that supports Pascal; the installed 441.45 is from
   2019 and has no WSL CUDA). llama.cpp's `ubuntu-cuda-12.8` build includes `61-virtual`. A 1.7B
   model fits fully in 2 GB and should answer in roughly 10–20 s.
2. **CPU share:** raising `.wslconfig` to `processors=8`, `memory=10GB` roughly doubles CPU speed.
   It needs `wsl --shutdown`, which restarts the site.

## Data

`go run ./cmd/llmdata gen -n 20000` writes chat-format JSONL (`.runtime/llm-data/train.jsonl`):
seeded random births across Indian and world cities → engine facts → the exact production prompt →
the grounded answer as target. It has no copyright or privacy questions and can be made as large as
needed. Its targets are correct but plain; the next step is better targets (below).

| Source | Status | Rule |
|---|---|---|
| Engine-generated charts | done (`cmd/llmdata`) | unlimited |
| Public-domain classics | 1 text (Brihat Jataka, Iyer 1885) | only translations published before 1930 or with an explicit free licence; record each in `corpus/sources.json` with its rights basis |
| Owner's own files | waiting for the path | check each item's rights before use |
| Astrisk users' questions | not collected | needs an explicit opt-in and anonymisation first; birth details never leave the server |

## Measuring it

`go run ./cmd/llmdata eval -n 10 [-readings] -model <name> -base <url>` scores any OpenAI-compatible
model on a fixed set of charts and questions that never overlaps the training seed:

- readings: valid JSON, every placement copied exactly, exactly the detected yogas (`ValidateReading`)
- chat: valid JSON, 60–320 words, the disclaimer, no forbidden predictions, the right mahadasha lord
  for timing questions, no wrong Moon or lagna sign (`ScoreAnswer`)

Results go to `.runtime/llm-data/eval-<model>.json`. Every model change must beat the previous score.

## Roadmap

1. **Now:** local model live, eval baseline recorded, data generator.
2. **Better targets:** rewrite the grounded answers with a strong teacher model (when credits or a
   free tier are available), keep only rewrites that pass the same checks, add translations.
3. **Corpus:** add licence-checked public-domain classics; owner's data once reviewed.
4. **Fine-tune** (needs a GPU: free Colab/Kaggle for ≤4B, or a rented GPU for 7–8B): LoRA on the
   filtered pairs, export to GGUF, serve with the same llama.cpp service, compare on the eval set.
5. **Opt-in user feedback** (thumbs up/down on answers) as a preference signal for later rounds.
