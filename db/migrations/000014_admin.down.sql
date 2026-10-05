CREATE OR REPLACE FUNCTION member_blocked(a TEXT,b TEXT) RETURNS BOOLEAN LANGUAGE SQL STABLE AS $$
 SELECT EXISTS(SELECT 1 FROM member_blocks WHERE (actor=a AND target=b) OR (actor=b AND target=a))
$$;
ALTER TABLE member_audit DROP COLUMN details;
DROP INDEX member_accounts_admin_order;
ALTER TABLE member_accounts DROP CONSTRAINT superadmin_active, DROP COLUMN role, DROP COLUMN suspended;
