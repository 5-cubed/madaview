# ADR: Korean Content Rendering

**Date:** 2026-07-11

## Context
Korean text inside markdown documents viewed through madaview (prose, headings, fenced code
blocks, Mermaid diagram labels) and Korean-named files/folders in the sidebar must render
correctly on Windows/macOS/Linux, using the OS's own installed Korean fonts — no bundled webfont,
no network calls. Full goal, failure criteria, and constraints live in
`.context/direction/20260711-191851-korean-content-rendering.md`. This is content-rendering
fidelity, not UI localization — madaview's interface strings stay English.

Two prior wiki entries already resolved parts of this: font strategy is OS-fallback, not bundling
([[20260711-191934-cjk-font-strategy-os-fallback]]), and Korean filenames already round-trip
correctly through the API — `web/src/api.ts` wraps `path` in `encodeURIComponent` for both
`/api/tree` and `/api/file` — so no server or `api.ts` change is needed
([[20260711-191934-korean-filenames-already-work]]).

Current state confirmed by reading the code: `web/src/index.css` declares no explicit
`font-family` anywhere — body, sidebar, and content pane all inherit Tailwind v4's default sans
stack via preflight. `web/src/hydrate.ts` calls `mermaid.initialize({ startOnLoad: false })` with
no `fontFamily` option. `web/src/components/Sidebar.tsx` builds nav links directly from
`entry.path` (`/view/${entry.path}`) with no extra encoding — relies on the browser's native
UTF-8 URL handling, consistent with the wiki's filename finding.

## Decision

### Font-stack wiring (CSS)
Override Tailwind v4's `--font-sans` theme variable in `web/src/index.css` via an `@theme` block,
prepending Korean-capable system font names ahead of Tailwind's existing default sans stack:
`"Apple SD Gothic Neo", "Malgun Gothic", "Noto Sans KR"`. This is a single override point — every
element that inherits the default font via preflight, or uses the `font-sans` utility explicitly,
picks it up automatically. No new CSS selectors, no risk of losing to Tailwind's own cascade-layer
ordering (which a plain `body { font-family: ... }` rule would have to contend with).

Fenced code blocks and KaTeX remain untouched, per the direction — default browser per-glyph
substitution already covers code blocks; KaTeX has no Korean text path.

### Font-stack wiring (Mermaid)
`hydrate.ts` reads the resolved stack at runtime via `getComputedStyle(document.body).fontFamily`
and passes it straight into `mermaid.initialize({ fontFamily })`, called once before any
`mermaid.render()`. CSS remains the single source of truth for the font list; there is nothing to
manually keep in sync between the CSS and the Mermaid config, so the two can never drift apart.

### Sidebar filenames
No code change. `entry.path` flowing into `Link to={/view/${entry.path}}` and back through
`fetchFile`'s `encodeURIComponent` already round-trips UTF-8 correctly (browsers percent-encode
non-ASCII path segments on navigation and React Router decodes them back before handing off
`params['*']`). This ADR's only sidebar-related work is the shared font-stack override above,
which the sidebar inherits like every other element.

### Test-loop: `e2e/render-korean-content/`
New scenario (existing scenarios don't cover Korean-specific fixtures or assertions — structurally
justified per test-loop reuse-before-creation policy). Follows the existing `run`/`verify`/`test`
pattern (see `e2e/render-mermaid/` for reference), using `e2e/lib/harness.mjs` and
`e2e/lib/browser.mjs` unchanged.

**Fixture** (`e2e/render-korean-content/data/`):
- `fixture.md` at the fixture root containing: a Korean heading, a Korean prose paragraph, a
  fenced code block with a Korean comment, and a Mermaid flowchart with one Korean-labeled node.
- A Korean-named folder (e.g. `문서/`) containing a Korean-named file (e.g. `문서/안내.md`) with
  its own short Korean heading/body — exercises a nested sidebar expand (`/api/tree` with a Korean
  path segment) followed by a file click (`/api/tree` + `/api/file` on a nested Korean path), not
  just a flat top-level file.

