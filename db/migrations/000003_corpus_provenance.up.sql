-- Corpus ingestion: rights-cleared, provenance-tracked, engine-keyed entries.
-- A key may hold several rows (one per source passage) so retrieval can merge
-- parallel grounding from multiple texts.
CREATE TABLE corpus_sources (
 id TEXT PRIMARY KEY, title TEXT NOT NULL, author TEXT, edition TEXT,
 translator TEXT, year INTEGER, url TEXT, sha256 TEXT,
 rights TEXT NOT NULL CHECK (rights IN ('public_domain','self_authored','licensed','copyrighted_blocked','not_acquired')),
 rights_basis TEXT NOT NULL, system TEXT NOT NULL DEFAULT 'parashari',
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
DELETE FROM astro_corpus;
ALTER TABLE astro_corpus
 ADD COLUMN system TEXT NOT NULL DEFAULT 'parashari' CHECK (system IN ('parashari','jaimini')),
 ADD COLUMN source_id TEXT NOT NULL REFERENCES corpus_sources(id),
 ADD COLUMN ref TEXT NOT NULL DEFAULT '',
 ADD COLUMN rights TEXT NOT NULL CHECK (rights IN ('public_domain','self_authored','licensed')),
 ADD COLUMN content_hash TEXT NOT NULL,
 ADD COLUMN embedding_model TEXT,
 ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 ALTER COLUMN source SET NOT NULL,
 ADD CONSTRAINT astro_corpus_doc_type_check CHECK (doc_type IN ('yoga','graha_in_house','graha_in_sign','nakshatra','dasha','dignity','bhava','aspect','karaka'));
CREATE UNIQUE INDEX astro_corpus_entry_uniq ON astro_corpus(system, doc_type, key, language, source_id, ref);
DROP INDEX IF EXISTS idx_corpus_key;
CREATE INDEX idx_corpus_lookup ON astro_corpus(system, language, doc_type, key);
-- ivfflat on a few hundred rows with default lists is poor; HNSW needs no training.
DROP INDEX IF EXISTS idx_corpus_embedding;
CREATE INDEX idx_corpus_embedding_hnsw ON astro_corpus USING hnsw (embedding vector_cosine_ops);
