CREATE TABLE member_accounts (
 id TEXT PRIMARY KEY,
 handle TEXT UNIQUE NOT NULL,
 password_hash TEXT NOT NULL,
 profile JSONB NOT NULL DEFAULT '{}',
 consent_version TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE member_sessions (
 token_hash TEXT PRIMARY KEY,
 account_id TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 expires_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX member_sessions_account ON member_sessions(account_id);
