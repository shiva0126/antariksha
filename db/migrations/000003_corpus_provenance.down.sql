DROP INDEX IF EXISTS idx_corpus_embedding_hnsw;
DROP INDEX IF EXISTS idx_corpus_lookup;
DROP INDEX IF EXISTS astro_corpus_entry_uniq;
DELETE FROM astro_corpus;
ALTER TABLE astro_corpus DROP CONSTRAINT IF EXISTS astro_corpus_doc_type_check,
 DROP COLUMN system, DROP COLUMN source_id, DROP COLUMN ref, DROP COLUMN rights,
 DROP COLUMN content_hash, DROP COLUMN embedding_model, DROP COLUMN updated_at,
 ALTER COLUMN source DROP NOT NULL;
CREATE INDEX idx_corpus_key ON astro_corpus(doc_type,key,language);
CREATE INDEX idx_corpus_embedding ON astro_corpus USING ivfflat (embedding vector_cosine_ops) WHERE embedding IS NOT NULL;
DROP TABLE corpus_sources;
