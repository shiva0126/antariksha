-- pgvector is optional at startup; install it in PostgreSQL before applying
-- this migration to enable semantic retrieval.
CREATE EXTENSION IF NOT EXISTS vector;
CREATE TABLE astro_corpus (
 id BIGSERIAL PRIMARY KEY, doc_type TEXT NOT NULL, key TEXT NOT NULL,
 language TEXT NOT NULL DEFAULT 'en', title TEXT, body TEXT NOT NULL,
 source TEXT, embedding vector(1536)
);
CREATE INDEX idx_corpus_key ON astro_corpus(doc_type,key,language);
CREATE INDEX idx_corpus_embedding ON astro_corpus USING ivfflat (embedding vector_cosine_ops) WHERE embedding IS NOT NULL;
CREATE TABLE reading_cache (
 id BIGSERIAL PRIMARY KEY, chart_hash TEXT UNIQUE NOT NULL,
 facts JSONB NOT NULL, reading JSONB NOT NULL, model TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_reading_cache_created ON reading_cache(created_at);
