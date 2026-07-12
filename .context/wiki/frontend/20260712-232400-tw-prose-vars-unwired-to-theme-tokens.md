# `.prose` content is only themed where themes.css explicitly overrides Typography plugin vars

**Resolved** by `.context/adr/20260712-233457-theme-contrast-hierarchy-fix.md` — see
[[20260712-233500-text-muted-and-border-not-safe-for-wcag-gated-prose]] for the follow-on
finding about which theme tokens were safe to reuse for the fix.

`web/src/components/ContentPane.tsx` renders markdown with plain `className="prose max-w-none"` — no `prose-invert`, no `dark:` variant. `@tailwindcss/typography` styles everything inside `.prose` (body paragraphs, bold text, blockquotes, table borders, list bullets/counters, captions, `<hr>`, inline code, pre blocks) through its own CSS custom properties (`--tw-prose-body`, `--tw-prose-bold`, `--tw-prose-quotes`, `--tw-prose-th-borders`, `--tw-prose-bullets`, `--tw-prose-counters`, `--tw-prose-captions`, `--tw-prose-hr`, `--tw-prose-pre-*`, etc.), which default to a hardcoded light-mode palette baked into the plugin itself.

`web/src/themes.css`'s per-theme `[data-theme='...']` blocks only ever overrode `--h1`..`--h6` and `--md-link` via direct selectors (`.prose h1`, `.prose a`). They never touched the Typography plugin's own `--tw-prose-*` vars. Result: on every dark theme, ordinary paragraph/list/blockquote/table text rendered in the plugin's dark near-black default — invisible against a dark `--bg` — even though the theme's own `--text`/`--text-muted` tokens already had good contrast and were simply unused by `.prose` sub-elements.

**Takeaway**: overriding `.prose h1`-style selectors directly is not sufficient to theme a `.prose` block — anything the Typography plugin styles via its own `--tw-prose-*` vars needs those vars explicitly remapped to theme tokens per `[data-theme]` block, or it silently falls back to the plugin's built-in light palette regardless of what theme is active. See [[20260712-221510-nested-data-theme-scopes-cascade-correctly]] for how the `data-theme` cascade itself works.

**Fix applied**: a single theme-agnostic `.prose { --tw-prose-*: var(--token); }` block was added to
`web/src/themes.css` (once, not per theme) mapping `body`/`headings`/`bold`/`quotes`/`captions`/
`code`/`pre-code`/`bullets`/`counters` to `--text`, `hr`/`quote-borders`/`th-borders`/`td-borders` to
`--border`, and `pre-bg` to `--code-bg`. `--tw-prose-links` was left unwired since the existing
`.prose a` selector already wins on specificity; `--tw-prose-kbd`/`--tw-prose-kbd-shadows`/
`--tw-prose-lead` were left unwired since GFM markdown can't produce `<kbd>` or `class="lead"` without
raw HTML.

**Correction (same day)**: the first pass of this fix also left `--tw-prose-headings` unwired, on the
theory that `.prose h1`..`.prose h6` already win on specificity — true for the heading elements
themselves, but the Typography plugin *also* reads `--tw-prose-headings` for `thead th` (table header
cells), which has no such override. Result: table header row text rendered in the plugin's dark
default, invisible against dark-theme backgrounds. `--tw-prose-headings` is now wired to `--text`
alongside the other text-level vars, and `e2e/theme-switching` gained a `tableHeader` representative
check (4.5:1 gate) covering it. **Lesson**: don't reason about a `--tw-prose-*` var from "which elements
I intended it for" — grep the Typography plugin's own `styles.js` for every selector that reads it
before deciding a var is safely covered by another override.
