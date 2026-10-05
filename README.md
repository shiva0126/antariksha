# Astrisk: Panchang, Kundali and a chart-grounded assistant

Domain: **astrisk.space**. Branding is updated; public DNS/HTTPS cutover is tracked
in [the domain deployment guide](docs/astrisk-domain.md). The GitHub repository and
internal storage identifiers retain their original names for compatibility.

Free for everyone: no payments, no premium tier, no remedies for sale.

**Features:** Rashi kundali (South Indian, North Indian, circular and 3D views) with divisional charts D2, D3, D7, D9, D10 and D12 · Vimshottari maha, antar and pratyantar dashas · Yogini dasha · Ashtakavarga (Bhinna and Sarva) · Shadbala (six-fold strength with Ishta/Kashta phala) · yogas, dignity, combustion · a daily "Today" view (transits from Lagna and Moon, tara bala, chandra bala, Sade Sati) · a grounded reading and the **Ask Astrisk** chat, with saved Q&A · **Kundli matching** (Ashtakoota 36 gunas, doshas and cancellations, Mangal dosha for both) · a **muhurta finder** for seven event types · the daily Panchang and Hindu calendar with festivals, named Ekadashis and .ics export · Amanta or Purnimanta months · English, Hindi, Marathi, Kannada, Tamil, Telugu, Malayalam, Gujarati and Bengali · worldwide birthplace search (GeoNames) · family profiles · printable report · consent and delete-my-data.


Go/cgo Panchang engine backed by Swiss Ephemeris, PostgreSQL read-through cache, and a React/Three.js geocentric chart client.

## Run locally

The app opens with email/password login. Registration collects DOB and birth time;
hobbies and other details stay in the separate profile page. All application APIs
require authentication. `/#community` provides photo/text posts, audience controls,
followers, private family groups/trees, optional matrimony, mutual-match messages,
an in-app notification inbox and explicit legacy device-chart import,
revocable family-assisted matrimony shortlists,
blocking, recovery keys and export. Phone linking requires Twilio configuration;
private chart Save/Load and email recovery are implemented. SMTP delivery requires
provider configuration; external social OAuth connections remain pending.
See the [implementation status and remaining roadmap](docs/community-matrimony-plan.md)
and [native operations guide](docs/community-operations.md). Legacy browser charts
and unowned conversations are not automatically assigned to accounts.

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
- `GET /api/chart/shadbala?date&time&lat&lon&tz` — six-fold planetary strength
- `POST /api/match` `{boy, girl}` — Ashtakoota Guna Milan
- `GET /api/muhurta?event=marriage&date=…&days=30&lat&lon&tz[&bdate&btime&blat&blon&btz]`, plus `GET /api/muhurta/events`
- `GET /api/today?date&time&lat&lon&tz` — transits, tara bala, chandra bala, Sade Sati, upcoming festivals
- `GET /api/places?q=bangalore` — search about 34,000 cities with timezones
- `GET /api/calendar.ics?year&lat&lon&tz` — the year's festivals as iCalendar

## Validation status

Automated tests include concurrent chart comparisons to the upstream Swiss Ephemeris CLI, a Bengaluru Drik Panchang date check, western-timezone civil dates, and browser checks of chart placement, calendar navigation and WebGL fallback. See [validation details](docs/validation.md) for reference values, tolerances and known limits. Amanta naming is calculated from new moons; full festival materialization, Purnimanta variants and the originally requested multi-date external certification remain unfinished.

The frontend does no astronomy. It renders the single `/api/chart` response and lazy-loads the Three.js dome. The SVG chart remains available when WebGL is unavailable.

The signed-in navigation also has **More readings**. Numerology is a documented Pythagorean letter/date method; Western astrology uses Swiss Ephemeris tropical positions and offers Placidus or whole-sign houses; Tarot draws without replacement from a 78-card deck with original reflection prompts. Each explains its convention and presents symbolic prompts in everyday language. These routes are private, require a signed-in account and do not save the submitted name or reading. The numerology endpoint uses the date you submit rather than silently assuming your account birth date.

## Interpretation layer

`/api/chart/facts` returns deterministic engine output: dignity, combustion, retrograde, Vimshottari Maha–Antara periods, and yoga geometry. `/api/reading` gives a structured interpretation. With no `OPENAI_API_KEY`, it returns a safe deterministic fallback so the route remains usable. Set `OPENAI_API_KEY`, `OPENAI_BASE_URL`, and `OPENAI_MODEL` to enable an OpenAI-compatible hosted model. The post-generation validator rejects a response if its yoga names or count differ from the engine facts.

Readings use exact-key retrieval from `astro_corpus`; the engine, not the LLM, supplies placements, periods and detected yogas. The user-facing fallback explains key ideas before traditional terms and keeps technical placement details expandable. The LLM prompt explicitly asks for plain English, and the response yoga validator remains active. The corpus comprises project-authored interpretations plus 108 cited passages from one verified 1885 Brihat Jataka translation. This is a small RAG library, not model training or a complete set of classical books. **Semantic retrieval runs on a free local model:** `astrisk-embedding.service` serves bge-small-en-v1.5 (MIT licence, ONNX, CPU) on loopback port 18091, and every corpus row is embedded with it. Ask Astrisk adds the library passages closest to a question, chosen only from placements in the asker's chart, even with no LLM configured. See [corpus ingestion](docs/corpus.md#local-embeddings).

**Matrimony** is a top-level section with horoscope-aware discovery: guna score, Mangal dosha and nakshatra on every card between members who both opt in, structured biodata with filters and reasons, interests with notes, live chat, contact sharing, selfie verification, printable biodata with a kundali page, email and push alerts. See [docs/matrimony.md](docs/matrimony.md).

**Ask Astrisk** answers questions about a chart (Kundali → Ask Astrisk) and about a compared pair (Matching, below the result). Both correlate numerology with astrology: the birth-date root number (mulank) and destiny number (bhagyank) are mapped to their ruling grahas in Indian numerology, and those grahas are looked up in the computed chart (placement, Shadbala, lagna/Moon/nakshatra lordship, current dasha); for a pair, the number grahas' friendship is compared with Graha Maitri. Matching conversations are not stored.

Family-prepared matrimony biodata, consent-based photo publishing, private character profiles and moderation are described in [the community plan](docs/community-matrimony-plan.md).
