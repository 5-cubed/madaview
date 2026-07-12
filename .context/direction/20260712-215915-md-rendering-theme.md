# Markdown Rendering Theme

## Goal
Let a user pick a named visual theme (e.g. GitHub Light, GitHub Dark, Kanagawa, Gruvbox) from Settings that recolors the whole app-facing surface consistently: per-level header colors (h1-h6), code block background and syntax-token colors, tab strip, sidebar file tree, top header bar, and the Settings page itself. The choice applies instantly, with no page reload and no server round-trip.

## Failure Criteria
- Switching themes leaves any themed surface (headers, code blocks, tabs, sidebar tree, header bar, Settings page) still showing the previous theme's colors — a partial/inconsistent switch.
- Code block syntax tokens (keywords, strings, comments) don't recolor with the theme — today they render with zero color at all (see Constraints), so "no visible change" on theme switch for code blocks is a regression, not a neutral outcome.
- Theme switch requires a page reload, or visibly resets in-memory tab/split/scroll state (contradicts the tabs-and-split-view "kept mounted" contract, [[20260712-000044-tabs-and-split-view]]).
- Theme switch triggers a fetch to the Go server (re-rendering markdown, re-highlighting code, or any `/api/*` call) — switching must be a pure client-side CSS change.
- The saved theme doesn't persist across a reload of the same browser (localStorage read failure, wrong key, etc.), forcing the user to re-pick every visit.
- A freshly loaded page shows a jarring flash of the wrong theme (severe FOUC) before settling on the saved one.
- Adding or fixing a theme's palette requires touching more than one place per surface (e.g. header colors defined in three different files) — signals the token model is leaking rather than centralized.

## Ambiguous Zone
- A brief, low-severity flash-of-unstyled-theme on cold page load (before localStorage is read and applied) is acceptable friction for v1, not a failure — full FOUC elimination (e.g. inline blocking script in `index.html`) is a nice-to-have, not required.
- Exact theme names/count beyond "GitHub Light, GitHub Dark, Kanagawa, Gruvbox" as reference points — the final list (4-6 themes) is a /planning-level detail, not blocking here.
- Precise color values per theme/per token (exact hex for h3 vs h4, exact Chroma token-class mapping) are implementation detail for /planning, not modeled in this direction.
- Whether theme data lives as TS objects, JSON files, or CSS files is an implementation choice for /planning.

## Direction
Build a client-side-only theme system:

- **Themes**: a small hardcoded set (~4-6) of named palettes (e.g. GitHub Light, GitHub Dark, Kanagawa, Gruvbox). Each theme defines colors for: h1-h6, code block background/border, Chroma syntax-token classes (`.chroma .kd`, `.nf`, `.p`, `.s`, `.c`, etc. — see Constraints), tab strip (active/inactive/hover), sidebar file tree, top header bar, and the Settings page.
- **Selection UI**: a dropdown/list in Settings (extending the existing page at `web/src/pages/Settings.tsx`) with a live preview sample (small rendered h1-h3 + code block + link) that re-colors as the user browses/selects options, before committing.
- **Persistence**: browser localStorage, per-viewer. Not stored in the Go `Config` struct / `~/.config/madaview/config.json` — this is a personal display preference, not a shared server setting, consistent with multiple LAN devices being able to view the same server with different theme preferences.
- **Application mechanism**: a theme is applied by swapping a data attribute (e.g. `data-theme="kanagawa"` on `<html>` or a root element) that drives CSS custom properties / class-scoped rules. No page reload, no server request — pure client-side CSS swap, instant across all open tabs/panes since it doesn't disturb React component state.
- **No custom/user-authored themes** in v1 — adding a new theme is a code change (new token map), not a user-facing theme-authoring feature.

## Constraints
- **Real bug discovered, now in scope to fix as part of this feature**: code blocks are rendered server-side via Chroma with `chromahtml.WithClasses(true)` (`internal/markdown/codeblock.go:98`), which emits semantic CSS classes (`.chroma`, `.kd`, `.nf`, `.p`, `.s`, `.c`, etc.) into the HTML — but no corresponding CSS stylesheet exists anywhere in the repo (confirmed: no `WriteCSS` call, no `.chroma` rule in any CSS file, no server route serving one). Code blocks currently render as unstyled black text with zero syntax highlighting. This theme feature must supply per-theme CSS rules for these Chroma token classes — there is no "current styled state" to preserve or migrate, only a gap to fill correctly per theme.
- Markdown HTML itself is still rendered server-side (Goldmark + bluemonday, unchanged) — theming only affects CSS applied client-side to the HTML/DOM the server already produces. No changes to `internal/markdown/` rendering logic beyond what's needed to keep Chroma's class-based output as-is (i.e. do not switch to `WithClasses(false)`/inline styles).
- Settings page (`web/src/pages/Settings.tsx`) gains the theme picker alongside the existing root-folder control; no backend API changes required for theme (contrast with root-folder, which does hit `POST /api/root`).
- Existing dark-mode-via-`prefers-color-scheme` Tailwind `dark:` classes are superseded by the theme system for all themed surfaces — GitHub Light/Dark become explicit theme choices rather than the OS deciding.
- Must not regress the tabs-and-split-view in-memory state contract ([[20260712-000044-tabs-and-split-view]]) — theme switching must not remount panes/tabs or reset scroll.

## Out of Scope
- Server-side/shared theme setting persisted in `~/.config/madaview/config.json` (rejected — this is a per-browser preference).
- User-authored or custom theme files/schema (rejected for v1 — deferred, not ruled out forever).
- Re-rendering or re-fetching markdown/code HTML from the Go server when switching themes (rejected — must be a pure client CSS operation).
- Changing Chroma's rendering mode to inline styles (`WithClasses(false)`) (rejected — class-based output is what enables theme-swappable CSS).
- Full FOUC elimination via blocking inline scripts (deferred — minor flash on cold load is accepted friction, not solved now).
- Theming the markdown *content* interactions unrelated to color (e.g. font family, spacing/line-height systems) — this direction is about color palettes only, not typography scale changes.
