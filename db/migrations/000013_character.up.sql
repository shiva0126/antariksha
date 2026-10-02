CREATE TABLE member_characters (
 account_id TEXT PRIMARY KEY REFERENCES member_accounts(id) ON DELETE CASCADE,
 profile JSONB NOT NULL CHECK(jsonb_typeof(profile)='object'),
 revision BIGINT NOT NULL DEFAULT 1,
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
