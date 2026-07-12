# Theme Contrast & Heading Hierarchy Fix

## Goal
Every theme in `web/src/themes.css` satisfies one rule everywhere: by default, all text contrasts with its background. Concretely:
- Heading colors (h1-h6) form a real visual hierarchy: h1 is the most distinct from normal body text, and each level down (h2...h6) steps closer to body-text color, with h6 the closest — but never fully equal to it. Today several themes violate this (h1 is literally identical to `--text` in github-light, github-dark, kanagawa, and solarized-light; h5/h6 collapse onto `--text-muted`).
- All content rendered inside `.prose` (paragraphs, bold text, list items/bullets/counters, blockquotes and their border, table cells and borders, captions, `<hr>`, inline code, code-block wrapper) is legible against `--bg` in every theme — not just headings and links. Today only h1-h6 and links are wired to theme tokens; everything else inherits `@tailwindcss/typography`'s hardcoded light-mode default palette via its own `--tw-prose-*` CSS variables, which is why dark themes currently render normal paragraph text as dark-on-dark.

## Failure Criteria
- Any heading level in any of the 6 themes (github-light, github-dark, kanagawa, gruvbox, solarized-light, one-dark) is color-identical to `--text` or to an adjacent heading level (e.g. h1 == h2, h3 == h4).
- Any heading level fails WCAG AA contrast against `--bg` (< 3:1, "large text" threshold — headings are always large/bold under the Typography plugin's default scale).
- Body text (`--tw-prose-body`) or any other `.prose` content var fails WCAG AA against `--bg` (< 4.5:1 for normal-size text) in any theme.
- Any `.prose` sub-element (bold text, blockquote text/border, table borders, list bullets/counters, captions, `<hr>`, inline code) is left unstyled by the theme system and falls through to the Typography plugin's raw default color in a way that's visibly wrong against that theme's `--bg` (e.g. still dark-gray on a dark background).
- h6 (or h5) becomes exactly equal to `--text`, eliminating the "still a heading" visual signal entirely.
- Fixing one theme's palette breaks another theme's already-correct values (e.g. hand-tuning kanagawa regressing gruvbox, which was already correct).

## Ambiguous Zone
- Exact hex values for each theme's h1-h6 ramp are a hand-tuned, per-theme judgment call (not derived from a formula) — as long as the WCAG bar and hierarchy ordering hold, the specific color choice is a matter of taste, not something to re-litigate against a mechanical rule.
- gruvbox and one-dark already read as correct (h1 already distinct from body text, already passes the informal hierarchy check) — they still get re-checked against the formal WCAG bar and hierarchy rule as part of "hand-tune all 6," but are expected to need little or no change.
- App chrome (header bar, sidebar tree, tab strip, Settings page — `--text`/`--text-muted`/`--accent` against `--bg`/`--bg-subtle`) is out of scope for this pass; it was spot-checked during grilling and already has good contrast in all 6 themes. If a specific chrome contrast bug surfaces later, it's a separate fix.
- Heading font-size is explicitly not touched — the Typography plugin's default type scale (h1 largest → h6 near body-size) already satisfies "header is bigger," confirmed during grilling. This fix is color-only.

## Direction
Hand-tune all 6 themes' `--h1` through `--h6` values in `web/src/themes.css` against a written hierarchy rule: h1 has the strongest color differentiation from `--text`, each subsequent level steps closer to `--text`, and h6 stays slightly but always visibly distinguishable from plain body text (never color-identical). Every heading level in every theme must independently clear WCAG AA for large text (≥ 3:1 against `--bg`).

Separately, wire the full set of Tailwind Typography plugin CSS variables to theme tokens so nothing inside `.prose` is left on the plugin's hardcoded defaults: `--tw-prose-body`, `--tw-prose-bold`, `--tw-prose-quotes`, `--tw-prose-quote-borders`, `--tw-prose-captions`, `--tw-prose-hr`, `--tw-prose-bullets`, `--tw-prose-counters`, `--tw-prose-th-borders`, `--tw-prose-td-borders`, `--tw-prose-code`, `--tw-prose-pre-code`, `--tw-prose-pre-bg`, and any others the plugin defines that render visible content. Map these to the existing `--text`/`--text-muted`/`--border`/`--code-*` tokens per theme (no new token family needed — the tokens already exist and already have good contrast against `--bg`, they're just unused by `.prose` sub-elements today). Body text (`--tw-prose-body`) must clear WCAG AA for normal text (≥ 4.5:1 against `--bg`) in all 6 themes.

Both fixes land together in `web/src/themes.css` since they're the same root cause category (theme tokens not fully wired to `.prose` output) and the same acceptance bar (WCAG AA, checked per theme, in-browser or via a contrast calculator).

## Constraints
- Fix lives entirely in `web/src/themes.css` (token values + new `.prose`-scoped CSS rules for the Typography plugin vars). No changes to `ContentPane.tsx`, `theme.ts`, `ThemePicker.tsx`, or the Go markdown pipeline.
- Must not regress the existing Chroma code-block token styling (`.chroma`, `[class^='k']` etc. — already correctly wired, untouched by this fix) or the h1-h6/link rules' existing selector structure — only the token *values* and the *addition* of new `.prose`-scoped rules for previously-unwired Typography vars.
- Must not regress the nested `data-theme` hover-preview mechanism ([[20260712-221510-nested-data-theme-scopes-cascade-correctly]]) — new rules must be theme-token-driven (`var(--...)`), not hardcoded per-theme selectors outside the `[data-theme='...']` blocks.
- Acceptance bar is WCAG AA (4.5:1 body text, 3:1 headings/large text) against that theme's `--bg`, checked per theme, all 6 themes.

## Out of Scope
- App chrome contrast audit (header bar, sidebar, tabs, Settings) — already spot-checked as fine; not part of this fix.
- Heading font-size changes — existing Typography plugin scale already satisfies "bigger" requirement.
- A mechanical/derived color formula (e.g. deriving h1-h6 from `--text` via fixed lightness/opacity steps) — rejected in favor of hand-tuned per-theme values, since two themes (gruvbox, one-dark) already show hand-tuning produces good results, and a shared reference palette (GitHub, Kanagawa, etc.) argues for eyeballed fidelity over a generated ramp.
- Adding new themes or changing the theme list/count.
- Reopening `.context/direction/20260712-215915-md-rendering-theme.md` — that direction described the now-shipped theme-system build; this is a follow-up fix, tracked separately.
