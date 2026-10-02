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

The English self-authored entries were revised to explain themes in ordinary language, offer optional reflection questions and avoid presenting a chart as evidence about a person's lived experience. Historical public-domain passages remain available for grounding and citations, but reading and chat responses do not repeat old prediction wording verbatim. These authored explanations are editorial interpretations, not translations.

### Editorial update — 2 October 2026

The follow-up rewrites 38 dignity/combustion/retrograde guides and 27 nakshatra guides, and expands seven yoga guides with examples, relationship reflections and limitations. These are 72 revised entries, **not 72 new books or additional corpus rows**. The 27 nakshatra and seven yoga entries now declare 34 concept-level sets of editorial references to already cleared Brihat Jataka passages. Build checks reject unknown verses, references to another concept/school, duplicate references and uncleared sources. Commentary is explicitly marked original editorial discussion, not translation or independent scholarly review.

The same Foster Press 1885 edition has a [Public Domain Mark in the Wellcome Collection](https://wellcomecollection.org/works/afmgm695). In contrast, the [Bhrigu Sutras transcription checked on Sanskrit Documents](https://sanskritdocuments.org/doc_z_misc_sociology_astrology/bhrigusUtram.html) requires permission for commercial copying/reposting. It has not been acquired or ingested; find another cleared edition or obtain permission. A text's ancient origin does not clear every digital transcription.

Readings now keep complete planet-guide paragraphs rather than clipping sentences. The cache key includes a digest of the retrieved bodies, titles and provenance plus model identity, so corpus edits invalidate generated prose automatically. This does not rewrite past conversation messages.

**More readings → Reading library & AI availability** and the Kundali Reading tab display actual database entry counts by source, pending/blocked sources, provider configuration and embedding readiness. A listed manifest source is never presented as an ingested book merely because it appears in the catalogue. Live API access, embeddings, new book editions and external kundli verification are still separate requirements.

## Retrieval and model status

The optional runtime semantic path and reading validation changes are documented in [Detailed matching explanations](match-explanations.md). `GET /api/reading/status` exposes current aggregate readiness to signed-in users. With `RAG_SEMANTIC_ENABLED=true`, chat enrichment searches only detected Parashari facts in English and only vectors from the configured model; exact-key retrieval remains the primary path. This flag alone does not create vectors or acquire books.

The live schema supports `vector(1536)` and has an HNSW index, but the corpus currently has **447 rows and zero embeddings**. Reading requests retrieve on exact engine keys. Semantic search cannot return useful neighbors until vectors are generated with a configured, compatible model. No book has been used for fine-tuning. Embedding this library requires a configured provider and its API spend; check row status after loading before describing semantic search as available.

## Lal Kitab and editions awaiting review

`corpus/sources.json` records the original five Lal Kitab volumes in their own `lalkitab` namespace as `not_acquired`. No Lal Kitab text or modern English rendering has been ingested. A scan site is not proof of commercial reuse permission, and a 1939 or 1952 publication date alone does not settle rights across jurisdictions, authorship, translations or editions. The Copyright Office of India application list includes an application for a 1939-titled Lal Kitab edition; that record is a reason to check the claimant and edition, not a final ownership ruling. Lal Kitab should remain distinct from Parashari rules even after permissions are cleared. Its vocabulary and house conventions need reviewed implementation before its passages can be mapped to detections.

Other requested texts—including BPHS, Bhrigu Sutras, Saravali, Phaladeepika, Jataka Parijata, Uttara Kalamrita, Sarvartha Chintamani, Hora Sara and Jaimini Sutras—remain source records with `not_acquired` status, not ingested books. The BPHS Santhanam English translation is explicitly blocked. A public-domain Sanskrit source still needs a careful human rendering and provenance check. References: [GRETIL's public-domain text catalogue](https://gretil.sub.uni-goettingen.de/gretil.html), [the Lal Kitab edition record on the Copyright Office site](https://copyright.gov.in/Documents/New_Applications/New_Applications_November_2024.pdf), and [Project Gutenberg's Manual of the Enumeration](https://www.gutenberg.org/ebooks/35998), which describes that separate numerology work as public domain in the US. The last source has not been added to Astrisk's corpus; territorial rights were not verified for every market, and Astrisk currently uses its own disclosed Pythagorean convention.

## pgvector without root

The system PostgreSQL has no pgvector. `scripts/install-pgvector.sh` builds a user-owned relocated copy of the PostgreSQL 16 install in `.runtime/pgdist` and adds Ubuntu's `postgresql-16-pgvector` package to it. The native instance runs from that copy.

## Tests

`go test ./corpus` covers the gate, segmentation, matching, coverage and namespacing, plus the public-domain build if `corpus/raw` has been acquired. Set `CORPUS_TEST_DATABASE_URL` to a disposable migrated database to also exercise sync, re-embedding and HNSW search against a fake embeddings server.

The 2 October editorial release passed all Go packages, targeted vet/race checks, the database source-count test, frontend unit/build checks, and two browser journeys covering Kundali and the additional readings. An isolated PostgreSQL load retained 447 entries; five-chart exact retrieval found all required tokens (269 passages). Production sync updated 72 rows, retained 375 unchanged, and removed none. The corpus validator still fails its embedding requirement (447 unembedded) and skips semantic search; automated OCR checks are not a substitute for a human passage review or independent chart validation.
