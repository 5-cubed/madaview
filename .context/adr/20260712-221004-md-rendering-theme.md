# ADR: Markdown Rendering Theme

**Date:** 2026-07-12

## Context
Direction `.context/direction/20260712-215915-md-rendering-theme.md` asks for a client-side,
instant, no-reload theme picker in Settings that recolors every app-facing surface — markdown
headers (h1-h6), code blocks (background + Chroma syntax tokens), tab strip, sidebar file
tree, top header bar, and Settings itself — from a hardcoded set of named palettes, persisted
per-browser in localStorage.

Two things make this non-trivial rather than a pure CSS reskin:
- **Chroma emits classes with no stylesheet** ([[20260712-220007-chroma-css-classes-never-styled]]):
  `internal/markdown/codeblock.go` renders fenced code via
  `chromahtml.New(chromahtml.WithClasses(true))`, which emits semantic classes (`.chroma`,
  `.kd`, `.nf`, `.p`, `.s`, `.c`, ...) into server-rendered HTML, but no CSS anywhere colors
  them — code blocks render as unstyled black text today. This is a real gap to fill, not a
  migration.
- **The tabs-and-split-view "kept mounted" contract**
  ([[20260712-001324-tabs-and-split-view]]) means theme switching must not remount `Pane`/
  `ContentPane` or touch `Workspace` state — it has to be pure CSS cascade, not a React
  re-render that touches the DOM tree those components own.

`internal/markdown/sanitize.go` already calls `p.AllowAttrs("class").Globally()`, so no
backend sanitizer change is needed for Chroma's classes to survive to the client.

## Decision
A frontend-only theme system, entirely CSS-custom-property-driven:

- **Themes are CSS, not TS/JSON.** Each theme is a `[data-theme="x"] { --var: value; ... }`
  block in a new `web/src/themes.css`. Switching themes is exactly one DOM mutation
  (`document.documentElement.dataset.theme = id`) — the browser's own cascade does the
  recoloring with zero JS re-render and zero risk of disturbing React-owned state. This also
  directly satisfies the direction's "one place per surface" failure criterion: retinting a
  surface means editing that theme's variable value, once.
- **The attribute is set synchronously in `main.tsx`, before `ReactDOM.render`** — the
  earliest possible point in execution, minimizing the (accepted, per the direction's
  Ambiguous Zone) FOUC window on cold load. No new entry file, no blocking inline
  `<script>` in `index.html`.
- **Two token families, one model:**
  - *Chrome tokens* (header bar, sidebar, tab strip, divider, Settings): one shared,
    general-purpose set — `--bg`, `--bg-subtle`, `--border`, `--text`, `--text-muted`,
    `--accent` — reused across every non-markdown surface via Tailwind arbitrary-value
    utilities in JSX (e.g. `border-[var(--border)]`), replacing the `dark:` utility pairs
    that exist today in `App.tsx`, `Sidebar.tsx`, `TabStrip.tsx`, `Pane.tsx`,
    `PaneDivider.tsx`, `Settings.tsx`. These surfaces already share 2-3 shades of neutral
    gray today, so one shared set matches reality and keeps the token count minimal.
  - *Markdown-content tokens* (`--h1`...`--h6`, `--md-link`, `--code-bg`, `--code-border`,
    `--chroma-fg` + 7 category vars): consumed by a small number of theme-agnostic CSS rules
    (`.prose h1 { color: var(--h1); }`, `.chroma [class^="k"] { color:
    var(--chroma-keyword); }`, etc.) written once, not per theme.
- **Chroma's ~80 distinct token classes are covered by ~8 category rules**, not
  enumerated individually. Chroma's classes follow Pygments' short-code convention: the
  first letter names the category (`k*`=keyword, `n*`=name, `s*`=string, `c*`=comment,
  `o*`=operator, `m*`=number, `g*`=generic, `l*`=literal; `p`=punctuation is an exact
  match). CSS attribute-prefix selectors (`.chroma [class^="k"]`) map every sub-class in a
  category to one variable in one rule. A base `.chroma { color: var(--chroma-fg); }`
  covers anything unmatched, so no token is ever left at browser-default black. This stays
  100% frontend-owned (no Go-side `WriteCSS` build step, no dependency on Chroma's bundled
  style palettes matching our theme names).
- **Six themes**: GitHub Light, GitHub Dark, Kanagawa, Gruvbox, Solarized Light, One Dark —
  two light, four dark-leaning, each with a genuinely distinct palette family (neutral gray,
  ink-wash blue/purple, warm earth, saturated warm-light, blue-gray dark).
- **Default on first visit (no localStorage entry): always GitHub Light.** No
  `prefers-color-scheme` detection — simplest correct behavior; the direction explicitly
  supersedes OS-driven dark mode for themed surfaces, so there's no reason to let it pick
  the *initial* choice either.
