# Superadmin

## Account security and operations

Admin → Operations shows aggregate adoption, matching and review-backlog counts,
plus configuration readiness. Admin → Security offers opt-in authenticator setup.
Use your own authenticator app with the displayed manual key and confirm one code.
Setup expires after ten minutes; enabling/disabling signs out every device. Subsequent
sign-ins require the password and a fresh six-digit code. Recently used codes cannot
be reused, including immediately after setup: wait for the next code.

`ADMIN_TOTP_KEY_FILE` must name a regular, owner-readable-only file containing 32
random bytes. The service drop-in is `deploy/panchang-admin-security.conf`. Preserve
the key through restarts and back it up separately with encryption: database-only
backups are insufficient for restoring enrolled authenticators. Secrets in PostgreSQL
are AES-GCM encrypted with account-bound associated data. Missing/wrong keys fail
closed for enrolled accounts; they never bypass the authenticator.

Recovery keys and email password resets preserve the factor. A lost authenticator
requires verified server-operator assistance: identify the exact owner, transactionally
delete that owner's `member_totp` row and sessions, record an audit reason, then have
the owner enroll again. Do not expose a public bypass or use email-based role grants.
The operator password-recovery script does not reset MFA.

Me → Security now includes password change and per-session revocation. Changing a
password revokes all sessions and recovery credentials. Full account and provider
delivery work remaining is tracked in [the non-AI plan](non-ai-delivery-plan.md).

Migration 14 adds persisted `member`, `moderator` and `superadmin` roles and suspension state. Open `/#admin` for the administrator sign-in screen, using the existing account email and password. Once signed in, **Admin** is visible in the header on desktop and mobile, as well as in the profile menu. Only the role returned by the authenticated server enables that navigation; all `/api/admin/*` requests independently check the role on the server. A user cannot set their own role through registration or profile fields. The administrator sign-in screen uses the normal authentication endpoint; it does not grant privileges or create a separate password.

The page includes a paginated user list with literal email/handle search and status filtering, aggregate account counts, post/profile moderation queues, the reading library status and a paginated action log. User listings include account identifiers, email/verification status, role, join date, active-session count and opt-in flags. They do not expose passwords, phone ciphertext, private birth details, personal notes or messages.

Superadmins can suspend/reactivate ordinary or moderator accounts, revoke sessions, assign/remove moderator access and permanently delete an account. Every action requires the current superadmin password and a reason. Deletion additionally requires the exact target handle. Mutations and their audit records commit together. Failed attempts have a per-actor request budget; admin responses are not cached. Passwords are not logged or stored in the browser. Role changes, suspension, reactivation and explicit sign-out revoke sessions and pending email tokens. Suspension also pauses community and matrimony; reactivation does not republish them. Already-running requests are not automatically cancelled.

Superadmin accounts cannot be suspended, downgraded or deleted through these endpoints, including self-deletion. The owner still uses Community → Security for their own session revocation and recovery key. Changing or transferring superadmin ownership requires server access. There is no public bootstrap URL, default password or automatic privilege assignment based on an email string.

An operator grants access only to an identified existing account by running `scripts/grant-superadmin.sql` with `psql -v account_id=... -v ON_ERROR_STOP=1`; confirm that exactly one expected subject is returned. Server database credentials and account passwords must stay out of Git. The previous `MODERATOR_HANDLES` environment allowlist is replaced by database roles; migrate existing assignments before upgrading.

If the owner cannot sign in and email delivery is unavailable, a server operator can run `DATABASE_URL=... node scripts/recover-superadmin.mjs owner@example.com`. This requires exactly one existing, active superadmin; it never grants privileges. It replaces that account's recovery key and records the action. Instructions and a single-use key are saved in an owner-readable-only `.runtime/superadmin-recovery-*.txt` file, not printed. Use **Forgot password?** to choose a new password with the key. The key has no automatic expiry: use it promptly, then delete the local file. Resetting revokes sessions and pending email tokens, preserves account data and does not mark an unverified email as verified.

Deletion removes the account's relational data and owned conversations using the same cascade rules as self-deletion. An admin audit record retains the target ID/handle, action and reason. Local database backups can retain deleted data; off-device backup and retention policy remain operator responsibilities. The admin page does not display or export private conversations.

The reading library is retrieval, not model training: at implementation time, 108 mapped passages from one Brihat Jataka edition plus 339 original project entries are loaded (447 rows), with no vectors. Other book records represent acquisition plans, not ingested works. More approved sources can be added through the existing reviewed ingestion pipeline; the admin panel currently displays their status and does not run ingestion or paid embedding jobs.

Verified on 5 October 2026: all Go packages with PostgreSQL account integration enabled, API vet checks, frontend unit tests and TypeScript/Vite build passed. The admin browser test checks access denial, search, suspension, blocked sign-in, reactivation, audit history and mobile width. Existing matchmaking browser tests also passed. Tests run only against a disposable database; no real member was suspended or deleted during verification.
