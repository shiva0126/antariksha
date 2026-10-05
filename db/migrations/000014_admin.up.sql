ALTER TABLE member_accounts
 ADD COLUMN role TEXT NOT NULL DEFAULT 'member' CHECK(role IN ('member','moderator','superadmin')),
 ADD COLUMN suspended BOOLEAN NOT NULL DEFAULT false,
 ADD CONSTRAINT superadmin_active CHECK(role<>'superadmin' OR NOT suspended);
CREATE INDEX member_accounts_admin_order ON member_accounts(created_at DESC,id DESC);
ALTER TABLE member_audit ADD COLUMN details JSONB NOT NULL DEFAULT '{}';
CREATE OR REPLACE FUNCTION member_blocked(a TEXT,b TEXT) RETURNS BOOLEAN LANGUAGE SQL STABLE AS $$
 SELECT EXISTS(SELECT 1 FROM member_blocks WHERE (actor=a AND target=b) OR (actor=b AND target=a))
 OR EXISTS(SELECT 1 FROM member_accounts WHERE id IN(a,b) AND suspended)
$$;
