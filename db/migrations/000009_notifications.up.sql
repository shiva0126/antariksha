CREATE TABLE member_notifications (
 id BIGSERIAL PRIMARY KEY,
 recipient TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 actor TEXT NOT NULL REFERENCES member_accounts(id) ON DELETE CASCADE,
 kind TEXT NOT NULL CHECK(kind IN ('follow','family','interest','message')),
 subject TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 read_at TIMESTAMPTZ,
 dismissed BOOLEAN NOT NULL DEFAULT false,
 UNIQUE(recipient,actor,kind,subject)
);
CREATE INDEX member_notifications_inbox ON member_notifications(recipient,id DESC);

-- Events are recorded in the same transaction as their source. No message text
-- or family details are copied into notification storage.
CREATE OR REPLACE FUNCTION record_member_notification() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE who TEXT; sender_id TEXT; event_kind TEXT; item TEXT;
BEGIN
 IF TG_OP='UPDATE' THEN RETURN NEW; END IF;
 IF TG_TABLE_NAME='member_follows' THEN
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
  DO UPDATE SET id=EXCLUDED.id,created_at=now(),read_at=NULL,dismissed=false;
 END IF;
 RETURN NEW;
END $$;
CREATE TRIGGER notify_follow AFTER INSERT OR UPDATE ON member_follows FOR EACH ROW EXECUTE FUNCTION record_member_notification();
CREATE TRIGGER notify_family AFTER INSERT OR UPDATE ON family_members FOR EACH ROW EXECUTE FUNCTION record_member_notification();
CREATE TRIGGER notify_interest AFTER INSERT OR UPDATE ON matrimony_interests FOR EACH ROW EXECUTE FUNCTION record_member_notification();
CREATE TRIGGER notify_message AFTER INSERT ON member_messages FOR EACH ROW EXECUTE FUNCTION record_member_notification();

CREATE FUNCTION notification_visible(n member_notifications,u TEXT) RETURNS BOOLEAN LANGUAGE SQL STABLE AS $$
 SELECT n.recipient=u AND NOT n.dismissed AND NOT member_blocked(n.actor,u) AND CASE n.kind
 WHEN 'family' THEN EXISTS(SELECT 1 FROM family_members m JOIN family_groups f ON f.id=m.group_id WHERE m.group_id=n.subject AND m.account_id=u AND NOT m.accepted AND f.owner=n.actor)
 WHEN 'follow' THEN EXISTS(SELECT 1 FROM member_follows f WHERE f.follower=n.actor AND f.target=u AND NOT f.accepted)
 WHEN 'interest' THEN EXISTS(SELECT 1 FROM matrimony_interests i WHERE i.sender=n.actor AND i.recipient=u AND i.status='pending')
 WHEN 'message' THEN EXISTS(SELECT 1 FROM member_messages m WHERE m.id::text=n.subject AND m.sender=n.actor AND m.recipient=u) AND EXISTS(SELECT 1 FROM matrimony_interests i WHERE i.status='accepted' AND ((i.sender=n.actor AND i.recipient=u) OR (i.sender=u AND i.recipient=n.actor)))
 ELSE false END
 AND (n.kind='family' OR (SELECT count(*)=2 FROM member_settings s WHERE s.account_id IN(n.actor,u) AND s.community AND s.birth_date<=CURRENT_DATE-INTERVAL '18 years'))
$$;
