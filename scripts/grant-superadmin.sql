-- Server-operator-only bootstrap, using the exact existing account ID:
-- psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -v account_id=... -f scripts/grant-superadmin.sql
-- Confirm the returned subject matches the intended existing account.
BEGIN;
WITH changed AS (
 UPDATE member_accounts SET role='superadmin' WHERE id=:'account_id' AND NOT suspended RETURNING id
)
INSERT INTO member_audit(actor,action,subject,details)
 SELECT NULL,'admin_bootstrap',id,'{"reason":"Superadmin granted by server operator"}'::jsonb FROM changed
 RETURNING subject;
COMMIT;
