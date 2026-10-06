import { test } from 'node:test';
import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { readFileSync, unlinkSync, statSync } from 'node:fs';
import { randomBytes } from 'node:crypto';
import { fileURLToPath } from 'node:url';
import { resolve } from 'node:path';

test('operator recovery preserves permissions and resets only the selected account', { skip: !process.env.ADMIN_RECOVERY_TEST_DATABASE_URL }, async () => {
  const dsn = process.env.ADMIN_RECOVERY_TEST_DATABASE_URL;
  const origin = process.env.ADMIN_RECOVERY_TEST_ORIGIN;
  assert.ok(new URL(dsn).pathname.startsWith('/astrisk_admin_recovery_test_'));
  assert.equal(new URL(origin).hostname, '127.0.0.1');
  assert.notEqual(new URL(origin).port, '3000');
  const root = fileURLToPath(new URL('../', import.meta.url));
  const email = `recovery_${randomBytes(8).toString('hex')}@example.com`;
  const password = 'original-test-password', next = 'replacement-test-password';
  const request = (path, data, cookie = '') => fetch(origin + path, { method: data ? 'POST' : 'GET', headers: { Origin: origin, 'Content-Type': 'application/json', Cookie: cookie }, ...(data ? { body: JSON.stringify(data) } : {}) });
  const registration = await request('/api/auth/register', { email, password, birth_date: '1996-01-01', birth_time: '10:15', consent: true });
  assert.equal(registration.status, 201);
  const cookie = registration.headers.get('set-cookie').split(';')[0];
  const account = await (await request('/api/me', undefined, cookie)).json();
  assert.match(account.id, /^[a-f0-9]{64}$/);
  const sql = statement => execFileSync(resolve(root, '.runtime/pgdist/usr/lib/postgresql/16/bin/psql'), [dsn, '-XAt', '-v', 'ON_ERROR_STOP=1', '-c', statement], { encoding: 'utf8', stdio: ['pipe', 'pipe', 'pipe'] }).trim();
  const issue = mail => execFileSync(process.execPath, [resolve(root, 'scripts/recover-superadmin.mjs'), mail], { env: { ...process.env, DATABASE_URL: dsn }, encoding: 'utf8', stdio: ['pipe', 'pipe', 'pipe'] });
  let file;
  try {
    assert.throws(() => issue(email), 'ordinary member cannot be bootstrapped by recovery');
    assert.equal(sql(`SELECT recovery_hash IS NULL FROM member_accounts WHERE id='${account.id}'`), 't');
    sql(`UPDATE member_accounts SET role='superadmin' WHERE id='${account.id}'`);
    const output = issue(email);
    file = output.match(/saved privately to (.+\.txt)\./)[1];
    assert.equal(statSync(file).mode & 0o777, 0o600);
    const key = readFileSync(file, 'utf8').match(/Recovery key: ([a-f0-9]{64})/)[1];
    assert.ok(!output.includes(key));
    assert.equal(sql(`SELECT count(*) FROM member_audit WHERE subject='${account.id}' AND action='admin_recovery_issued'`), '1');
    assert.equal((await request('/api/auth/recover', { handle: email, key, password: next })).status, 200);
    assert.equal((await request('/api/me', undefined, cookie)).status, 401);
    assert.equal((await request('/api/auth/recover', { handle: email, key, password: next })).status, 401);
    assert.equal((await request('/api/auth/login', { email, password })).status, 401);
    const login = await request('/api/auth/login', { email, password: next });
    assert.equal(login.status, 200);
    const newCookie = login.headers.get('set-cookie').split(';')[0];
    assert.equal((await request('/api/admin/summary', undefined, newCookie)).status, 200);
    assert.equal(sql(`SELECT role || ':' || (email_verified_at IS NULL)::text || ':' || (recovery_hash IS NULL)::text FROM member_accounts WHERE id='${account.id}'`), 'superadmin:true:true');
  } finally {
    if (file) unlinkSync(file);
    sql(`DELETE FROM member_audit WHERE subject='${account.id}'; DELETE FROM member_accounts WHERE id='${account.id}'`);
  }
});
