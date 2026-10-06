# Transits and forecasts

Phases 1-2 are in place: the engine, the book rules, the forecast scoring, the Timing tab and the
planets-now card on Today. Ask Astrisk forecast answers and alerts follow in phases 3-4.

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

## Exactness

- `engine/transit_test.go`: the house table read off 104.4; 2025 sign changes and stations against
  published dates (Saturn into Meena 29 Mar 21:45 IST, Jupiter into Mithuna 14 May 22:36 IST, and so on);
  every ingress has the old sign two minutes before and the new one two minutes after.
- `corpus/maps/brihat_samhita_iyer_1884.json`: 84 clauses (Sun to Saturn, 12 houses), each matched against
  the scan; `corpus/map_exact_test.go` checks each names its own house and planet.

## Wording

Future predictions are windows and themes, never events. The book warns of illness, prison and the
like; plain summaries say what the book warns of and frame it as a time for care.