- **Settings picker is a custom button list, not a native `<select>`.** Native
  `<option>` elements don't reliably fire hover/focus events across browsers, and the
  direction requires the preview to recolor "as the user browses... before committing."
  Each theme renders as a `<button>`; `onMouseEnter`/`onFocus` sets the *preview swatch's*
  own `data-theme` attribute (a small wrapper div containing sample h1-h3 + code block +
  link — CSS custom properties cascade correctly to a nested `data-theme` scope, overriding
  the outer one for just that subtree); `onClick` commits — applies to
  `document.documentElement` and writes localStorage — in the same action. No separate
  "apply"/"save" step, matching the direction's "instant, no round-trip" framing.

## Design

### Token registry — `web/src/themes.css` (new)
Six `[data-theme="..."]` blocks, each defining:
- Chrome (6): `--bg`, `--bg-subtle`, `--border`, `--text`, `--text-muted`, `--accent`
- Headers (6): `--h1` ... `--h6`
- Markdown link (1): `--md-link`
- Code (11): `--code-bg`, `--code-border`, `--chroma-fg` (default/fallback token color) +
  8 category vars: `--chroma-keyword`, `--chroma-name`, `--chroma-string`,
  `--chroma-comment`, `--chroma-number`, `--chroma-operator`, `--chroma-generic`,
  `--chroma-literal`

24 variables × 6 themes. Component-facing CSS rules (theme-agnostic, written once) live in
the same file or appended to `web/src/index.css`:
```css
.prose h1 { color: var(--h1); } /* ...h2..h6 */
.prose a { color: var(--md-link); }
.chroma { background: var(--code-bg); border-color: var(--code-border); color: var(--chroma-fg); }
.chroma [class^="k"] { color: var(--chroma-keyword); }
.chroma [class^="n"] { color: var(--chroma-name); }
.chroma [class^="s"] { color: var(--chroma-string); }
.chroma [class^="c"] { color: var(--chroma-comment); }
.chroma [class^="m"] { color: var(--chroma-number); }
.chroma [class^="o"] { color: var(--chroma-operator); }
.chroma [class^="g"] { color: var(--chroma-generic); }
.chroma [class^="l"] { color: var(--chroma-literal); }
.chroma .p { color: var(--chroma-fg); }
```

### Application module — `web/src/theme.ts` (new)
Deep module, small interface:
```ts
export interface ThemeMeta { id: string; label: string }
export const THEMES: ThemeMeta[] // the 6 entries, id matches data-theme value
export const DEFAULT_THEME = 'github-light'
export function getStoredTheme(): string   // reads localStorage, falls back to DEFAULT_THEME
export function applyTheme(id: string): void // sets document.documentElement.dataset.theme + writes localStorage
```
`getStoredTheme`/`applyTheme` are the only two functions anything outside this file ever
calls — `main.tsx` calls both (read then apply) once at startup; the Settings picker calls
only `applyTheme` on click.

### Wiring — `main.tsx`
```ts
import { applyTheme, getStoredTheme } from './theme'
applyTheme(getStoredTheme()) // before createRoot(...).render(...)
```

### Settings picker — `web/src/pages/Settings.tsx` (extended)
A `ThemePicker` section: `THEMES.map(...)` renders one button per theme; a preview wrapper
div (sample `<h1>`/`<h2>`/`<h3>` + a small fenced-code sample + a link) carries its own
`data-theme` attribute, updated `onMouseEnter`/`onFocus` per button and reset to the
committed theme `onMouseLeave`. `onClick` calls `applyTheme(id)` and updates local React
state (for the "currently selected" highlight) — no network call, consistent with the
existing Settings page's other non-networked concerns.

### Chrome surfaces — `App.tsx`, `Sidebar.tsx`, `TabStrip.tsx`, `Pane.tsx`,
`PaneDivider.tsx`, `Settings.tsx`
Every `<color>-<shade> dark:<color>-<shade>` pair (17 occurrences across 7 files, confirmed
via `grep -rl "dark:" web/src`) becomes a single arbitrary-value utility against the shared
chrome tokens, e.g. `border-neutral-200 dark:border-neutral-800` → `border-[var(--border)]`.
`ContentPane.tsx`'s `prose dark:prose-invert` becomes plain `prose` (color now governed by
the `.prose h1..h6`/`.prose a` rules above, not Tailwind Typography's own dark variant).

## Action Sequence
1. Create `web/src/themes.css`: the 6 theme token blocks (24 vars each) + the theme-agnostic
   `.prose h*`/`.prose a`/`.chroma`-family component rules; import it from `main.tsx`
   alongside `index.css`.
2. Create `web/src/theme.ts`: `ThemeMeta`, `THEMES`, `DEFAULT_THEME`, `getStoredTheme`,
   `applyTheme`.
3. Wire `main.tsx`: call `applyTheme(getStoredTheme())` synchronously before
   `createRoot(...).render(...)`.
4. Replace `dark:` utility pairs with `var(--...)` arbitrary-value utilities in `App.tsx`,
   `Sidebar.tsx`, `TabStrip.tsx`, `Pane.tsx`, `PaneDivider.tsx`, `Settings.tsx` (chrome
   tokens only).
