CREATE TABLE matrimony_delegates (
 owner TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 delegate TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 family_id TEXT NOT NULL REFERENCES family_groups(id) ON DELETE CASCADE,
 accepted BOOLEAN NOT NULL DEFAULT false,
 expires_at TIMESTAMPTZ NOT NULL DEFAULT now()+INTERVAL '30 days',
 PRIMARY KEY(owner,delegate), CHECK(owner<>delegate)
);
CREATE TABLE matrimony_shortlist (
 owner TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 delegate TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 candidate TEXT REFERENCES member_accounts(id) ON DELETE CASCADE,
 note TEXT NOT NULL DEFAULT '',
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 PRIMARY KEY(owner,delegate,candidate),
 FOREIGN KEY(owner,delegate) REFERENCES matrimony_delegates(owner,delegate) ON DELETE CASCADE,
 CHECK(candidate<>owner AND candidate<>delegate)
);
CREATE FUNCTION delegate_allowed(o TEXT,d TEXT) RETURNS BOOLEAN LANGUAGE SQL STABLE AS $$
 SELECT EXISTS(SELECT 1 FROM matrimony_delegates g JOIN matrimony_profiles p ON p.account_id=o
 WHERE g.owner=o AND g.delegate=d AND g.accepted AND g.expires_at>now() AND p.active
 AND family_access(g.family_id,o) AND family_access(g.family_id,d) AND NOT member_blocked(o,d)
 AND (SELECT count(*)=2 FROM member_settings s WHERE s.account_id IN(o,d) AND s.community AND s.birth_date<=CURRENT_DATE-INTERVAL '18 years'))
$$;

-- Leaving a group or blocking a helper permanently revokes the grant. Rejoining
-- or unblocking must never silently restore delegated access.
CREATE FUNCTION revoke_family_grants() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF TG_TABLE_NAME='family_members' THEN
  DELETE FROM matrimony_delegates WHERE family_id=OLD.group_id AND (owner=OLD.account_id OR delegate=OLD.account_id);
  RETURN OLD;
 ELSE
  DELETE FROM matrimony_delegates WHERE (owner=NEW.actor AND delegate=NEW.target) OR (owner=NEW.target AND delegate=NEW.actor);
  RETURN NEW;
 END IF;
END $$;
CREATE TRIGGER revoke_departed_helper AFTER DELETE ON family_members FOR EACH ROW EXECUTE FUNCTION revoke_family_grants();
CREATE TRIGGER revoke_blocked_helper AFTER INSERT ON member_blocks FOR EACH ROW EXECUTE FUNCTION revoke_family_grants();
