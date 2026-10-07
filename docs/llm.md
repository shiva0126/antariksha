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

## Chart accuracy

`go run ./cmd/chartcheck -n 60` compares the engine with NASA JPL Horizons (an independent ephemeris)
at random moments from 1900 to 2050: apparent geocentric longitudes of the Sun, Moon and five
planets, and that every sidereal position equals the tropical one minus the true Lahiri ayanamsa.

Result on 2026-10-06, 60 moments: Sun 0.21″, Mercury 0.29″, Venus 0.22″, Mars 0.21″, Jupiter 0.23″,
Saturn 0.24″ largest difference; Moon 3.5″ (in 2050, where Earth-rotation timing is extrapolated).
One arc-second is 1/3600 of a degree, so every sign, nakshatra pada and house is exact.
Rahu and Ketu (true node) are not in Horizons; they are checked against `swetest` in
`engine/chart_reference_test.go`.

## Book rules on real charts

`go run ./cmd/bookbench -n 2000` (after `go run ./cmd/corpus build`) re-derives, from planet longitudes
alone, every condition Brihat Jataka (tr. Iyer 1885) defines and compares it with the engine; checks
that every detected condition has the book's passage for exactly that condition; and writes
`.runtime/llm-data/book-bench.jsonl`: question, verse, exact clause and plain answer, the test set a
model must answer exactly.

Each book entry in `corpus/maps/brihat_jataka_iyer_1885.json` holds the exact clause for its condition
(verified against the scan), `via` for rules that defer to another planet ("the same effects as the Sun
in those places", also verified), an `ocr` reading where the scan is too damaged to match the
corrected text, and a `plain` summary shown in answers as "Classical view (Brihat Jataka 20.3): …".
`corpus/map_exact_test.go` fails if a house passage does not speak about its own house.

Result on 2026-10-06, 2,000 charts: planet in house, Moon sign and birth star agree in all 18,000
checks. Two yoga rules differ from the book by design choice, pending a decision:

| Yoga | Charts that differ | Engine | Book |
|---|---|---|---|
| Adhi | 945 | one benefic in the 6th, 7th or 8th from the Moon is enough | benefics in the 6th, 7th and 8th (13.2) |
| Kemadruma | 833 | cancelled when any planet is in a kendra from the Moon (later Parashari rule) | no cancellation of that kind; Garga's cancellation is the Moon in a kendra from the ascendant or joined by a planet (13.3) |

Fixed on the same day: 33 house passages that carried another house's (or another planet's) clause,
for example Mercury in the 5th showed the Mars-in-the-ascendant verse, and the Anapha and Kemadruma
results, which had carried the Sunapha and Durudhura verses.

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

## Training on the books

Downloaded and verified (sha256 in `corpus/sources.json`, files under `corpus/raw/`, fetched again with
`go run ./cmd/corpus acquire`): Brihat Jataka in three translations (Iyer 1885 and 1905, Vijnanananda
1912) and Brihat Samhita in two (Iyer 1884, Kern 1865), about 456,000 words. Every other source in the
manifest is still `not_acquired`: pre-1931 English translations of Saravali, Phaladeepika, Jataka
Parijata and the others were not found on archive.org (only Sanskrit editions). Kalaprakashika (Subramonia
Iyer, first published 1917) is on archive.org as an undated reprint; it stays out until its rights are
confirmed.

`scripts/pack-llm-data.sh` builds `.runtime/llm-data/astrisk-train.zip`:

| File | What | Use |
|---|---|---|
| `books.jsonl` | book passages by chapter and verse | stage 1: read the books |
| `book-qa.jsonl` | 360 questions on the 141 checked rules, answered in the app's format with the verse | stage 2 |
| `train.jsonl` | ~4,600 engine-computed charts: chat questions and readings, answered from facts and rules (the draft answer is removed from these prompts so the model learns to compose, not copy) | stage 2 |
| `book-bench.jsonl` | 2,130 exact-answer test items | check |

`training/astrisk_finetune.ipynb` runs on a free Colab or Kaggle T4: LoRA on Qwen3-1.7B, stage 1 on the
books, stage 2 on the answers, a book-benchmark check, and export to `astrisk-qwen3-1.7b-Q8_0.gguf` for
`astrisk-llm.service`. The benchmark checks recall of the checked rules (the same rules are in training);
`cmd/llmdata eval` on unseen charts measures whether the model generalises.

## Roadmap

1. **Now:** local model live, eval baseline recorded, data generator.
2. **Better targets:** rewrite the grounded answers with a strong teacher model (when credits or a
   free tier are available), keep only rewrites that pass the same checks, add translations.
3. **Corpus:** add licence-checked public-domain classics; owner's data once reviewed.
4. **Fine-tune** (needs a GPU: free Colab/Kaggle for ≤4B, or a rented GPU for 7–8B): LoRA on the
   filtered pairs, export to GGUF, serve with the same llama.cpp service, compare on the eval set.
5. **Opt-in user feedback** (thumbs up/down on answers) as a preference signal for later rounds.


## Answers in the member's language

Chat answers are composed in English, then machine-translated by a local,
free model when the member's language is not English:

- **Model:** AI4Bharat IndicTrans2, distilled 200M English-to-Indic (MIT
  licence), int8 on the CPU, served by `scripts/translate-server.py` on
  loopback port 18093 (`astrisk-translate.service`). Covers Hindi, Marathi,
  Kannada, Tamil, Telugu, Malayalam, Gujarati and Bengali.
- **What is translated:** "In short" and "What this means for you", sentence
  by sentence (cached in memory). Headings and the closing line stay English
  in the stored text as markers and the app shows them in the member's
  language; the technical "Chart details" stay English.
- **The original is kept:** `chat_messages.original` and `lang` (migration
  22); the app shows "Machine translation from English · Show English".
- **Fallback:** without `TRANSLATE_URL`, or when the service fails or takes
  longer than `TRANSLATE_TIMEOUT` (45 s), the English answer is served.
- **Install:** `scripts/install-local-translation.sh`. AI4Bharat's own repo
  needs a signed-in Hugging Face account, so the script downloads a public
  copy and verifies every file against the checksums the official repo
  publishes (git blob SHA-1, or LFS SHA-256 for the weights); any mismatch
  stops the install. The weights are loaded from safetensors.
- **Cost on this machine:** about 1.5 GB of RAM (cap 2 GB), roughly 0.5–1 s
  per sentence on 2 threads; colons are translated around, since the model
  turns ":" into a visarga.
- Browser tests use `scripts/fake-translate.py`, which tags text with its
  language instead of translating.
