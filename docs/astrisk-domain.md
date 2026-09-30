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

## Hosting on this Windows/WSL system

The owner selected this PC as the host; a VPS is not required. It runs Ubuntu in
WSL2 NAT mode on Windows behind a home router. Port 80 is already occupied by
another application and must not be replaced. Use `deploy/Caddyfile.wsl` for
internal HTTP 8088 and HTTPS 8443, mapped from public standard ports 80/443.

1. `bash scripts/install-native-edge.sh` installs a pinned Caddy binary from its
   official release and checks its published cryptographic digest. It validates configuration
   but intentionally does not request certificates or start public serving.
2. Review `scripts/connect-wsl-edge.ps1`, then run it in elevated Windows
   PowerShell. It adds only dedicated 8088/8443 forwarding and Private-profile
   firewall rules. It fails on existing mappings; it never resets all portproxy
   rules. WSL address changes require manually updating these two mappings.
3. Reserve the PC's LAN address at the router. Forward WAN TCP 80 to PC TCP 8088
   and WAN TCP 443 to PC TCP 8443. Do not forward database or API port 3000.
   Compare the router WAN address with the current public egress address. If they
   differ due to carrier NAT, ask the ISP for incoming connectivity or configure
   an authenticated tunnel instead. No tunnel is installed or connected here.
4. Set the domain's A record to the current public IPv4, and www CNAME to the
   apex. Dynamic public IPs require a DNS update mechanism. Nameservers and email
   records can stay unchanged. The parked-domain address is not this PC.
5. Copy `deploy/astrisk-edge.service` into the user's systemd units after checking
   its paths. Enable it only once DNS/networking are ready. Apply the secure-cookie
   API override when switching to HTTPS; do not enable it while using HTTP login.
6. `systemctl --user daemon-reload` and
   `systemctl --user enable --now astrisk-edge.service` start the native proxy.
   Test from a separate internet connection, not just the same LAN.

Windows must remain awake with WSL running. A Linux user service cannot boot WSL
by itself. Configure Windows startup and user lingering only with administrator
access; neither is assumed to be enabled. Restart recovery for the native
PostgreSQL instance also needs verification before unattended/public use.

Microsoft WSL networking reference:
https://learn.microsoft.com/windows/wsl/networking