**Glyph-rendering check (tofu detection):** cannot be done via DOM text inspection alone — tofu
boxes are a font-fallback rendering artifact, not a distinct Unicode codepoint, so correct
codepoints being present in the DOM doesn't prove they rendered as real glyphs. Verify instead
using the FontFaceObserver technique: for each of the four text surfaces (prose, heading, Mermaid
label, code block), run an in-page script that measures the Korean string's width twice via
`canvas.measureText()` — once with the element's actual computed `font`, once with the same
string forced onto a font stack known to lack Hangul coverage (e.g. `'monospace'` alone, no
fallback list). If the two measured widths differ, a distinct Korean-capable font was actually
selected to render the string, not the generic notdef/tofu glyph. This is deterministic given a
Korean-capable font is present on the machine running the browser — which is exactly why CI needs
one (see below).

Mermaid's node-label text is DOM (`foreignObject > div > span > p`), not raw SVG `<text>` — same
`getComputedStyle` + canvas technique applies without special-casing SVG.

Also assert, per the direction's failure criteria: no U+FFFD replacement characters and no
tofu-box codepoints appear in any rendered text node (a cheap early-exit check before the
measureText check); requests.json shows zero non-`localhost` network requests during Mermaid
hydration (existing pattern from `render-mermaid`); the nested Korean-named file is clickable and
its content loads (200, correct rendered title) — not just an HTTP-status check.

### CI: install a Korean-capable font for the runner
`ubuntu-latest` has no CJK fonts by default, so the measureText check above would be structurally
unable to pass in CI without one. Add a step to `.github/workflows/ci.yml`, before "Install
Playwright browsers": `sudo apt-get update && sudo apt-get install -y fonts-nanum`. This is CI
test infrastructure, not a shipped artifact — it does not violate the "no bundled font in the
binary" constraint, since nothing is embedded in the `madaview` binary or `web/dist`. The exact
installed family name doesn't need to match `"Noto Sans KR"` literally: browsers do automatic
per-glyph system-fontconfig fallback for codepoints missing from every named font in the stack
(the same mechanism the direction already relies on for code blocks), so any installed
Korean-capable font makes the measureText check pass regardless of its family name.

## Design
```
web/src/index.css     @theme override: --font-sans prepends Korean font names
web/src/hydrate.ts     mermaid.initialize({ fontFamily: getComputedStyle(document.body).fontFamily })
                        called once, before hydrateMermaid's render loop
e2e/render-korean-content/
  data/fixture.md       heading, prose, code block, mermaid — all containing Korean text
  data/문서/안내.md      nested Korean-named folder + file
  run                   starts server rooted at data/, drives browser: loads fixture.md,
                         expands 문서/ in sidebar, clicks 안내.md, captures dom.html +
                         requests.json + per-surface canvas-measured widths + server.log
  verify                checks: no U+FFFD/tofu codepoints, measured-width divergence per
                         surface (prose/heading/mermaid-label/code-block), zero non-localhost
                         requests, nested file navigation reaches 200 + correct title
.github/workflows/ci.yml   + apt-get install fonts-nanum step before Playwright browser install
```

## Action Sequence
1. Add `@theme { --font-sans: ... }` override to `web/src/index.css` prepending
   `"Apple SD Gothic Neo", "Malgun Gothic", "Noto Sans KR"` ahead of Tailwind's existing default
   sans stack.
2. Update `hydrate.ts`: read `getComputedStyle(document.body).fontFamily` and pass it to
   `mermaid.initialize({ fontFamily })` before any `mermaid.render()` call.
3. Add `sudo apt-get update && sudo apt-get install -y fonts-nanum` step to
   `.github/workflows/ci.yml`, before the "Install Playwright browsers" step.
4. Create `e2e/render-korean-content/data/fixture.md` (Korean heading, prose paragraph, fenced
   code block with a Korean comment, Mermaid flowchart with one Korean-labeled node).
5. Create `e2e/render-korean-content/data/문서/안내.md` (nested Korean-named folder + file with
   its own short Korean heading/body).
6. Write `e2e/render-korean-content/run`: start the server rooted at `data/`, drive a Playwright
   page through `fixture.md`, expand the `문서` sidebar entry, click into `안내.md`, and write
   `dom.html`, `requests.json`, a `measured-widths.json` (per-surface actual-vs-fallback canvas
   measurements), and `server.log` to the result dir, plus `metadata.json`.
