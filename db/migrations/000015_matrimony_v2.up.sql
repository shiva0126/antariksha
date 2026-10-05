-- Matrimony v2: structured biodata (in details JSONB), horoscope sharing,
-- shortlists, interest notes, contact sharing, selfie verification, email and
-- push alerts. Birth details stay server-side; only derived results are shown.

-- Birthplace completes the kundali (lagna, Mangal dosha). {name,lat,lon,tz}
ALTER TABLE member_accounts ADD COLUMN birth_place JSONB;
ALTER TABLE member_accounts ADD COLUMN email_alerts BOOLEAN NOT NULL DEFAULT true;

-- Horoscope results (guna score, Moon sign, nakshatra, Mangal dosha) are shown
-- only between two members who have both opted in.
ALTER TABLE matrimony_profiles ADD COLUMN horoscope_visible BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE matrimony_profiles ADD COLUMN verified_at TIMESTAMPTZ;
ALTER TABLE matrimony_profiles ADD COLUMN saved_search JSONB NOT NULL DEFAULT '{}';

ALTER TABLE matrimony_interests ADD COLUMN note TEXT NOT NULL DEFAULT '' CHECK (length(note) <= 300);
CREATE INDEX matrimony_interests_sender_time ON matrimony_interests(sender, created_at);

-- A member's own shortlist ("saved") and "not now" list.
CREATE TABLE matrimony_saved (
 owner TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 target TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 kind TEXT NOT NULL CHECK (kind IN ('saved','skipped')),
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 PRIMARY KEY (owner, target), CHECK (owner <> target)
);

-- Contact details one member chooses to share with one accepted match.
CREATE TABLE matrimony_contacts (
 owner TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 peer TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 contact TEXT NOT NULL CHECK (length(contact) BETWEEN 3 AND 200),
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 PRIMARY KEY (owner, peer), CHECK (owner <> peer)
);

-- Selfie checked by a moderator against published photos. Deleted once reviewed.
CREATE TABLE matrimony_verifications (
 account_id TEXT PRIMARY KEY REFERENCES member_accounts(id) ON DELETE CASCADE,
 photo BYTEA,
 status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected')),
 note TEXT NOT NULL DEFAULT '',
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 reviewed_at TIMESTAMPTZ,
 reviewer TEXT REFERENCES member_accounts(id) ON DELETE SET NULL
);

CREATE TABLE member_push_subscriptions (
 endpoint TEXT PRIMARY KEY CHECK (length(endpoint) <= 1000),
 account_id TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 p256dh TEXT NOT NULL, auth TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX member_push_account ON member_push_subscriptions(account_id);

CREATE TABLE app_secrets (name TEXT PRIMARY KEY, value TEXT NOT NULL);

-- Alerts: "accepted" notices, and delivery bookkeeping for email and push.
ALTER TABLE member_notifications DROP CONSTRAINT member_notifications_kind_check;
ALTER TABLE member_notifications ADD CONSTRAINT member_notifications_kind_check CHECK (kind IN ('follow','family','interest','message','accepted'));
ALTER TABLE member_notifications ADD COLUMN delivered_at TIMESTAMPTZ;
CREATE INDEX member_notifications_undelivered ON member_notifications(created_at) WHERE delivered_at IS NULL;

CREATE OR REPLACE FUNCTION record_member_notification() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE who TEXT; sender_id TEXT; event_kind TEXT; item TEXT;
BEGIN
 IF TG_OP='UPDATE' THEN
  -- Nested: NEW.status exists only on matrimony_interests.
  IF TG_TABLE_NAME<>'matrimony_interests' THEN RETURN NEW; END IF;
  IF NEW.status<>'accepted' OR OLD.status='accepted' THEN RETURN NEW; END IF;
  who:=NEW.sender; sender_id:=NEW.recipient; event_kind:='accepted'; item:=NEW.recipient;
 ELSIF TG_TABLE_NAME='member_follows' THEN
  IF NEW.accepted THEN RETURN NEW; END IF;
  who:=NEW.target; sender_id:=NEW.follower; event_kind:='follow'; item:=NEW.follower;
 ELSIF TG_TABLE_NAME='family_members' THEN
  IF NEW.accepted THEN RETURN NEW; END IF;
  who:=NEW.account_id; SELECT owner INTO sender_id FROM family_groups WHERE id=NEW.group_id;
  event_kind:='family'; item:=NEW.group_id;
 ELSIF TG_TABLE_NAME='matrimony_interests' THEN
  IF NEW.status<>'pending' THEN RETURN NEW; END IF;
  who:=NEW.recipient; sender_id:=NEW.sender; event_kind:='interest'; item:=NEW.sender;
 ELSE
  who:=NEW.recipient; sender_id:=NEW.sender; event_kind:='message'; item:=NEW.id::text;
 END IF;
 IF who<>sender_id THEN
  INSERT INTO member_notifications(recipient,actor,kind,subject)
  VALUES(who,sender_id,event_kind,item) ON CONFLICT(recipient,actor,kind,subject)
  DO UPDATE SET id=EXCLUDED.id,created_at=now(),read_at=NULL,dismissed=false,delivered_at=NULL;
 END IF;
 RETURN NEW;
END $$;

CREATE OR REPLACE FUNCTION notification_visible(n member_notifications,u TEXT) RETURNS BOOLEAN LANGUAGE SQL STABLE AS $$
 SELECT n.recipient=u AND NOT n.dismissed AND NOT member_blocked(n.actor,u) AND CASE n.kind
 WHEN 'family' THEN EXISTS(SELECT 1 FROM family_members m JOIN family_groups f ON f.id=m.group_id WHERE m.group_id=n.subject AND m.account_id=u AND NOT m.accepted AND f.owner=n.actor)
 WHEN 'follow' THEN EXISTS(SELECT 1 FROM member_follows f WHERE f.follower=n.actor AND f.target=u AND NOT f.accepted)
 WHEN 'interest' THEN EXISTS(SELECT 1 FROM matrimony_interests i WHERE i.sender=n.actor AND i.recipient=u AND i.status='pending')
 WHEN 'accepted' THEN EXISTS(SELECT 1 FROM matrimony_interests i WHERE i.sender=u AND i.recipient=n.actor AND i.status='accepted')
 WHEN 'message' THEN EXISTS(SELECT 1 FROM member_messages m WHERE m.id::text=n.subject AND m.sender=n.actor AND m.recipient=u) AND EXISTS(SELECT 1 FROM matrimony_interests i WHERE i.status='accepted' AND ((i.sender=n.actor AND i.recipient=u) OR (i.sender=u AND i.recipient=n.actor)))
 ELSE false END
 AND (n.kind='family' OR (SELECT count(*)=2 FROM member_settings s WHERE s.account_id IN(n.actor,u) AND s.community AND s.birth_date<=CURRENT_DATE-INTERVAL '18 years'))
$$;

-- Blocking removes shortlists and shared contacts for that pair.
CREATE FUNCTION revoke_blocked_matrimony() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 DELETE FROM matrimony_saved WHERE (owner=NEW.actor AND target=NEW.target) OR (owner=NEW.target AND target=NEW.actor);
 DELETE FROM matrimony_contacts WHERE (owner=NEW.actor AND peer=NEW.target) OR (owner=NEW.target AND peer=NEW.actor);
 RETURN NEW;
END $$;
CREATE TRIGGER revoke_blocked_matrimony AFTER INSERT ON member_blocks FOR EACH ROW EXECUTE FUNCTION revoke_blocked_matrimony();
