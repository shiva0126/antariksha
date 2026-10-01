# Astrisk — astrisk.space

The product is now Astrisk. The existing GitHub repository remains
`shiva0126/antariksha`; service names, database identifiers, session cookies and
browser storage keys intentionally remain unchanged so rebranding does not lose
accounts or saved charts. Moving to a new origin still requires sign-in, and old
origin-local browser charts do not move automatically.

## Current deployment (2026-10-01)

`https://astrisk.space` and `https://www.astrisk.space` serve Astrisk from this
Windows/WSL PC through a named Cloudflare Tunnel on the Free plan. HTTP redirects
to HTTPS at Cloudflare. The registrar remains Hostinger; nameservers are now
`jermaine.ns.cloudflare.com` and `ligia.ns.cloudflare.com`. The two parked web
records were replaced with proxied CNAME routes to tunnel `astrisk-home`.

The tunnel connects outbound to Cloudflare and forwards locally to
`127.0.0.1:3000`. No public/static IP or inbound router forwarding is required.
The API production override binds only to loopback and sets `COOKIE_SECURE=true`.
Cloudflare Universal SSL is active. Keep the tunnel credentials and account
certificate outside Git; they live under `~/.cloudflared/`. The runtime tunnel
configuration and binary live in ignored `.runtime/cloudflared/`.

Enabled user services are `postgres-native.service`, `panchang.service`, and
`astrisk-cloudflare.service`. User lingering is enabled. The Windows scheduled
task `Astrisk WSL Server` starts Ubuntu at Windows sign-in and keeps WSL running.
The task is installed with `scripts/install-windows-startup.ps1`. Windows still
needs to stay awake, powered on, and connected to the Internet. Startup after a
full Windows reboot requires Windows sign-in; an unattended reboot was not tested.

The temporary localhost.run tunnel and direct Caddy edge are stopped and disabled.
Their templates remain available as alternatives. Existing router forwards are
not used by Cloudflare Tunnel.

```bash
systemctl --user status panchang.service astrisk-cloudflare.service
journalctl --user -u astrisk-cloudflare.service -n 50
curl https://astrisk.space/healthz
```

To recreate the connector, install the official `cloudflared` binary into
`.runtime/cloudflared/cloudflared`, authenticate with `cloudflared tunnel login`,
and use the existing tunnel's protected credential file (or deliberately create
a replacement tunnel and update both DNS routes). Copy
`deploy/cloudflared-config.example.yml` to `.runtime/cloudflared/config.yml`,
replace the tunnel ID and credential path, and validate it with
`cloudflared tunnel --config .runtime/cloudflared/config.yml ingress validate`.
Install `deploy/astrisk-cloudflare.service` into `~/.config/systemd/user/`, run
`systemctl --user daemon-reload`, then enable/start the service.

The installed cloudflared version is `2026.9.3`; the downloaded Linux amd64 Debian
package SHA-256 was checked against its official GitHub release digest:
`bc073ef293d504cf5ac533bd0aa1c824ef6b4f358765ccaa6628a8a95cacb4b7`.
Automatic binary updates are disabled; review and install updates explicitly.

## Alternative: direct DNS cutover to a public hosting target

This alternative is not used by the current Cloudflare deployment. Do not replace
the working tunnel CNAME records unless intentionally migrating hosting.

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
   therefore not available directly. The active Cloudflare Tunnel described above
   bypasses this limitation using outbound connections.
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

Windows must remain awake with WSL running. The current deployment uses the
Windows sign-in task and user lingering described above. A Linux user service
alone cannot boot WSL. These localhost.run commands are a manual fallback and
should not be enabled alongside the primary tunnel unless needed for diagnosis.

Microsoft WSL networking reference:
https://learn.microsoft.com/windows/wsl/networking
