# Panchang + Geocentric Birth Sky

Go/cgo Panchang engine backed by Swiss Ephemeris, PostgreSQL read-through cache, and a React/Three.js geocentric chart client.

## Run locally

The repository vendors the Swiss Ephemeris C source required by cgo and the current Sun/Moon ephemeris files. Swiss Ephemeris is AGPL; review `engine/swe/LICENSE` before distribution.

Run directly on this machine (no Docker):

```bash
bash scripts/start-native.sh
# App and API: http://localhost:3000
```

The script starts a dedicated PostgreSQL 16 instance using a private Unix socket, applies migrations and seeds, builds the app, and starts the installed `panchang.service` user service. Go serves the compiled frontend and API together. Database files and PostgreSQL logs live in `.runtime/`; application logs are available with `journalctl --user -u panchang.service`. Stop it with `bash scripts/stop-native.sh`; database data is retained.

For frontend development:

```bash
make test
make run
cd web && npm install && npm run dev
```

`DATABASE_URL` is optional. Without it, the API calculates values but does not cache them and `/api/festivals` returns 503.

## Endpoints

- `GET /healthz`
- `GET /api/panchang?date=2026-09-22&lat=12.97&lon=77.59&tz=Asia/Kolkata`
- `GET /api/month?year=2026&month=9&lat=12.97&lon=77.59&tz=Asia/Kolkata`
- `GET /api/chart?date=1996-05-14&time=10:15&lat=12.97&lon=77.59&tz=Asia/Kolkata&ayanamsa=lahiri`
- `GET /api/chart/facts?date=1996-05-14&time=10:15&lat=12.97&lon=77.59&tz=Asia/Kolkata&as_of=2026-09-22T00:00:00Z`
- `GET /api/reading?date=1996-05-14&time=10:15&lat=12.97&lon=77.59&tz=Asia/Kolkata&as_of=2026-09-22T00:00:00Z`
- `GET /api/festivals?year=2026&region=south`

## Validation status

Automated tests include concurrent chart comparisons to the upstream Swiss Ephemeris CLI, a Bengaluru Drik Panchang date check, western-timezone civil dates, and browser checks of chart placement, calendar navigation and WebGL fallback. See [validation details](docs/validation.md) for reference values, tolerances and known limits. Amanta naming is calculated from new moons; full festival materialization, Purnimanta variants and the originally requested multi-date external certification remain unfinished.

The frontend does no astronomy. It renders the single `/api/chart` response and lazy-loads the Three.js dome. The SVG chart remains available when WebGL is unavailable.

## Interpretation layer

`/api/chart/facts` returns deterministic engine output: dignity, combustion, retrograde, Vimshottari Maha–Antara periods, and yoga geometry. `/api/reading` gives a structured interpretation. With no `OPENAI_API_KEY`, it returns a safe deterministic fallback so the route remains usable. Set `OPENAI_API_KEY`, `OPENAI_BASE_URL`, and `OPENAI_MODEL` to enable an OpenAI-compatible hosted model. The post-generation validator rejects a response if its yoga names or count differ from the engine facts.