5. Update `ContentPane.tsx`: drop `dark:prose-invert`, keep `prose`.
6. Build the `ThemePicker` UI in `Settings.tsx`: button list + preview swatch
   (sample h1-h3 + code block + link) with hover/focus-driven preview `data-theme` and
   click-to-commit `applyTheme`.
7. Add `e2e/theme-switching/` scenario (`data/fixture.md` covering h1-h6, a code block with
   keyword/string/comment/name/number tokens, and a link; `run`, `verify`, `test`) per the
   Test-Loop Design below.
8. Refactor — review changed code against meta-pattern and deep-module: readable code,
   readable architecture, clear naming, simple over clever, remove unused files and dead
   code, flatten unnecessary abstractions.
9. Test — run the test-loop (`e2e/theme-switching`, plus `e2e/tabs-and-split-view` as a
   regression check) to verify nothing broke.
10. Update `.context/wiki/` via `/to-wiki` if new truths were found or existing entries are
    stale (the Chroma-unstyled wiki entry [[20260712-220007-chroma-css-classes-never-styled]]
    should be marked resolved).
11. Write a changelog entry via `/to-changelog`.
12. Remove the originating `.context/TODO.md` item via `/to-todo` — none exists for this
    direction, so this step is a no-op.

## Test-Loop Design
No existing scenario covers palette/color correctness — `render-gfm-basics` and
`render-korean-content` check structural HTML, not computed styles. New scenario required;
extends the harness conventions already used by `tabs-and-split-view` (`resetResultDir`,
`startServer`, `withPage`, `writeJSON`/`writeMetadata`).

- **`run`:** Reset `e2e/theme-switching/result/`. Start the server against
  `data/fixture.md` (h1 through h6, a fenced code block in a language exercising
  keyword/string/comment/name/number tokens, and one inline link). Drive Playwright:
  1. Load the page, open `fixture.md`.
  2. For each of the 6 themes in turn: click its button in Settings' `ThemePicker`, then
     `page.evaluate` to read `getComputedStyle` color on one representative element per
     surface — an `h1`, an `h6`, a `.chroma [class^="k"]` span, a `.chroma [class^="s"]`
     span, the tab strip root, the sidebar nav, the header bar, the Settings page root —
     and record into `result/computed-styles.json` keyed by theme id. Also record the
     cumulative `/api/*` request count after each switch.
  3. Cross-check against the tabs-and-split-view contract: before switching themes, open a
     second tab and scroll the first tab's content; after cycling through all 6 themes,
     assert tab count, active tab id, and scrollTop are unchanged and no new `/api/file`
     request fired for any switch.
  4. Reload the page after landing on the last-selected theme; assert `localStorage` still
     reports that theme and the applied `data-theme` attribute matches it (persistence
     check).
  Write `server.log`, `metadata.json` (scenario name, root, port, theme list) alongside
  `computed-styles.json`.

- **`verify`:** Reads `computed-styles.json`. For every surface (h1, h6, keyword-token,
  string-token, tab strip, sidebar, header bar, Settings), asserts the recorded color
  differs across all 6 themes pairwise (a surface stuck on one color across every theme is
  the "partial/inconsistent switch" failure). Asserts no keyword/string/comment token color
  ever equals the literal unstyled-black regression color. Asserts the request-count delta
  is zero across every theme switch (no server round-trip). Asserts the tab/scroll
  cross-check and reload-persistence checks from `run` both hold. Reports good/unexpected/
  ambiguous per check, with root cause traced through `server.log` for any unexpected
  network activity.

- **Scenarios:**
  - `theme-switching` → all per-surface/per-theme distinctness checks, zero-network-delta
    checks, tab/scroll-preservation cross-check, and reload-persistence check all pass.
  - `tabs-and-split-view` (regression) → all 11 existing checkpoints still pass unmodified.

## Evaluation Criteria
- **Good:** `e2e/theme-switching` reports all checks good — every themed surface's computed
  color changes across all 6 themes, no theme leaves a Chroma token at the old unstyled-
  black color, zero `/api/*` calls fire on any switch, the tab/scroll state opened before
  switching is byte-identical after cycling all 6 themes, and the last-selected theme
  survives a reload. `e2e/tabs-and-split-view` still reports all 11 checkpoints good (no
  regression). Manual smoke: open Settings, hover each theme button and see the preview
  swatch recolor before clicking, click through all 6, confirm the whole visible app
  (header, sidebar, tabs, code block, headers, Settings) recolors together with no flicker
  of mixed old/new colors and no console errors.
- **Ambiguous:** A brief flash of the previous theme's colors on a cold page load, before
  `main.tsx`'s synchronous `applyTheme` call runs — accepted per the direction's own
  Ambiguous Zone (full FOUC elimination via a blocking inline script is explicitly a
  nice-to-have, not required here).
- **Bad:** Any Failure Criteria bullet from the direction is observed — a surface stuck on
  the old theme, a Chroma token still unstyled, a reload or theme switch resetting tab/pane/
  scroll state, any `/api/*` call firing on switch, the saved theme not surviving a reload,
  severe FOUC, or a palette fix requiring edits in more than one place per surface.
