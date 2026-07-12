# ADR: Theme Contrast & Heading Hierarchy Fix

**Date:** 2026-07-12

## Context

Direction `.context/direction/20260712-232316-theme-contrast-hierarchy-fix.md` is a follow-up
to the shipped theme system (`.context/adr/20260712-221004-md-rendering-theme.md`). Two defects
were found by inspecting `web/src/themes.css` directly:

1. **Heading hierarchy is broken in 4 of 6 themes.** `h1` is color-identical to `--text` in
   github-light, github-dark, kanagawa, and solarized-light. Adjacent levels collapse onto each
   other (`h1==h2` in github-light/github-dark/solarized-light/one-dark; `h3==h4` in
   github-light/github-dark/solarized-light; `h5==h6` in github-light/github-dark/kanagawa/
   solarized-light/gruvbox). gruvbox and one-dark are already close to correct.
2. **Only h1-h6 and links are wired to theme tokens inside `.prose`.** Per
   [[20260712-232400-tw-prose-vars-unwired-to-theme-tokens]], everything else
   `@tailwindcss/typography` renders — body text, bold, blockquotes, list markers, tables,
   captions, `<hr>`, inline code — is styled via the plugin's own `--tw-prose-*` CSS custom
   properties, which default to a hardcoded light-mode palette the theme system never touches.
   Dark themes render this content dark-on-dark.

Both are the same root cause (theme tokens not fully wired to `.prose` output) and land in the
same file, per the direction's Constraints.

