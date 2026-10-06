DELETE FROM member_notifications WHERE kind='transit';
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
ALTER TABLE member_notifications DROP CONSTRAINT member_notifications_kind_check;
ALTER TABLE member_notifications ADD CONSTRAINT member_notifications_kind_check CHECK (kind IN ('follow','family','interest','message','accepted'));
ALTER TABLE member_accounts DROP COLUMN transit_alerts;
