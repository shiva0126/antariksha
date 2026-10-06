ALTER TABLE member_sessions
 ADD COLUMN id BIGSERIAL UNIQUE,
 ADD COLUMN created_at TIMESTAMPTZ,
 ADD COLUMN client_label TEXT NOT NULL DEFAULT 'Existing session';