**Contrast findings from grilling** (computed via WCAG relative-luminance formula against each
theme's actual `--text`/`--text-muted`/`--border` hex values):

| theme | `--text` vs `--bg` | `--text-muted` vs `--bg` | `--border` vs `--bg` |
|---|---|---|---|
| github-light | 15.80 | 6.11 | 1.45 |
| github-dark | 16.02 | 6.31 | 1.55 |
| kanagawa | 11.26 | 5.91 | 2.23 |
| gruvbox | 10.75 | 5.30 | 1.67 |
| solarized-light | 12.05 | 4.13 | 1.50 |
| one-dark | 6.57 | 2.32 | 1.43 |

Two things this reveals that change the naive "reuse `--text-muted`/`--border` everywhere"
approach:
- `--text-muted` fails the 4.5:1 text bar in solarized-light (4.13) and fails even the lighter
  3:1 bar in one-dark (2.32) — it cannot safely back any `.prose` var that WCAG-gates.
- `--border` fails 3:1 in every theme (1.43-2.23) — but the direction's "existing tokens
  already have good contrast" claim (Out of Scope) only ever covered
  `--text`/`--text-muted`/`--accent`, never `--border`. Hairline dividers are conventionally
  low-contrast by design; forcing 3:1 here would make every divider read as a heavy rule
  instead of a subtle line.

## Decision

**Heading hierarchy**: hand-tune `--h1`...`--h6` per theme so no level is ever color-identical
to `--text` or to an adjacent level, and every level independently clears 3:1 against `--bg`.
Verified by exact-equality checks + a per-level contrast-ratio check — not a monotonic
color-distance formula, since the direction's Ambiguous/Out-of-Scope sections explicitly reject
a generated/mechanical ramp in favor of hand-tuned, eyeballed values. The mechanical gate only
enforces the two literal Failure Criteria bullets (no identical levels, no sub-3:1 level), never
second-guesses which specific hex a level uses.

**`.prose` sub-element token mapping** — one theme-agnostic CSS rule block (`.prose { --tw-prose-*: var(--token); }`), resolved per-theme automatically via `var()`, added once, not per theme:

| `--tw-prose-*` var | Maps to | WCAG bar | Rationale |
|---|---|---|---|
| `body`, `bold`, `quotes`, `captions`, `code`, `pre-code` | `--text` | 4.5:1 (checked) | Actual rendered text. `--text-muted` was rejected — it fails 4.5:1 in solarized-light and 3:1 in one-dark, so it cannot be trusted to satisfy AA in every theme. |
| `bullets`, `counters` | `--text` | 3:1 (checked) | List-marker glyphs are visible content, not hairlines. Same `--text-muted` rejection applies (one-dark 2.32:1 fails even 3:1). |
| `hr`, `quote-borders`, `th-borders`, `td-borders` | `--border` | none (exempt) | Hairline dividers — never claimed "good contrast" by the direction (only `--text`/`--text-muted`/`--accent` were), and subtlety is the intended look. |
| `pre-bg` | `--code-bg` | none (defensive) | Only reachable via the rare goldmark tokenizer-fallback path (`.chroma` already wins for all normal fenced code — see codeblock.go). One-line insurance. |
| `kbd`, `kbd-shadows`, `lead` | *(left unwired)* | — | Not in the direction's enumerated list; GFM markdown can't produce `<kbd>` or `class="lead"` without an author writing raw HTML — an edge case outside this fix's Failure Criteria. |
| `headings`, `links` | *(left unwired)* | — | Already fully overridden by the existing higher-specificity `.prose h1`..`.prose a` selectors from the prior ADR; the plugin's own vars for these are moot. |

## Design

### `web/src/themes.css` changes

1. **Per-theme blocks**: hand-tune `--h1`...`--h6` values in each of the 6 `[data-theme='...']`
   blocks. No new tokens, no new vars — only value changes to the existing 6 heading vars per
   theme.
2. **New theme-agnostic rule**, appended near the existing `.prose h1`...`.prose a` rules:
   ```css
   .prose {
     --tw-prose-body: var(--text);
     --tw-prose-bold: var(--text);
     --tw-prose-quotes: var(--text);
     --tw-prose-captions: var(--text);
     --tw-prose-code: var(--text);
     --tw-prose-pre-code: var(--text);
     --tw-prose-bullets: var(--text);
     --tw-prose-counters: var(--text);
     --tw-prose-hr: var(--border);
     --tw-prose-quote-borders: var(--border);
     --tw-prose-th-borders: var(--border);
     --tw-prose-td-borders: var(--border);
     --tw-prose-pre-bg: var(--code-bg);
   }
   ```
   One block, not six — exactly matching how the existing `.prose h1 { color: var(--h1); }`
   rules already work: the values differ per theme because `--text`/`--border`/`--code-bg`
   differ per theme, not because this rule is duplicated per theme.

No changes to `ContentPane.tsx`, `theme.ts`, `ThemePicker.tsx`, the Go markdown pipeline, or the
existing `.chroma` rule family — all per Constraints.

### `e2e/lib/` — new WCAG contrast helper

New pure function (co-located with the existing `harness.mjs`/`browser.mjs` helpers, e.g.
`e2e/lib/contrast.mjs`):
```js
export function contrastRatio(colorA, colorB) // accepts "#rrggbb" or "rgb(r,g,b)", returns a number
```
Implements the standard WCAG relative-luminance formula. Small, dependency-free, single
responsibility — a deep module: one exported function, no config object, callers just pass two
colors.

### `e2e/theme-switching/` extensions (not a new scenario)

- **`data/fixture.md`**: add, after the existing h1-h6/paragraph/link/code-block content —
  one bold phrase, one blockquote, one unordered list, one ordered list, one table (with a
  header row), one `<hr>` (`---`), one inline-code span. Headings are untouched.
- **`run`**: for each of the 6 themes (already-looped in Pass B), additionally read:
  - Resolved `--tw-prose-*` custom-property values via
    `getComputedStyle(article).getPropertyValue('--tw-prose-body')` etc. for every var in the
    mapping table above — this is the primary mechanical check and works even though it reads
    a CSS variable rather than an element (so it doesn't depend on fixture markup covering
    every possible sub-element).
  - Real computed `.color`/`.borderColor` on a few representative rendered elements from the
    new fixture content (bold text, blockquote text, list item, table cell border, `<hr>`) as a
    cascade/specificity sanity check — matching the existing h1/h6 pattern already in the
    harness, catching any case where a var resolves correctly but a stray selector still wins
    on the actual element.
  - All 6 `--h1`...`--h6` resolved values (already partially read via the `h1`/`h6` DOM colors
    in the existing Pass B — extend to capture all 6 levels plus the resolved `--text` value,
    needed for the exact-equality and per-level contrast checks).
  Write everything into the existing `computed-styles.json` per theme (extending its shape, not
  adding a new file).
- **`verify`**: extend the existing per-surface distinctness loop with:
  1. **Heading hierarchy**: per theme, assert `h1`...`h6` and `--text` are 6+1 pairwise-distinct
     values (no exact-equality violation), and each of `h1`...`h6` clears `contrastRatio(hN, bg) >= 3`.
  2. **WCAG text-var gate**: per theme, for `body`/`bold`/`quotes`/`captions`/`code`/`pre-code`,
     assert `contrastRatio(value, bg) >= 4.5`.
  3. **WCAG marker-var gate**: per theme, for `bullets`/`counters`, assert
     `contrastRatio(value, bg) >= 3`.
  4. **No gate** for `hr`/`quote-borders`/`th-borders`/`td-borders`/`pre-bg` — only that they
     resolve to a non-null, non-default value per theme (still wired, just not ratio-checked).
  5. Existing distinctness-across-6-themes checks extend to cover the new representative
     elements (bold, blockquote, list marker, table border).

## Action Sequence

1. Hand-tune `--h1`...`--h6` in all 6 `[data-theme='...']` blocks in `web/src/themes.css`
   against the hierarchy rule (no identical levels, no level identical to `--text`, every level
   ≥ 3:1 against that theme's `--bg`). gruvbox and one-dark expected to need little/no change.
2. Add the new theme-agnostic `.prose { --tw-prose-*: var(--...); }` rule block to
   `web/src/themes.css` per the mapping table above.
3. Create `e2e/lib/contrast.mjs`: the `contrastRatio` WCAG relative-luminance function.
4. Extend `e2e/theme-switching/data/fixture.md`: add bold text, a blockquote, an unordered
   list, an ordered list, a table, an `<hr>`, and an inline-code span.
5. Extend `e2e/theme-switching/run`: read resolved `--tw-prose-*` values, resolved
   `--h1`...`--h6`/`--text`, and computed colors on the new representative elements, per theme;
   write into `computed-styles.json`.
6. Extend `e2e/theme-switching/verify`: heading hierarchy check, WCAG text-var gate, WCAG
   marker-var gate, unwired-var presence check, extended distinctness checks — using
   `contrast.mjs`.
7. Run `e2e/theme-switching/test`, inspect failures, iterate on step 1's hex values until every
   check reports good.
8. Refactor — review changed code against meta-pattern and deep-module: readable code, readable
   architecture, clear naming, simple over clever, remove unused files and dead code, flatten
   unnecessary abstractions.
9. Test — run `e2e/theme-switching` and `e2e/tabs-and-split-view` (regression) to verify nothing
   broke.
10. Update `.context/wiki/` via `/to-wiki` — mark
    [[20260712-232400-tw-prose-vars-unwired-to-theme-tokens]] resolved, and capture the
    `--text-muted`/`--border` contrast findings as a new entry (future theme additions need to
    know these tokens aren't safe as-is for WCAG-gated `.prose` content).
11. Write a changelog entry via `/to-changelog`.
12. Remove the originating `.context/TODO.md` item via `/to-todo`, if one exists.

## Test-Loop Design

- **`run`:** Reset `e2e/theme-switching/result/`. Extends the existing 3-pass structure — no
  new scenario directory, no new server/fixture-loading mechanism. Adds: resolved
  `--tw-prose-*` custom-property reads (all 13 wired vars) and resolved `--h1`...`--h6`/`--text`
  values, per theme, during the existing Pass B per-theme loop; computed colors on new
  representative elements (bold, blockquote, list item, table border, `<hr>`) alongside the
  existing h1/h6/chroma reads. Writes into the existing `computed-styles.json`, `metadata.json`,
  `server.log` — no new output files.
- **`verify`:** Reads the extended `computed-styles.json`. New checks: (1) heading
  exact-equality + per-level 3:1 contrast, (2) 4.5:1 gate on
  body/bold/quotes/captions/code/pre-code, (3) 3:1 gate on bullets/counters, (4) presence check
  (non-null, wired) on hr/quote-borders/th-borders/td-borders/pre-bg, (5) extended
  distinctness-across-6-themes on the new representative elements. All existing checks
  (interaction, request-delta, tab/scroll cross-check, reload persistence) run unmodified.
- **Scenarios:**
  - `theme-switching` (extended) → all existing checks pass, plus: every theme's heading set has
    no exact-equality violations and every level clears 3:1; every WCAG-gated `.prose` var
    clears its bar in every theme; every unwired-but-mapped var is present and non-default; the
    new representative elements differ across all 6 themes.
  - `tabs-and-split-view` (regression) → all 11 existing checkpoints still pass unmodified.

## Evaluation Criteria

- **Good:** `e2e/theme-switching test` reports all checks good, including the new WCAG/hierarchy
  checks. `e2e/tabs-and-split-view` still reports all 11 checkpoints good. Manual smoke: open
  each theme in Settings, confirm headings visibly step down in emphasis h1→h6 without any level
  vanishing into body text, and confirm bold text/blockquotes/list markers/table borders/`<hr>`/
  inline code are all legible (not dark-on-dark or invisible) on every dark theme.
- **Ambiguous:** None expected — every criterion in this fix is either a hard WCAG ratio or an
  exact-equality check, both fully mechanical.
- **Bad:** Any Failure Criteria bullet from the direction is observed — a heading identical to
  `--text` or an adjacent level, a heading below 3:1, body/content text below 4.5:1, a `.prose`
  sub-element still falling through to the Typography plugin's raw default, h5/h6 collapsing
  fully onto `--text`, or a fix to one theme regressing another theme's already-passing values.
