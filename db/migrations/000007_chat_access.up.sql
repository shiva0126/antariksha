CREATE TABLE chat_guest_owners (
 session_id TEXT PRIMARY KEY REFERENCES chat_sessions(id) ON DELETE CASCADE,
 token_hash TEXT NOT NULL
);
CREATE INDEX chat_guest_owners_token ON chat_guest_owners(token_hash);
CREATE TABLE member_sms_budget (
 account_id TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 day DATE NOT NULL DEFAULT CURRENT_DATE,
 sends INTEGER NOT NULL DEFAULT 0,
 PRIMARY KEY(account_id,day)
);
