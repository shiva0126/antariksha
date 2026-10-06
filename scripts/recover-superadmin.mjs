// Server-operator-only recovery. Never creates accounts or grants roles.
// DATABASE_URL=... node scripts/recover-superadmin.mjs owner@example.com
import { randomBytes, createHash } from 'node:crypto';
import { execFileSync } from 'node:child_process';
import { openSync, writeFileSync, closeSync, unlinkSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { resolve } from 'node:path';

const root = fileURLToPath(new URL('../', import.meta.url));
const email = (process.argv[2] || '').trim().toLowerCase();
if (!process.env.DATABASE_URL || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
  console.error('Set DATABASE_URL and supply the exact existing superadmin email.');
  process.exit(1);
}
const key = randomBytes(32).toString('hex');
const hash = createHash('sha256').update(key).digest('hex');
const output = resolve(root, '.runtime', `superadmin-recovery-${Date.now()}.txt`);
// Exclusive, owner-only file. The secret is never printed or placed in Git.
const fd = openSync(output, 'wx', 0o600);
try {
  writeFileSync(fd, `Astrisk administrator recovery\n\nOpen https://astrisk.space/#admin and click Forgot password?\nEmail: ${email}\nChoose your own NEW password (12–72 bytes).\nRecovery key: ${key}\n\nThis key works once and has no automatic expiry. Keep it private, use it now, and delete this file afterwards.\nRecovery signs out existing sessions and preserves your role and account data.\nAfter resetting, sign in at https://astrisk.space/#admin with your new password.\n`);
} finally { closeSync(fd); }
try {
  execFileSync(process.env.PSQL || resolve(root, '.runtime/pgdist/usr/lib/postgresql/16/bin/psql'),
    [process.env.DATABASE_URL, '-X', '-v', 'ON_ERROR_STOP=1', '-v', `account_email=${email}`, '-v', `recovery_hash=${hash}`], {
      stdio: ['pipe', 'pipe', 'pipe'],
      input: `BEGIN;
CREATE TEMP TABLE recovery_target ON COMMIT DROP AS
 SELECT id FROM member_accounts WHERE lower(email)=:'account_email' AND role='superadmin' AND NOT suspended FOR UPDATE;
DO $$ BEGIN
 IF (SELECT count(*) FROM recovery_target) <> 1 THEN
  RAISE EXCEPTION 'Expected exactly one active superadmin';
 END IF;
END $$;
UPDATE member_accounts SET recovery_hash=:'recovery_hash' WHERE id IN (SELECT id FROM recovery_target);
INSERT INTO member_audit(actor,action,subject,details)
 SELECT NULL,'admin_recovery_issued',id,'{"reason":"Server operator issued one-use recovery key at owner request"}'::jsonb FROM recovery_target;
COMMIT;
`,
    });
} catch {
  unlinkSync(output);
  console.error('Recovery was not issued. Confirm database access and the exact active superadmin email.');
  process.exit(1);
}
console.log(`Recovery instructions saved privately to ${output}. Password is unchanged until the key is used.`);
