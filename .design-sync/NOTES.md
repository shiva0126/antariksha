# design-sync notes (Astrisk)

- The design system lives in the app at `web/src/ds` (no separate package). `buildCmd` builds it as a
  library into `web/dist-ds` (vite.ds.config.ts + tsconfig.ds.json); the vite plugin writes
  `dist-ds/package.json` (name `@astrisk/ds`) so the converter treats dist-ds as the package root.
  Run the converter with `--entry ./web/dist-ds/index.js --node-modules web/node_modules`.
- `srcDir` and `docsDir` are relative to `web/dist-ds` (hence `../src/ds`, `../src/ds/docs`).
  Per-component docs with `category` frontmatter live in `web/src/ds/docs/<Name>.md`.
- Fonts (DM Sans, Marcellus, Noto Serif Devanagari) come from Google Fonts: the library CSS starts with
  the same `@import` as `web/index.html`, and `runtimeFontPrefixes` suppresses FONT_MISSING.
- Astrisk is dark-only, but the card harness paints `body` white. Every preview wraps its stories in a
  local `Night` surface (var(--c-bg)). Dialog is the exception: it renders in the top layer.
- `ds.css` starts with a zero-specificity `:where(body)` base so designs using only the DS get the dark
  page and brand type; the app's own `styles.css` still wins.
- Playwright: the repo pins 1.63 (chromium_headless_shell-1243) but the machine cache holds build 1234;
  `~/.cache/ms-playwright/chromium_headless_shell-1243` is a symlink to 1234. Chromium also needs
  `LD_LIBRARY_PATH=<scratch>/chromelibs/root/usr/lib/x86_64-linux-gnu` on this WSL box.

## Known render warns
- none outstanding (GRID_OVERFLOW resolved with `cardMode: column` overrides).

## Re-sync risks
- `useFeedback` and `personName` are exported but not components; they ship in the bundle only.
- The Google Fonts URL is duplicated in `web/index.html` and `web/vite.ds.config.ts`; keep them in step.
- Docs in `web/src/ds/docs` are hand-written; update them when a component's props change.
