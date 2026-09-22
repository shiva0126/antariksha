-- Chat history: one session per birth chart per browser; every question and
-- answer is kept with the corpus passages that grounded it.
CREATE TABLE chat_sessions (
 id TEXT PRIMARY KEY, chart_hash TEXT NOT NULL, birth JSONB NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_chat_sessions_chart ON chat_sessions(chart_hash);
CREATE TABLE chat_messages (
 id BIGSERIAL PRIMARY KEY,
 session_id TEXT NOT NULL REFERENCES chat_sessions(id) ON DELETE CASCADE,
 role TEXT NOT NULL CHECK (role IN ('user','assistant')),
 content TEXT NOT NULL, topics TEXT[] NOT NULL DEFAULT '{}',
 sources JSONB NOT NULL DEFAULT '[]', model TEXT NOT NULL DEFAULT '',
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_chat_messages_session ON chat_messages(session_id, id);
-- Readings cached before the Vimshottari fix carry wrong dashas; the cache is
-- now keyed by as-of day as well.
DELETE FROM reading_cache;
