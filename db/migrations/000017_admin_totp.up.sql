CREATE TABLE member_totp (
 account_id TEXT PRIMARY KEY REFERENCES member_accounts(id) ON DELETE CASCADE,
 secret BYTEA NOT NULL,
 enabled BOOLEAN NOT NULL DEFAULT false,
 pending_until TIMESTAMPTZ NOT NULL DEFAULT now()+interval '10 minutes',
 last_counter BIGINT NOT NULL DEFAULT -1
);
