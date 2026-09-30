ALTER TABLE member_accounts ADD COLUMN email TEXT;
ALTER TABLE member_accounts ADD COLUMN birth_date DATE;
ALTER TABLE member_accounts ADD COLUMN birth_time TIME;
CREATE UNIQUE INDEX member_accounts_email ON member_accounts(lower(email)) WHERE email IS NOT NULL;
ALTER TABLE member_accounts ADD CONSTRAINT member_email_normalized CHECK(email IS NULL OR (email=lower(btrim(email)) AND length(email)<=254));
