# Non-AI delivery plan — 6 October 2026

Scope: user delegated AI training/model work to Claude. Do not change training,
model service configuration, corpus content or weights as part of this release.
This is a staged plan, not a claim that the whole competitor roadmap is finished.

## Implemented in this release

- Self-service password change in Me → Security. Current password required;
  sessions, recovery key and pending email tokens invalidated atomically.
- Active session list with coarse browser/device labels, current-session marker,
  creation/expiry times, and password-confirmed per-session revocation. Older
  sessions truthfully show unknown creation dates. Session token hashes are never
  returned. Browser labels are self-reported, not verified device identities.
- Administrator authenticator enrollment in Admin → Security. RFC 6238 TOTP,
  30-second steps, six digits, ±one step, single-use counters, rate limits,
  encrypted per-account secrets, ten-minute pending enrollment and audit events.
  Confirming or disabling revokes every session. Login requires the second factor
  once enabled. Password recovery deliberately does not disable it.
- Admin → Operations: signup cohorts, verified emails, accounts with unexpired
  sessions, active matrimony profiles, missing birthplaces, interest creation and
  acceptance cohort, old reports, pending photo-review age and provider readiness.
  These are not DAU, marriage-success probabilities or a private-message viewer.
- No horoscope calculation for matrimony without both confirmed birthplaces and
  timezones. Removed the fallback that assumed Asia/Kolkata. Discovery remains
  available without horoscope results.
- Photo-review badges say what was checked. Changes to a published photo set
  invalidate review; captions/private uploads alone do not. No claim of verified
  income, employment, identity documents or character.
- Email verification and unavailable-provider status visible in Security.

## Setup dependencies (not completed by writing code)

- SMTP email delivery needs a configured sending provider and credentials. Until
  then no verification/reset emails can be delivered. Do not fake a successful send.
- Phone verification requires the SMS provider configuration and encryption key.
- The owner must enroll their own authenticator; never enroll one on their behalf.
- Securely back up the authenticator encryption key separately from the database.
  Existing database-only backups do not include this key. Loss of it prevents MFA
  sign-in until a server operator performs audited factor recovery.
- Off-device backup destinations and notification destinations need operator setup.

## Next non-AI increments

1. Recovery and reliability: login/security alerts, durable abuse-rate limits,
   recovery-code UX, delivery monitoring, independent security review and a tested
   off-device recovery drill. Keep account recovery workable without bypassing MFA.
2. Accuracy: unknown/approximate birth-time handling, source/version selection for
   differing yoga conventions, regional festival and muhurta comparisons. Coordinate
   with Claude before any change that affects interpretation facts.
3. Matrimony: explicit flexible vs essential preferences, profile freshness, more
   conversation prompts, family-prepared labels and moderation assignment/appeals.
   Existing profile explanations and family helpers should be extended, not rebuilt.
4. Community: opt-in topic discovery and interest groups, field-level visibility,
   consent-based provider imports with preview/revocation. OAuth access is not ID
   verification; no automatic social-account attachment based on names.
5. Later: safe video introductions, stories, expert review and native applications.
   Define operating/moderation costs and safety controls before implementing these.

## Validation

Account/authorization/database tests use a disposable `astrisk_delivery_test_*`
database, not production members. New tests cover password/session ownership,
stale credential races, RFC TOTP vectors, replay rejection, preservation of MFA
through recovery, birth-timezone requirements and review invalidation. Browser
tests cover password-change/sign-in, session invalidation and admin access at
mobile/tablet widths. AI/training evaluations are outside this release.
