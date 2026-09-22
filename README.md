# Antariksha: Panchang, Kundali and a chart-grounded assistant

Free for everyone: no payments, no premium tier, no remedies for sale.

**Features:** Rashi kundali (South Indian, North Indian, circular and 3D views) with divisional charts D2, D3, D7, D9, D10 and D12 · Vimshottari maha, antar and pratyantar dashas · Yogini dasha · Ashtakavarga (Bhinna and Sarva) · yogas, dignity, combustion · a daily "Today" view (transits from Lagna and Moon, tara bala, chandra bala, Sade Sati) · a grounded reading and the **Ask Antariksha** chat, with saved Q&A · **Kundli matching** (Ashtakoota 36 gunas, doshas and cancellations, Mangal dosha for both) · a **muhurta finder** for seven event types · the daily Panchang and Hindu calendar with festivals, named Ekadashis and .ics export · Amanta or Purnimanta months · English or Hindi · worldwide birthplace search (GeoNames) · family profiles · printable report · consent and delete-my-data.


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
- `POST /api/chat` `{birth:{date,time,lat,lon,tz}, question, session_id?}` — chart-grounded answer; the Q&A is stored
- `GET /api/chat/history?session_id=…` — every question and answer in a session
- `DELETE /api/chat/session?session_id=…` — delete a conversation
- `GET /api/chart/varga?…&n=9` — divisional chart (1, 2, 3, 7, 9, 10, 12)
- `POST /api/match` `{boy, girl}` — Ashtakoota Guna Milan
- `GET /api/muhurta?event=marriage&date=…&days=30&lat&lon&tz[&bdate&btime&blat&blon&btz]`, plus `GET /api/muhurta/events`
- `GET /api/today?date&time&lat&lon&tz` — transits, tara bala, chandra bala, Sade Sati, upcoming festivals
- `GET /api/places?q=bangalore` — search about 34,000 cities with timezones
- `GET /api/calendar.ics?year&lat&lon&tz` — the year's festivals as iCalendar

## Validation status

Automated tests include concurrent chart comparisons to the upstream Swiss Ephemeris CLI, a Bengaluru Drik Panchang date check, western-timezone civil dates, and browser checks of chart placement, calendar navigation and WebGL fallback. See [validation details](docs/validation.md) for reference values, tolerances and known limits. Amanta naming is calculated from new moons; full festival materialization, Purnimanta variants and the originally requested multi-date external certification remain unfinished.

The frontend does no astronomy. It renders the single `/api/chart` response and lazy-loads the Three.js dome. The SVG chart remains available when WebGL is unavailable.

## Interpretation layer

`/api/chart/facts` returns deterministic engine output: dignity, combustion, retrograde, Vimshottari Maha–Antara periods, and yoga geometry. `/api/reading` gives a structured interpretation. With no `OPENAI_API_KEY`, it returns a safe deterministic fallback so the route remains usable. Set `OPENAI_API_KEY`, `OPENAI_BASE_URL`, and `OPENAI_MODEL` to enable an OpenAI-compatible hosted model. The post-generation validator rejects a response if its yoga names or count differ from the engine facts.

Readings are grounded in a rights-cleared classical corpus: every detected fact (yoga, graha-in-house/sign, nakshatra, dasha, dignity, lagna) is fetched by exact key from `astro_corpus` (PostgreSQL + pgvector) and returned as `grounding`. It comprises self-authored entries with full engine coverage plus cited public-domain passages from the 1885 Brihat Jataka translation. See [corpus ingestion](docs/corpus.md).
