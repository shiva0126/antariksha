# Transits and forecasts

All four phases are in place: the engine, the book rules, the forecast scoring, the Timing tab, the
planets-now card on Today, timing answers in Ask Astrisk, the classical dasha results and planet alerts.

## Method

Classical transit results count houses from the **natal Moon** (Brihat Samhita ch. 104, tr. Chidambaram
Iyer 1884); houses from the lagna are shown as well. For each planet the engine reports:

| Field | Source |
|---|---|
| `favourable` | 104.4: benefic houses from the Moon per graha (none for Rahu and Ketu) |
| `active` | 104.49-50: the Sun and Mars act in the first half of a sign, the Moon and Saturn in the second |
| `weakened` | 104.53: good results fail when the planet is debilitated, in an enemy sign or combust |
| `bindus`, `sarva` | the natal ashtakavarga score of the sign being transited (4+ bindus, 28+ sarva: supportive) |
| `aspects` | natal planets occupied or aspected (all aspect the 7th; Mars 4/8, Jupiter 5/9, Saturn 3/10) |
| book view | the exact 104 clause for that planet and house, with a plain summary |

Also: Sade Sati, Kantaka (4th) and Ashtama (8th) Saturn, and Jupiter-Saturn double transit (modern, labelled).
104.46 is the rule for prediction: a transit gives results only according to the running dasha, so
forecasts combine the two (phase 2).

`engine.TransitEvents` lists sign changes, retrograde and direct stations (to the minute) and eclipses;
`GET /api/transits?date=&time=&lat=&lon=&tz=&months=24` returns today's reading and the personal events.
Rahu and Ketu use the true node, as birth charts do; panchangs often publish mean-node dates, a week or
two apart.

## Forecast periods (`reading/forecast.go`)

The coming months are cut at every sign change of Jupiter, Saturn, Rahu and Ketu and every change of
antardasha. In each period, each life area (career = 10th, money = 2nd and 11th, partnership = 7th,
home = 4th, learning = 5th, effort = 3rd, wellbeing = 1st and 6th, change = 8th, travel = 9th and 12th)
is read from:

- the dasha lords that sit in or rule its houses, toned by their natal strength (dignity, a 6th/8th/12th
  placement, combustion);
- Jupiter and Saturn occupying or aspecting its houses, toned by the 104.4 verdict, 104.53 weakening
  and ashtakavarga support, and halved when the dasha does not touch the area (104.46);
- Rahu and Ketu, and double transit, as emphasis without a direction.

An area is shown when at least two real factors touch it (or one plus double transit). Tone follows the
score (≥1 supportive, ≤-1 challenging, otherwise mixed); confidence counts the factors that lean one way
and agree. Every reason carries its source. `reading/forecast_test.go` checks sourcing, tone and that no
text makes a forbidden prediction.

In the app: Kundali → Timing (the `#kundali/dasha` tab, renamed) shows the periods with folded reasons
and the coming sign changes and eclipses; Today shows the planets read from the Moon with the book's view.

## Ask Astrisk

Questions with "when", "this year", "next year", "coming months", "future", "will I", "transit" and the
like are classified as `forecast` (`reading/chat.go`), and only then does the chat handler compute the
next 12 months of periods. A question that names an area (career, marriage, money, studies, health,
travel, home…) gets that area's supportive and patience windows with a reason each, and always: "A chart
cannot say whether or when a particular event will happen…". Health questions point to a doctor; death
questions get the safety reply. General questions get the current period, the next two, and Jupiter's
and Saturn's coming sign changes. The local model never answers timing questions (its chart lines carry
no forecast). `reading/forecast_test.go` covers these, including "Will my business fail next year?".

## Dasha results

Brihat Jataka 8.12-8.18 (tr. Iyer 1885) give a benefic and a malefic result for each planet's dasha, the
Sun to Saturn; which applies depends on the planet's strength (8.5-7, 8.19). They are mapped as `dasha`
passages with a plain summary of both, shown in dasha answers and under each forecast period's main
dasha in Timing. The scan lost the verse numbers after 8.5, so these cite their true verses with
`"in": "8.5"`.

## Planet alerts

Opt-in (Me → Notifications; needs birth date, time and place): a week ahead of each sign change of
Jupiter, Saturn and Rahu/Ketu, each member gets one notification read for their chart, e.g. "On 29 March
Saturn moves into Meena: your 1st house from the Moon, a harder house for it in the classical books".
The alert worker queues them every six hours (`api/transit_alerts.go`) and delivers them like other
notifications, by push and, if the email is verified, email. The subject encodes the event, and
member_notifications is unique per recipient and subject, so each event is sent once. Migration 20 adds
`member_accounts.transit_alerts` and the `transit` notification kind. `api/transit_alerts_test.go` checks
opt-in, once-only queueing, the text, the listing and that turning alerts off hides them.

## Exactness

- `engine/transit_test.go`: the house table read off 104.4; 2025 sign changes and stations against
  published dates (Saturn into Meena 29 Mar 21:45 IST, Jupiter into Mithuna 14 May 22:36 IST, and so on);
  every ingress has the old sign two minutes before and the new one two minutes after.
- `go run ./cmd/chartcheck -ingress 2025`: every sign change of the Sun to Saturn the engine finds in
  two years, checked against NASA JPL Horizons two minutes either side. 2026-10-06: 96 of 96 agree.
- `corpus/maps/brihat_samhita_iyer_1884.json`: 84 clauses (Sun to Saturn, 12 houses), each matched against
  the scan; `corpus/map_exact_test.go` checks each names its own house and planet.

## Wording

Future predictions are windows and themes, never events. The book warns of illness, prison and the
like; plain summaries say what the book warns of and frame it as a time for care.
