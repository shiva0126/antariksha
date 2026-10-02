ALTER TABLE matrimony_profiles ADD COLUMN hidden BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE matrimony_drafts (
 id TEXT PRIMARY KEY,
 creator TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 recipient TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 relationship TEXT NOT NULL CHECK (relationship IN ('parent','sibling','relative','friend')),
 details JSONB NOT NULL,
 status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','pending')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 CHECK (creator<>recipient),
 CHECK ((status='draft' AND recipient IS NULL) OR (status='pending' AND recipient IS NOT NULL))
);
CREATE INDEX matrimony_drafts_recipient ON matrimony_drafts(recipient);
CREATE INDEX matrimony_drafts_creator ON matrimony_drafts(creator);

CREATE TABLE matrimony_photos (
 id TEXT PRIMARY KEY,
 owner TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 draft_id TEXT REFERENCES matrimony_drafts(id) ON DELETE CASCADE,
 data BYTEA NOT NULL,
 alt TEXT NOT NULL DEFAULT '',
 published BOOLEAN NOT NULL DEFAULT false,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 CHECK (draft_id IS NULL OR NOT published)
);
CREATE INDEX matrimony_photos_owner ON matrimony_photos(owner);
CREATE INDEX matrimony_photos_draft ON matrimony_photos(draft_id);

CREATE TABLE matrimony_reports (
 id BIGSERIAL PRIMARY KEY,
 reporter TEXT REFERENCES member_accounts(id) ON DELETE SET NULL,
 subject TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 reason TEXT NOT NULL,
 status TEXT NOT NULL DEFAULT 'open' CHECK(status IN ('open','dismissed','hidden')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- The same eligibility check protects both profile cards and photo delivery.
CREATE FUNCTION matrimony_visible(subject TEXT,viewer TEXT) RETURNS BOOLEAN LANGUAGE SQL STABLE AS $$
 SELECT EXISTS(
 SELECT 1 FROM matrimony_profiles p JOIN member_settings c ON c.account_id=p.account_id
 JOIN matrimony_profiles mine ON mine.account_id=viewer JOIN member_settings me ON me.account_id=viewer
 WHERE p.account_id=subject AND subject<>viewer AND p.active AND mine.active AND NOT p.hidden AND NOT mine.hidden
 AND c.community AND me.community AND NOT member_blocked(subject,viewer)
 AND c.birth_date<=CURRENT_DATE-INTERVAL '18 years' AND me.birth_date<=CURRENT_DATE-INTERVAL '18 years'
 AND date_part('year',age(c.birth_date)) BETWEEN (mine.details->>'min_age')::int AND (mine.details->>'max_age')::int
 AND date_part('year',age(me.birth_date)) BETWEEN (p.details->>'min_age')::int AND (p.details->>'max_age')::int
 )
$$;

CREATE OR REPLACE FUNCTION delegate_allowed(o TEXT,d TEXT) RETURNS BOOLEAN LANGUAGE SQL STABLE AS $$
 SELECT EXISTS(SELECT 1 FROM matrimony_delegates g JOIN matrimony_profiles p ON p.account_id=o
 WHERE g.owner=o AND g.delegate=d AND g.accepted AND g.expires_at>now() AND p.active AND NOT p.hidden
 AND family_access(g.family_id,o) AND family_access(g.family_id,d) AND NOT member_blocked(o,d)
 AND (SELECT count(*)=2 FROM member_settings s WHERE s.account_id IN(o,d) AND s.community AND s.birth_date<=CURRENT_DATE-INTERVAL '18 years'))
$$;

CREATE FUNCTION revoke_blocked_biodata() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 DELETE FROM matrimony_drafts WHERE (creator=NEW.actor AND recipient=NEW.target) OR (creator=NEW.target AND recipient=NEW.actor);
 RETURN NEW;
END $$;
CREATE TRIGGER revoke_blocked_biodata AFTER INSERT ON member_blocks FOR EACH ROW EXECUTE FUNCTION revoke_blocked_biodata();
