# Detailed matching explanations

Implemented 2026-10-02. Two separate reports are available:

- **Matching**: the server computes Ashtakoota points and Mars flags, then explains each of the eight factors plus Mars in everyday language. The points are not a probability of a successful marriage.
- **Community → Matrimony**: select an eligible discovered profile to compare location, timeline, family plans, relocation, lifestyle, values, hobbies and shared selected interests. Missing information stays unknown. Normalized text equality is explicitly distinguished from agreement in meaning.

Both reports include conversation questions, limitations, sources/method labels and a deterministic guide when no model is configured. These are original project-authored guides, not translations or claims to have consulted additional books.

## Optional AI

The checkbox requests a hosted-model explanation using the existing configured LLM. Chart explanations send scores and flags, not names, raw birth details or caste classifications. Profile explanations send only comparison statuses and shared-interest counts, never raw free text, private character data, social URLs, photos, messages or contacts. No scraping or personal-trait inference is performed.

The model returns one interpretive section for each server-selected factor. IDs, order, content lengths and conservative safety patterns are checked. Evidence, scores, sources, summary, limitations and disclaimer remain server-owned. Model failures and rejected output fall back visibly; unvalidated output is never streamed. These checks are not a guarantee that every natural-language claim is correct. Expert evaluation and stronger semantic review remain open.

Requests have a 30-second model deadline, a shared per-account AI budget of four requests per minute, same-origin protection and ordinary authentication. Matrimony eligibility is checked before retrieval and again after the model returns. Blocks, profile pauses, reciprocal age restrictions and moderation apply. Explanations are not persisted or cached; they are recomputed from currently visible facts. There is no paid-provider call when AI is unchecked.

Endpoints:

- `POST /api/match`: existing `boy`/`girl` inputs plus optional `use_ai`; response adds `explanation`.
- `POST /api/matrimony/explanation/{peer}`: `{ "use_ai": false }`; ownership is always the signed-in member, not a client-supplied account.
- `GET /api/reading/status`: authenticated aggregate corpus/provider readiness and source manifest decisions; no credentials or user data.

## Reading and RAG changes

Reading validation now requires exact canonical planet identity/placement fields and the exact multiset of yoga names and strengths (the same named yoga may legitimately form for multiple planets). It additionally checks catalogued yoga names in narrative, common explicit house assertions, current-period pairs and dates. The reading cache format was versioned to invalidate earlier generated output. Narrative paraphrases can still evade pattern checks; do not claim a complete hallucination-proof system.

Set `RAG_SEMANTIC_ENABLED=true` only after loading compatible vectors. Chat then performs exact retrieval first and optionally enriches context with up to six semantic matches restricted **before ranking** to the chart's detected tokens, Parashari namespace, English and configured embedding model. Empty indexes do not trigger paid query embeddings. Exact retrieval remains available when enrichment fails.

Use the existing `corpus load -embed` and `corpus validate` commands with a configured provider. Fake vectors used by tests are never production embeddings. This release does not add or clear book editions: the current 447-entry corpus, pending Lal Kitab and other sources remain accurately reported. Book acquisition, human review and independent five-birth yoga/dasha comparison remain outstanding; the new five-birth regression test is not an external validation.

The API adapter continues to use JSON mode with application-level validation; it does not claim strict provider-schema enforcement. Design reference: [official OpenAI structured-output guidance](https://developers.openai.com/api/docs/guides/structured-outputs).

## Verification for this release

- `go test ./...`, targeted `go vet`, and race tests for `reading` / `corpus` pass.
- An isolated migrated PostgreSQL database exercises profile consent, post-generation revocation, community access, embedding synchronization and detected-key semantic filtering with a fake provider.
- Browser tests cover all nine chart factors, the clearly labeled unavailable-AI fallback, eight profile factors, unknown preferences and blocked-profile denial. Frontend TypeScript build and seven existing unit tests pass.
- Production corpus remains 447 rows / 0 vectors. There was no paid model call, external five-birth certification or book ingestion in this release.
