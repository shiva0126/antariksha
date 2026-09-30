# Astrisk — astrisk.space

The product is now Astrisk. The existing GitHub repository remains
`shiva0126/antariksha`; service names, database identifiers, session cookies and
browser storage keys intentionally remain unchanged so rebranding does not lose
accounts or saved charts. Moving to a new origin still requires sign-in, and old
origin-local browser charts do not move automatically.

## Current state (2026-09-30)

`astrisk.space` resolved to `2.57.91.91` and returned Hostinger's parked-domain
page, not this application. The local native application runs on port 3000 on a
private-address machine. Buying the domain has not provisioned a public Go/Postgres
host. No DNS records, nameservers, ownership details or renewal settings have been
changed. No registrar contact data belongs in the source repository.

## DNS cutover — requires a public hosting target

Keep the existing nameservers unless deliberately migrating DNS providers. At the
current DNS provider, replace only the parked web records for the following names:

| Type | Name | Value |
| --- | --- | --- |
| A | @ | The chosen server's public IPv4 address (not its private address) |
| CNAME | www | astrisk.space |

Do not use the parking IP as the application IP. Publish AAAA only if IPv6 reaches
the same server; remove stale web AAAA records during cutover. Preserve mail MX,
SPF, DKIM, DMARC and verification records. Review the existing zone before changes.
The public server's address and authorized DNS access are still needed.

## Native HTTPS host (no Docker)

1. Provision or select a Linux server capable of running the native Go/cgo engine
   and PostgreSQL. Back up the existing database before any migration; do not copy
   live database files as a backup. Transfer data only to a user-approved host.
2. Deploy this repository and native dependencies. Configure PostgreSQL, ephemeris
   files and protected service environment; migrate through version 10. Verify the
   app locally on that server before DNS changes.
3. Bind the API to `127.0.0.1:3000` and set `COOKIE_SECURE=true`. The example
   `deploy/astrisk-production.conf` is a systemd override for the existing service;
   apply it on the public host, not the current HTTP-only development installation.
4. Install Caddy and review `deploy/Caddyfile`. Allow inbound TCP 80/443; keep
   PostgreSQL, the API port and Caddy's admin interface private. Configure DNS as
   above, then validate the Caddyfile before starting/reloading Caddy.
5. Check HTTPS, apex/www redirects, login cookies, same-origin POSTs, media access,
   `/healthz`, and account isolation. Confirm HTTP redirects to HTTPS. Do not claim
   the domain is live until it serves Astrisk rather than the parking page.

Caddy handles certificate issuance/renewal when DNS and reachability are correct:
https://caddyserver.com/docs/quick-starts/https
https://caddyserver.com/docs/automatic-https

These files are deployment templates, not evidence of completed public hosting.
The public launch safety and provider checklist remains in community-operations.md.
