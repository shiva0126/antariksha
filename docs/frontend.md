# Frontend structure and design system

The web app (`web/`) follows the redesign plan from the October 2026 frontend
audit. Design reference: the Claude Design canvas "Astrisk brand, design system
and key screens" (north star logo, colour, type, components, desktop and phone
screens).

## Layout

| Path | What lives there |
| --- | --- |
| `src/ds/` | The design system: `tokens.css` (every colour, size, space, radius), `ds.css` (component styles and the baseline for native controls) and the React components (`Button`, `Field`, `Input`, `Select`, `Textarea`, `Checkbox`, `Tabs`, `Segmented`, `PageHeader`, `Card`, `StatTile`, `Chip`, `EmptyState`, `Notice`, `Skeleton`, `Avatar`, `Dialog`, `FeedbackProvider`/`useFeedback`, `NorthStar`, `Logo`). Import only from `src/ds`. |
| `src/lib/` | `api.ts` (the one HTTP client) and `router.ts` (routes). |
| `src/features/` | New sections: `shell` (header, bell, Me menu, phone tab bar), `me`, `admin`, `panchang`. |
| `src/ui/` | Pages not yet moved to `features/` (Kundali, Matching, Matrimony, Community); they already use the design system. |

## Navigation

Five sections — Kundali, Panchang, Matching, Matrimony, Community — plus the
notifications bell and the **Me** menu (Profile, Saved charts, Notifications,
Security, Privacy and data, Settings, Admin, Sign out). Every page and tab has
an address `#section/sub`, for example `#kundali/dasha`, `#panchang/muhurta`,
`#me/settings`. Old addresses (`#day`, `#month`, `#muhurta`, `#match`,
`#readings`, `#account`, `#privacy`) redirect.

Kundali tabs: Overview · Chart · Planets (positions, Shadbala, Ashtakavarga) ·
Dasha · Reading (with the printable report) · Ask Astrisk · Other systems
(numerology, Western astrology, tarot).

## Rules (enforced by `src/ui-rules.test.ts`)

- No browser `prompt`/`confirm`/`alert`: use `useFeedback().confirm`, `.prompt`, `.toast`.
- Code in `src/features/` uses design-system controls: no raw `<button>`, no inline `style={{…}}`.
- Colours in `src/ds` and `src/features` CSS come from `tokens.css`.
- Every visible label goes through `t()`; add new keys to `src/i18n/strings.ts` with all eight translations.

## Charts

Graha chart marks (dasha bar, dots) use `grahaMark` in `src/astro/rashi.ts`, a
palette validated with the dataviz checker on the dark surface (lightness band,
chroma floor, colour-blind and normal-vision separation in dasha order, 3:1
contrast). Text stays in text colours; a coloured mark sits beside the name.
`grahaColor` is only for planet glyphs.

## Testing

`npm run test` (unit and UI rules) and `scripts/test-browser.sh` (Playwright on
a disposable database). Screens are reviewed at 1366 px and 390 px wide.
