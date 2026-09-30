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

The owner selected this PC as the host; a VPS is not required for LAN testing. It
runs Ubuntu in WSL2 mirrored-network mode on Windows behind a home router. Port 80
is already occupied by another application and must not be replaced. Use
`deploy/Caddyfile.wsl` for HTTP 8088 and HTTPS 8443.

1. `bash scripts/install-native-edge.sh` installs a pinned Caddy binary from its
   official release and checks its published cryptographic digest. It validates configuration
   but intentionally does not request certificates or start public serving.
2. WSL mirrored networking is enabled in the Windows user's `.wslconfig`, so the
   Linux interface uses the reserved LAN address directly (`192.168.1.3`). The
   Windows portproxy script remains available for NAT-mode machines, but is not
   needed in mirrored mode.
3. The router has a DHCP reservation for `192.168.1.3` and active WAN0 TCP/UDP
   forwards for WAN 8088 → PC 8088 and WAN 8443 → PC 8443. The firmware rejects
   forwarding external 80/443 as reserved management ports. Do not forward the
   database or API port 3000.
4. The router WAN address (`100.234.203.11`) differs from the public egress address
   (`223.185.134.11`), which is carrier-grade NAT. Direct Internet reachability is
   therefore not available yet; ask Airtel for a public/static IPv4 (and inbound
   ports) or use an authenticated tunnel. No tunnel is installed or connected here.
5. Set the domain's A record to the current public IPv4, and www CNAME to the
   apex. Dynamic public IPs require a DNS update mechanism. Nameservers and email
   records can stay unchanged. The parked-domain address is not this PC.
6. Copy `deploy/postgres-native.service`, `deploy/panchang.service` and
   `deploy/astrisk-edge.service` into the user's systemd units after checking
   its paths. Enable it only once DNS/networking are ready. Apply the secure-cookie
   API override when switching to HTTPS; do not enable it while using HTTP login.
7. `systemctl --user daemon-reload` and
   `systemctl --user enable --now astrisk-edge.service` start the native proxy.
   Test from a separate internet connection, not just the same LAN.

## Temporary public access through localhost.run

When Airtel CGNAT prevents inbound connections, `deploy/astrisk-tunnel.service`
provides a free encrypted SSH tunnel to the API/frontend on port 3000. It does
not require Docker, a public IP, or Cloudflare. The service prints its generated
HTTPS URL in the user journal:

```bash
systemctl --user enable --now astrisk-tunnel.service
journalctl --user -u astrisk-tunnel.service -f
```

The free no-key URL changes when the tunnel reconnects. Registering an SSH key at
localhost.run can provide a longer-lived free subdomain. This URL is suitable for
testing and demos; keep the application login enabled and do not expose database
or administration ports.

Windows must remain awake with WSL running. A Linux user service cannot boot WSL
by itself. Configure Windows startup and user lingering only with administrator
access; neither is assumed to be enabled. Restart recovery for the native
PostgreSQL instance also needs verification before unattended/public use.

Microsoft WSL networking reference:
https://learn.microsoft.com/windows/wsl/networking
