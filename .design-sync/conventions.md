## Astrisk conventions

Astrisk is a **dark-only** Vedic astrology app (kundali, panchang, matching, matrimony). Build every
screen on the dark page surface; never place components on white.

- **Brand:** the eight-point `NorthStar` in gold (`--c-accent`). One `Logo` per page, top-left. Display
  type is Marcellus (`--font-display`) for titles; body is DM Sans (`--font-body`).
- **Tokens only:** colours, spacing and radii come from CSS variables (`--c-*`, `--s-*`, `--r-*`,
  `--fs-*`). Do not invent hex colours. Planet colours are for chart marks only, never for UI chrome.
- **Page anatomy:** `PageHeader` (kicker = section, title = subject) → optional `Tabs` with
  `base="<section>"` → content in `Card`s. Do not nest cards. Use `StatTile` for a single key value.
- **Actions:** one `Button variant="primary"` per screen or card; `danger` only for destructive actions,
  confirmed via `useFeedback().confirm`. Labels are sentence-case verbs.
- **Forms:** every control inside a `Field`; two columns via `FormGrid` (collapses on phones).
- **Feedback:** wrap the app in `FeedbackProvider`; toasts for transient results, `Notice` for in-place
  messages, `EmptyState` for empty lists, `Skeleton` while loading. Status `Chip`s always carry words.
- **Layout helpers** (classes in the stylesheet): `ds-stack`, `ds-row`, `ds-grid-auto`, `ds-muted`,
  `ds-small`, `dialog-actions`.
- Everything in the product is free: never design pricing, paywalls or upgrade prompts.
