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

## Matching, muhurta and additional calculations

- **Ashtakoota:** Varna, Vashya (Dhanu and Makara split at 15°), Tara (both directions), Yoni (the standard 14×14 table, tested symmetric), Graha Maitri (naisargika), Gana, Bhakoot (2/12, 5/9 and 6/8 dosha) and Nadi. Regression check: identical Moons score the classical 28, with Nadi dosha. Vashya and Gana scoring differ between schools; the commonly published tables are used, and the UI says so. Cancellations shown: Bhakoot (same or mutually friendly lords); Nadi (same nakshatra and different rashi, or the reverse, or different padas). Mangal dosha is lagna-based for both charts, and mutual cancellation is noted.
- **Divisional charts** follow the BPHS rules for D2 (Parashara hora), D3, D7, D9, D10 and D12, with tests at the division boundaries.
- **Ashtakavarga** uses the BPHS bindu tables; per-graha totals of 48, 49, 39, 54, 56, 52 and 39 (SAV 337) are asserted.
- **Yogini dasha** starts from (nakshatra number + 3) mod 8 and applies the birth balance. **Pratyantardashas** are proportional sub-periods of the running antardasha.
- **Named Ekadashis** come from the Amanta month and paksha, with Padmini and Parama in Adhika months (all present in 2026).
- **Muhurta** is evaluated at day level from the sunrise nakshatra, tithi and weekday, excluding Rikta tithis, Amavasya, Bhadra at sunrise and inauspicious nitya yogas, plus personal tara bala and chandra bala. Suggested windows avoid Rahu kaal, Yamaganda and Gulika, and Abhijit is not offered on Wednesdays. Not evaluated: month-level rules (Kharmas, Chaturmas, combust Jupiter or Venus), lagna shuddhi, disha shula.
- **Holika Dahan:** the engine gives 2 Mar 2026. Drik Panchang reports 2 Mar in some western states and 3 Mar in most of India (Bhadra and the lunar eclipse).

## Shadbala

All six balas follow BPHS ch. 27, in virupas, compared against the classical minimums (Sun 5, Moon 6, Mars 5, Mercury 7, Jupiter 6.5, Venus 5.5, Saturn 5 rupas):
- **Sthana:** uchcha, saptavargaja (D1, D2, D3, D7, D9, D12, D30, with compound friendship), ojha-yugma, kendradi, drekkana.
- **Dig bala.**
- **Kala:** nathonnata, paksha (doubled for the Moon), tribhaga, year/month/weekday/hora lords, ayana (from tropical declination; doubled for the Sun), yuddha.
- **Chesta** (the Sun's from ayana bala, the Moon's from paksha bala).
- **Naisargika.**
- **Drik:** sphuta drishti with the special aspects of Mars, Jupiter and Saturn.

Ishta and Kashta phala come from uchcha and chesta.

Tests pin, for 14 May 1996 10:15 Bengaluru: the Sun's uchcha bala (160.14°/3), weekday lord Mars (a Tuesday), hora lord Moon, near-maximal dig bala for Mars in the 10th, retrograde chesta for Mercury and Jupiter, the compound-relationship table, and the special aspects.

Approximations, stated in the UI notes:
- Dig bala uses equal-house cusps from the ascendant.
- Nathonnata bala uses local mean time.
- Chesta bala for Mars–Saturn uses the eight-motion (ashta chesta) classification of apparent speed rather than the full chesta-kendra computation.
- Year and month lords come from the Kali ahargana.

Results can therefore differ by a few virupas from software that computes the full chesta kendra.

## Languages

The UI is available in English, Hindi, Marathi, Kannada, Tamil, Telugu, Malayalam, Gujarati and Bengali.
- **UI strings** are hand-translated (59 per language); native-speaker review is welcome.
- **Sanskrit terms** come from one canonical Devanagari table, transliterated by code point into Kannada, Telugu, Malayalam, Gujarati and Bengali (Bengali writes va as ba). Regional names are used for weekdays and Mars where they differ.
- **Tamil** has its own table of almanac forms (for example Thiruvathirai for Ardra, Pournami for Purnima).
- **Fonts:** a Noto font for the selected script is loaded on demand.
- **Not translated:** long explanatory sentences, readings and chat answers remain in English.
