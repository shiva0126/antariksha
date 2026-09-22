# Classical corpus (RAG)

The reading backend grounds each detected chart fact in corpus passages. The engine detects a fact, retrieval fetches the passages keyed to exactly that fact, and the LLM interprets them. Nothing is fine-tuned.

## Keys are the engine's vocabulary

`engine/corpus_keys.go` defines every token the engine can emit (`engine.CorpusVocabulary`) and the tokens present in a given chart (`engine.CorpusKeys`):

| doc_type | key examples | required |
|---|---|---|
| yoga | `gajakesari`, `budha_aditya`, `neecha_bhanga_raja_yoga` | 18 |
| graha_in_house | `mars_in_7` | 108 |
| graha_in_sign | `saturn_in_makara` | 108 |
| nakshatra | `revati` (Moon's star); `revati_pada2` optional | 27 |
| dasha | `dasha_saturn` (maha, antara and next maha lords) | 9 |
| dignity | `dignity_exalted`, `debilitated_mars`, `combust_mercury`, `retrograde_saturn` | 38 |
| bhava | `bhava_1`, `bhava_7`, `bhava_10`, `lagna_karka` | 24 |

Retrieval (`corpus.Retrieve`) is an exact `(doc_type, key)` fetch restricted to `system = 'parashari'`. Jaimini entries use the `karaka` namespace and can never answer a Parashari detection. A key may hold several rows, one per source passage. Readings return them as `grounding`.

## Layout

```
corpus/sources.json        rights manifest: every canonical text, with a decision and basis
corpus/authored/*.json     self-authored entries covering every required token
corpus/maps/*.json         public-domain verse → key maps with hand-corrected excerpts
corpus/raw/                downloaded sources (gitignored, sha256-pinned)
corpus/build/              entries.jsonl, report.json, audit/<source>/{clean.txt,segments.jsonl}
```

## Pipeline

```bash
export DATABASE_URL=postgresql:///panchang?host=$PWD/.runtime&port=55432&user=panchang&sslmode=disable
go run ./cmd/corpus acquire      # public_domain sources only; the download must match the manifest sha256
go run ./cmd/corpus build        # clean → segment → map → rights gate → entries.jsonl
go run ./cmd/corpus coverage     # exits non-zero unless every required token has an entry
go run ./cmd/corpus load -embed  # sync into astro_corpus and embed new or changed rows
go run ./cmd/corpus validate     # the §6 checklist against the live database
go run ./cmd/corpus search -q "Jupiter in a kendra from the Moon"
```

`scripts/start-native.sh` runs `acquire` and `load` on every start, adding `-embed` when `OPENAI_API_KEY` is set. It applies the migrations first.

**Rights gate.** Every entry must cite a manifest source whose decision is `public_domain`, `self_authored` or `licensed`, and must be a real engine token. A public-domain decision needs a year of 1929 or earlier, or a recorded life+60 basis. Known modern translations (for example Santhanam's BPHS) can only be recorded as `copyrighted_blocked`. Any gate failure aborts the build.

**Provenance.** Each public-domain excerpt is hand-corrected from the OCR and must fuzzy-match the segmented OCR of the stanza it cites (`corpus.MatchScore` ≥ 0.80). Corrected excerpts score about 0.85–1.0. A paraphrase scores about 0.2, and the same formula in the wrong stanza about 0.6. An `…` marks an omission, and each part must match on its own. Editor's glosses are stored visibly labelled.

**Sync.** Upserts are keyed on `(system, doc_type, key, language, source_id, ref)`. A row whose content hash changes loses its embedding until it is re-embedded. Rows missing from the build are deleted, so the build is the source of truth.

## Current contents

- 332 self-authored Parashari entries (full required coverage) and 7 Jaimini karaka entries.
- 108 passages from *The Brihat Jataka of Varaha Mihira*, tr. N. Chidambaram Iyer (1885, public domain), chapters XIII (Chandra yogas), XIV (planet pairs), XVI (Moon's nakshatra), XVII (Moon in signs) and XX (planets in houses, dignity scaling).
- Other canonical texts are recorded in the manifest as `not_acquired` or `copyrighted_blocked`, with notes on what clearing each would take.

## pgvector without root

The system PostgreSQL has no pgvector. `scripts/install-pgvector.sh` builds a user-owned relocated copy of the PostgreSQL 16 install in `.runtime/pgdist` and adds Ubuntu's `postgresql-16-pgvector` package to it. The native instance runs from that copy.

## Tests

`go test ./corpus` covers the gate, segmentation, matching, coverage and namespacing, plus the public-domain build if `corpus/raw` has been acquired. Set `CORPUS_TEST_DATABASE_URL` to a disposable migrated database to also exercise sync, re-embedding and HNSW search against a fake embeddings server.
