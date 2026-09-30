# Community validation build

The native application runs on port 3000 via `panchang.service`. No Docker is used.
Start from a fresh checkout with `scripts/start-native.sh`; migrations 6–10 add
community tables, conversation ownership, email/birth onboarding, notifications
and family-assisted shortlisting. Back up the
database before upgrading. New API requests require an authenticated session;
`/healthz` remains public. Existing clients must register/login first.

## User flow

Opening the app shows email/password sign-in. Create account asks for email,
password (12–72 bytes), DOB and birth time. Birth place is selected when making a
chart; birth time is local, not sufficient by itself to calculate a chart.
Email ownership is not yet verified. Users with pre-email accounts can use the
explicit legacy-handle option. Hobbies and personal details live in My account;
community sharing and private reflection prompts live under Profile & interests.
Community and matrimony are separate opt-ins. DOB is self-declared, not verified ID.

Private posts are visible only to the owner; follower posts need approved follows;
family posts need an accepted group membership; community posts need community
opt-in. All media reads use the same authorization checks. JPEG/PNG uploads are
decoded and re-encoded (no EXIF), bounded to 8 MB/16 MP, and capped at 100 MB per
account. This is a database-backed MVP, not an object-storage/CDN media service.

Family records require permission to share, or unnamed placeholders. Creating a
relative does not create an account or enroll them in matrimony. Match suggestions
explain shared interests and preferences; no character or marriage-success score
is inferred from astrology. Messages require mutual acceptance and respect blocks.

## Operator configuration

Community → Family assistance allows an active adult matrimony member to invite
an adult member of a shared private family group to help shortlist. The helper must
accept. Grants expire after 30 days; renewing requires acceptance again. Helpers see
only eligible published candidates and cannot act as the owner. Notes are private
to the owner and helper and included in their exports, never sent to candidates.
Owners can revoke at any time; leaving the shared group or blocking either way
permanently deletes the grant and shortlist. Pausing the owner's matrimony profile
suspends access. Expiry and reciprocal age eligibility are checked at read/write.
Existing copies or screenshots cannot be recalled. Notifications for helper grants
are not yet emitted; pending grants appear in the Family assistance tab.

Keep values in the protected service environment, never Git:

- `COOKIE_SECURE=true` when served over HTTPS behind a trusted proxy.
- `MODERATOR_HANDLES`: comma-separated existing account handles for the report queue.
- `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, `TWILIO_VERIFY_SERVICE_SID`:
  a configured Twilio Verify service for optional phone linking.
- `CONTACT_ENCRYPTION_KEY`: 64 hexadecimal characters (32 random bytes); back up
  securely. Losing or replacing it without a migration makes saved numbers unreadable.

Phone linking requires password reauthentication, consent, provider approval,
cooldown, attempt limits and a daily send budget. An absent configuration disables
the UI/API flow; no fake verification is returned. Phone access does not verify
identity, age or suitability. SMS cost, email verification, moderators, retention,
backups, abuse response and production TLS remain operator responsibilities.

Recovery uses a one-time key generated after password reauthentication. It revokes
all sessions. No email reset messages are sent. Store the key securely.

## Validation

Run `go test ./...`; supply `ACCOUNT_TEST_DATABASE_URL` to also run PostgreSQL-backed
account isolation, media access, family permission, matching, blocking, moderation
and recovery tests. `npm run build` in `web` type-checks and builds the UI.
Run `npx playwright test` in `web` against the running port-3000 service for browser
login, chart, Panchang, language, privacy and community flows. Tests create and
delete their own accounts; do not use a production database for test traffic.

The astronomy route unit tests exercise the inner router deliberately; a separate
authentication-boundary regression test ensures the public handler rejects guests.
Legacy unowned chats remain inaccessible; possession of a session ID is not proof
of ownership. Browser-saved legacy charts are not automatically imported. My account
has a preview/select/consent importer which preserves the original records and
imports only validated chart inputs, never old chat IDs.

The Community Notifications tab is an in-app inbox, not browser push, email or SMS.
Database triggers commit source events and notifications atomically. Reads exclude
blocked actors and revoked invitations; accepting a pending request removes its
actionable notice. Message notices require an active mutual connection and adult
community eligibility. No message text is copied. New source events are captured
from migration 9 onward; old events are not backfilled. Notifications are included
in account export and cascade on account deletion. No background delivery worker
or third-party notification provider is needed for this inbox.

Known limitations and remaining roadmap are tracked in `community-matrimony-plan.md`.
Do not publish this validation build as fully audited or claim unavailable providers
are operational. Do not scrape or infer hidden social accounts from contact details.
