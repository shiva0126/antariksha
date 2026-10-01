CREATE TABLE member_chart_store (
 account_id TEXT PRIMARY KEY REFERENCES member_accounts(id) ON DELETE CASCADE,
 profiles JSONB NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(profiles)='array'),
 revision BIGINT NOT NULL DEFAULT 0,
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE member_accounts ADD COLUMN email_verified_at TIMESTAMPTZ;
CREATE TABLE member_email_tokens (
 account_id TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 purpose TEXT NOT NULL CHECK (purpose IN ('verify','reset')),
 token_hash TEXT NOT NULL UNIQUE,
 email TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 expires_at TIMESTAMPTZ NOT NULL,
 PRIMARY KEY(account_id,purpose)
);
CREATE INDEX member_email_token_expiry ON member_email_tokens(expires_at);
