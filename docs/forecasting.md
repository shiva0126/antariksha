# Transits and forecasts

Phase 1 (engine, book rules, API) is in place; the Timing tab, scoring, Ask Astrisk answers and alerts
follow in phases 2-4.

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

## Exactness

- `engine/transit_test.go`: the house table read off 104.4; 2025 sign changes and stations against
  published dates (Saturn into Meena 29 Mar 21:45 IST, Jupiter into Mithuna 14 May 22:36 IST, and so on);
  every ingress has the old sign two minutes before and the new one two minutes after.
- `corpus/maps/brihat_samhita_iyer_1884.json`: 84 clauses (Sun to Saturn, 12 houses), each matched against
  the scan; `corpus/map_exact_test.go` checks each names its own house and planet.

## Wording

Future predictions are windows and themes, never events. The book warns of illness, prison and the
like; plain summaries say what the book warns of and frame it as a time for care.