7. Write `e2e/render-korean-content/verify`: fail on any U+FFFD/tofu codepoint in rendered text,
   fail on any surface where actual-vs-fallback measured width is equal (no real glyph rendered),
   fail on any non-`localhost` request, fail if the nested Korean file doesn't return 200 with the
   correct rendered title; otherwise report good.
8. Write `e2e/render-korean-content/test` (thin `runThenVerify` wrapper, matching every other
   scenario).
9. Run `go vet ./...`, `go test ./...`, and the full e2e suite locally (or via CI) to confirm no
   regressions in the existing 8 scenarios and that the new scenario passes.
10. Write domain truths or key decisions to the wiki via `/to-wiki` (the `--font-sans` override
    approach, the measureText tofu-detection technique, and the CI font-install rationale).
11. Write a changelog entry via `/to-changelog`.
12. Remove the originating `.context/TODO.md` item via `/to-todo`, if one exists.

## Test-Loop Design
- **`run`:** resets `e2e/render-korean-content/result/` (delete + recreate), starts the real
  `madaview` binary rooted at `e2e/render-korean-content/data/` on a fresh port, drives a
  Playwright page (recording all outgoing request URLs) through: load `fixture.md`, expand `문서`
  in the sidebar, click `안내.md`. Writes to the result dir: `dom.html` (rendered article
  innerHTML after both navigations), `requests.json` (all recorded request URLs),
  `measured-widths.json` (canvas-measured actual vs. fallback-font width per text surface:
  prose, heading, mermaid-label, code-block), `server.log`, and `metadata.json` (version/commit,
  fixture root, port, timestamp, OS/arch).
- **`verify`:** reads all of the above from the result dir. Checks, per scenario: (1) no
  U+FFFD or known tofu-box codepoints in `dom.html` text nodes, (2) every entry in
  `measured-widths.json` shows actual width ≠ fallback width (proves a real Korean-capable font
  rendered the glyphs, not a notdef box), (3) every URL in `requests.json` is `localhost`-only,
  (4) the nested-file navigation reached the expected content (200 status implied by successful
  DOM capture + correct title text present). Reports good / unexpected / ambiguous with root
  cause traced through `server.log` and `dom.html`.
- **Scenarios:** `render-korean-content` → Korean prose, heading, code-block comment, and
  Mermaid node label all render as real glyphs (not tofu, not replacement chars); the nested
  Korean-named folder/file navigates correctly end-to-end; zero non-localhost network requests
  fire during the whole flow.

## Evaluation Criteria
- **Good:** the new `render-korean-content` scenario passes in CI (with `fonts-nanum` installed)
  and locally on macOS/Windows dev machines (which already have Korean system fonts per the
  wiki's OS-fallback precedent) — all four text surfaces show actual-vs-fallback width divergence,
  no U+FFFD/tofu codepoints anywhere, the nested Korean file/folder is clickable and its content
  loads with the correct title, and zero non-`localhost` requests fire. All 9 e2e scenarios
  (8 existing + this one) and all Go unit tests continue to pass.
- **Ambiguous:** a bare-minimum Linux environment with no CJK font installed shows tofu for
  Korean text — accepted per the direction's Ambiguous Zone as an OS-environment gap, not a
  madaview defect (this is why CI specifically installs a font rather than relying on the runner's
  default state — CI must actually exercise the "font is present" path to be a meaningful check
  at all).
- **Bad:** any tofu box or U+FFFD replacement character rendered in body text, headings, code
  blocks, or Mermaid labels on a machine that has a Korean-capable font installed. Mermaid SVG
  text renders Korean labels as boxes because `mermaid.initialize` wasn't given the font-family.
  A Korean filename/foldername breaks navigation (garbled URL, 404, or mis-decoded round-trip).
  Any network call fired during rendering/hydration (CDN font load or otherwise). The new e2e
  scenario passes without actually checking glyph presence (e.g. only an HTTP 200 check) — this
  ADR's whole test-loop design exists specifically to avoid that failure mode.
