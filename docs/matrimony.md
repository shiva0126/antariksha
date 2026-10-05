# Matrimony (v2)

Matrimony is its own section (`#matrimony`): **Discover · Interests · My matrimony
profile · Family · Safety**. It needs community opt-in with an adult date of
birth. Everything is free.

## What members get

| Feature | How it works |
| --- | --- |
| Horoscope matching | Opt-in per member ("Show horoscope matching"). Between two members who have **both** opted in, each card shows the guna score out of 36, Moon sign and nakshatra, Mangal dosha and the root-number graha. Computed on the server from stored birth details; date, time and place are never returned. Without a birthplace the guna score still works (it uses the Moon) and Mangal dosha shows as unknown. |
| Full comparison | "Compare horoscopes" opens all eight kootas, doshas and cancellations, Mangal dosha, numerology and **Ask Astrisk** for that pair (`/api/matrimony/compare/{peer}` and `/chat`). Not stored. |
| Structured biodata | Religion, community (free text, optional), mother tongue, diet, height, marital status, education level and field, occupation category and title, income band, family type, timeline, relocation, children. City from GeoNames. Values are validated against `matrimonyEnums` in `api/matrimony.go`. Completeness meter. |
| Discovery | Filters (looking for, age, city, religion, mother tongue, diet, marital status, education, minimum guna, Mangal dosha, verified), sort (best match, guna, newest), 12 per page, saved search, shortlist and "Not now". Every card lists why it was suggested. |
| Interests | Optional 300-character note; at most 20 interests a day. Received / connected / sent lists. Accepting notifies the sender. |
| Chat | Opens after mutual acceptance, refreshes every 10 s while visible. "Share my contact" reveals a phone or email to that one person only; blocking removes it. |
| Verification | Selfie + at least one published photo. A moderator compares them in Community → Moderation; the selfie is deleted after review and only the badge remains. |
| Biodata PDF | `#biodata`: printable biodata with an optional kundali page (South and North Indian charts and planet table), saved through the browser's print dialog. |
| Family | Family-prepared biodata and 30-day shortlist helpers (unchanged), now under the Family tab; helpers also see guna scores and reasons. |
| Safety | Guide with helplines (112, 181, 1930) and report reasons. |

## Alerts

`api/alerts.go` delivers each notification once, 90 seconds after it is created
(so an online member who reads it in the app is not emailed):

- **Email** to verified addresses with alerts on, when SMTP is configured
  (`SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM`,
  `PUBLIC_ORIGIN=https://astrisk.space`; see `deploy/email.env.example`). Free
  tiers of Brevo or Zoho Mail work. Without SMTP nothing is sent.
- **Web push** to devices where the member pressed "Turn on notifications".
  VAPID keys are generated once and stored in `app_secrets`. Only the browser
  vendors' push services are contacted. iPhones need Astrisk added to the Home
  Screen first (the app is installable: `manifest.webmanifest`, `sw.js`).

Alerts never include message text.

## Operations

- Migration 15 (`db/migrations/000015_matrimony_v2.up.sql`); `scripts/migrate.sh`
  applies any pending migration and `start-native.sh` calls it.
- `TRUSTED_PROXY_HEADER=CF-Connecting-IP` makes rate limits per visitor behind
  the Cloudflare tunnel (honoured only on loopback connections). Without it all
  visitors share one limit.
- `astrisk-health.timer` runs `scripts/healthcheck.sh` every 5 minutes and
  restarts the API, embedding model or tunnel if they stop answering. Set
  `ALERT_URL` (for example a free `https://ntfy.sh/<topic>`) to be told.
- `scripts/offsite-backup.sh` encrypts the newest backup with GPG and copies it
  to `RCLONE_REMOTE` (Google Drive, OneDrive, B2 … after `rclone config`) or
  `OFFSITE_DIR`. Needs `BACKUP_PASSPHRASE_FILE`; nothing is uploaded unencrypted.
- Browser tests: `scripts/test-browser.sh` runs Playwright against a throwaway
  API on port 3100 and the disposable `panchang_test` database
  (`scripts/test-db.sh`), never the live site.
