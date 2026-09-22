# Kundali and Panchang verification — 22 September 2026

## Fixes

- North Indian planets are grouped into fixed houses based on the Lagna sign.
- South Indian planets are grouped by their fixed zodiac sign.
- Whole-sign house boundaries use the beginning of the Lagna sign. The precise ascendant remains a separate marker.
- All Swiss Ephemeris calls in one calculation stay on one OS thread, with Lahiri and the ephemeris path configured on that thread. The bundled C library uses thread-local storage.
- A missing ephemeris file raises an error instead of silently switching astronomy engines.
- Calendar dates retain their local civil date west of UTC. Cache hits are rejected when the input timezone or coordinates differ; schema version invalidates earlier calculations.
- Night Choghadiya uses the following actual sunrise.
- Amanta month uses the Sun's sign at the ending new moon; consecutive new moons in the same solar sign mark an Adhika month. Rare lost-month naming and Purnimanta variants are not covered.

## Chart reference

Independently compile the upstream Swiss Ephemeris C command-line tool and run:

```sh
swetest -edir./ephe -b14.5.1996 -ut4:45 -p0123456t -sid1 -fPlbs -g, -head -house77.59,12.97,W
```

Birth instant: 14 May 1996, 10:15 Asia/Kolkata, latitude 12.97, longitude 77.59.

| Body | Sidereal longitude |
| --- | ---: |
| Sun | 29.8635591 |
| Moon | 350.5216950 |
| Mercury | 31.1869963 |
| Venus | 63.7989341 |
| Mars | 14.7410520 |
| Jupiter | 263.7047476 |
| Saturn | 340.2224565 |
| True Rahu | 172.9698223 |
| Ketu | 352.9698223 |
| Ascendant | 90.4806822 |

The Go test compares to these values within 0.000001 degrees in 32 concurrent calls. This checks the Go integration against the same astronomical library, not against an independent ephemeris. House 1 begins at 90°, in Karka. The birth form's more precise Bengaluru coordinates produce a slightly different ascendant.

Reference documentation: https://www.astro.com/swisseph/swephprg.htm

## External Panchang spot check

Drik Panchang's Bengaluru page was inspected on 22 September 2026:
https://www.drikpanchang.com/panchang/day-panchang.html?geoname-id=1277333

For that date: sunrise 06:09; sunset 18:16; Shukla Ekadashi until 21:43; Uttara Ashadha until 07:06; Atiganda until 16:29; Vanija until 08:55; Bhadrapada month. The regression test checks the names and month exactly and the times within one minute. This is one external date check, not the originally requested five-date/ten-date certification.

Moon events are reported within sunrise-to-sunrise, with a date on events after midnight, using the geometric lunar disc centre without refraction. Sunrise/sunset use the apparent upper limb. Rahu/Ketu use true nodes; mean-node charts can differ.

## UI checks

Browser tests cover correct sign/house grouping, all three chart layouts, top navigation, month selection, date selection, mobile horizontal overflow, and operation without WebGL. The dome and wheel must reuse one chart response.

The Daily Panchang and Hindu Calendar pages display graha transits at a separately selected local time, clearly separated from sunrise-based Panchang limbs.

Festival observance rules and regional calendars still need the dedicated festival milestone; this release does not claim a full validated festival calendar.

## Interpretation and festival fixes (22 September 2026)

- **Vimshottari:** the sequence skipped the second mahadasha (for example Ketu after a Mercury birth dasha), shifting every later period. Fixed, and the birth mahadasha's antardashas now run from its theoretical start before birth. Regression test: 14 May 1996 → Mercury to 2008-06, Ketu to 2015-06, Venus to 2035-06, and Venus–Jupiter on 22 Sep 2026.
- **Dignity:** enemy signs are now classified from the full naisargika friend/enemy table; previously most enemy placements came out neutral.
- **Yogas:** Raja Yoga checks every kendra-lord × trikona-lord pair (conjunction, mutual 7th aspect or exchange) and single-planet yogakarakas. Sunapha and Anapha no longer coexist with Durudhara. Neecha bhanga uses the lords of both the debilitation and exaltation signs, from the lagna or the Moon. Kala Sarpa counts both hemispheres.
- **Festivals** are evaluated at each observance's kaal (sunrise, madhyahna, aparahna, pradosh, nishita). Kshaya tithis are assigned to the Panchang day holding the tithi's middle, and named festivals are skipped in Adhika months. Tests pin these 2026 dates to Drik Panchang: Makar Sankranti 14 Jan, Maha Shivaratri 15 Feb, Holi 4 Mar, Ugadi 19 Mar (kshaya Pratipada), Raksha Bandhan 28 Aug, Janmashtami 4 Sep, Ganesh Chaturthi 14 Sep, Navaratri 11 Oct, Dussehra 20 Oct, Dhanteras 6 Nov, Diwali 8 Nov.
- **Known limits:** Bhadra avoidance (Holika Dahan gives 2 Mar where Drik may give 3 Mar), Bhai Dooj's aparahna rule, named Ekadashis and regional or Purnimanta variants are not implemented.
