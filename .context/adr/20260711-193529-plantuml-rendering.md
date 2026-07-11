# ADR: PlantUML Diagram Rendering

**Date:** 2026-07-11

## Context
The initial ADR (`.context/adr/20260711-173722-madaview-initial-setup.md`) explicitly deferred
PlantUML, reasoning it "typically needs a JVM or a server round-trip." That reasoning was
incomplete: PlantUML ships an official TeaVM-compiled, fully client-side build. Full goal and
failure criteria live in `.context/direction/20260711-192022-plantuml-rendering.md`; the
technical premise was previously captured in
`.context/wiki/20260711-192124-plantuml-client-side-rendering-exists.md`. This ADR reverses the
deferral and adds PlantUML rendering using the exact architectural pattern already proven for
Mermaid and KaTeX (`internal/markdown/codeblock.go`, `internal/markdown/math.go`,
`web/src/hydrate.ts`).

**Verification performed during planning** (not just trusting the wiki note): the wiki's assumed
package names (`plantuml-core`, `plantuml.js`) do not exist on the npm registry. The real,
current package is **`@plantuml/core`** (npm, MIT-licensed since 1.2026.6, zero deps, ~10.6MB
unpacked: `plantuml.js` the TeaVM engine + `viz-global.js` the Graphviz/Viz.js WASM layout
engine). Its public API is two functions from the `plantuml.js` ES module: `render(lines,
targetId, opts?)` (writes SVG directly into a DOM element by id) and `renderToString(lines,
onSuccess, onError)` (delivers the SVG as a string via callback).

Three things were empirically confirmed by building a throwaway Vite project and driving it with
Playwright (not assumed from docs):
1. **Zero network calls at runtime.** The Graphviz WASM binary is embedded as a base64 data URI
   inside `viz-global.js` itself (Emscripten's inline-wasm mode) — instantiated locally, never
   fetched. A real diagram render fired zero requests beyond the app's own bundled JS.
2. **The ES-import bundling path works end-to-end**, despite the package README recommending a
   classic `<script>` tag for `viz-global.js` (its UMD wrapper does a `typeof exports` check that
   could misbehave under a bundler). A production `vite build` with `import
   '@plantuml/core/viz-global.js'` followed by `import { renderToString } from '@plantuml/core'`
   in the same module was tested in a real headless browser: rendering succeeded, zero network
   calls fired, no page errors — confirmed correct despite `globalThis.Viz` not literally being
   set post-bundle (Vite's module linking wires the reference internally rather than through the
   naive global).
3. **`onError` essentially never fires for invalid PlantUML syntax.** PlantUML's engine follows
   its decades-old convention of always returning *some* SVG — even for empty input, garbage
   input, or missing `@startuml` — with the error rendered as text baked into the image itself.
   Two distinct native error-card templates were found by testing empty/garbage/broken input:
   - Missing/unrecognized diagram-type directive → `"Diagram not supported by this release of
     PlantUML ... is not recognized."` (only reachable pre-auto-wrap; see Decision below).
   - Content wrapped in `@startuml`/`@enduml` but invalid inside → ends in `"(Assumed diagram
     type: ...)"` (e.g. `"Syntax Error? (Assumed diagram type: sequence)"`, `"Empty description
     (Assumed diagram type: sequence)"`). This card also carries an unrelated, ugly version-nag
     banner ("This version of PlantUML is 173 days old...plantuml.com/download") because this
     npm build ships with its version placeholders (`$version$`/`$git.commit.id$`) unstamped.

## Decision
Bundle `@plantuml/core` into the frontend build and extend the already-proven Mermaid/KaTeX
placeholder-emit + client-side-hydrate pattern to PlantUML, with these specifics resolved during
grilling:

- **Fence recognition (server, `internal/markdown/codeblock.go`)**: `plantuml`, `puml`, and `uml`
  fence languages are all treated as aliases of one PlantUML placeholder, exactly mirroring the
  existing `language == "mermaid"` branch. The raw, unmodified, HTML-escaped fence source is
  emitted into `<div class="plantuml">...</div>` — no auto-wrap, no error detection, no rendering
  logic server-side. The server stays purely mechanical, identical in spirit to the mermaid path.
- **Auto-wrap (client, in hydration)**: if the placeholder's raw source (trimmed) does not already
  start with `@start`, hydration wraps it in `@startuml` / `@enduml` before rendering — removes
  boilerplate for the common case (sequence/class/etc. diagrams), matching Mermaid's zero-
  boilerplate fences. Content that already provides its own `@start...`/`@end...` pair (e.g.
  `@startmindmap`) is passed through unmodified.
- **API choice**: `renderToString`, Promise-wrapped, to mirror `hydrateMermaid`'s existing
  async/await + manual `innerHTML` assignment shape exactly (`render(lines, targetId)` was
  rejected — it diverges from the established catch-around-an-awaited-call pattern used for both
  Mermaid and KaTeX).
- **`viz-global.js` loading**: a static ES side-effect import in `hydrate.ts` —
  `import '@plantuml/core/viz-global.js'` above `import { renderToString } from '@plantuml/core'`
  — verified to work under a real Vite production build (see Context above). No lazy-loading,
  consistent with the direction's accepted bundle-size-growth tradeoff.
- **Error detection**: because `onError` doesn't fire for syntax errors, hydration inspects the
  resolved SVG string for PlantUML's own native error-card signatures using two patterns (covering
  both templates found during testing, not just the version-banner text which is a packaging
  quirk that could disappear in a future release):
  - `/Diagram not supported by this release of PlantUML/`
  - `/\(Assumed diagram type:/`

  If either matches, the SVG is discarded and a plain-text `PlantUML render error: <message>` is
  shown instead (message extracted from the matched summary line, falling back to a generic
  message if extraction fails) — true visual parity with Mermaid's `Mermaid render error:
  ${message}` pattern, and it hides the irrelevant version-nag banner. A genuine `onError` /
  Promise-rejection (rare JS-level failure, e.g. WASM instantiation failure) is caught the same
  way Mermaid's catch block already works.
- Bundle size growth (~8.2MB raw / ~2MB gzip for `plantuml.js` + `viz-global.js` combined,
  confirmed via test build) is accepted per the direction's Ambiguous Zone — no code-splitting.

## Design
```
internal/markdown/codeblock.go
  isPlantumlFence(language string) bool   — true for "plantuml" | "puml" | "uml"
  writePlantumlPlaceholder(w, code)       — <div class="plantuml">{escaped raw source}</div>
  codeBlockRenderer.render(...)           — branches to writePlantumlPlaceholder alongside the
                                             existing writeMermaidPlaceholder branch, before
                                             falling through to chroma highlighting

web/src/hydrate.ts
  hydratePlantuml(container)              — queries div.plantuml, for each block:
                                             1. read raw text source
                                             2. auto-wrap with @startuml/@enduml if not already
                                                start-tagged
                                             3. renderToString(lines) wrapped in a Promise
                                             4. on resolve: sniff SVG for native error-card
                                                signatures → plain-text error if matched,
                                                else block.innerHTML = svg
                                             5. on reject: plain-text error (same shape)
  hydrate(container)                      — Promise.all extended to include hydratePlantuml
                                             alongside hydrateMermaid / hydrateKatex

web/package.json
  dependencies: "@plantuml/core": "^1.2026.6"
```
Data flow mirrors the existing pipeline exactly: `/api/file` → goldmark (GFM + chroma +
Mermaid/KaTeX/PlantUML placeholder emission) → bluemonday sanitize → HTML string → client injects
into content pane → client hydrates Mermaid/KaTeX/PlantUML placeholders in-browser, in that order,
via `Promise.all`.

## Action Sequence
1. Add `"@plantuml/core": "^1.2026.6"` to `web/package.json` dependencies; `npm install` in `web/`.
2. In `internal/markdown/codeblock.go`: add `isPlantumlFence(language string) bool` (matches
   `plantuml`, `puml`, `uml`) and `writePlantumlPlaceholder(w util.BufWriter, code []byte)`
   (mirrors `writeMermaidPlaceholder`, emits `<div class="plantuml">...</div>`); branch
   `codeBlockRenderer.render` to call it when `isPlantumlFence(language)` is true, alongside the
   existing mermaid branch, before the chroma highlighting fallback.
3. Add Go tests in `internal/markdown/markdown_test.go` mirroring
   `TestRender_MermaidFenceEmitsPlaceholderWithRawSource`: one table-driven test (or three cases)
   asserting each of `plantuml`, `puml`, `uml` fences emit `<div class="plantuml">` with raw
   unescaped-for-markup source and are never highlighted/pre-rendered server-side.
4. In `web/src/hydrate.ts`: add the static side-effect import of `@plantuml/core/viz-global.js`
   and the `renderToString` named import from `@plantuml/core`.
5. Implement `hydratePlantuml(container)` in `web/src/hydrate.ts`: auto-wrap detection, Promise-
   wrapped `renderToString` call, native-error-signature sniffing (the two regex patterns above)
   with plain-text fallback, `innerHTML` assignment on real success — following the exact
   structure of `hydrateMermaid`.
6. Extend `hydrate()`'s `Promise.all([...])` in `web/src/hydrate.ts` to include
   `hydratePlantuml(container)`.
7. Create `e2e/render-plantuml/` following the `e2e/render-mermaid/` scenario structure (`run`,
   `verify`, `test`, `data/`).
8. Write `e2e/render-plantuml/data/fixture.md`: three fenced blocks using `plantuml`, `puml`, and
   `uml` respectively, each a small valid diagram (e.g. sequence), to assert all three aliases
   render identically; plus one fenced block with intentionally invalid PlantUML syntax.
9. Write `e2e/render-plantuml/run`: mirrors `render-mermaid/run` — start the server against the
   fixture root, load the page, wait for `div.plantuml svg`, capture the rendered `<article>`
   innerHTML, capture all fired request URLs, write both plus metadata to the result dir.
10. Write `e2e/render-plantuml/verify`: assert (a) all three alias fences produced a `<div
    class="plantuml">...<svg` match, (b) the invalid-syntax fence produced a `PlantUML render
    error:` text node and *not* an embedded `<svg>` and *not* PlantUML's native error-card
    markers, (c) zero requests outside `http://localhost:` fired, (d) the rest of the document
    (a GFM heading placed after the invalid block, included in the fixture) still rendered
    normally — reports good/unexpected exactly like `render-mermaid/verify`.
11. Write `e2e/render-plantuml/test` (identical boilerplate to the other scenarios' `test`,
    calling `runThenVerify`).
12. Run `go test ./...`, the new e2e scenario, and the full existing e2e suite to confirm no
    regressions.
13. Write domain truths / key decisions to the wiki via `/to-wiki` (the real package name and API,
    the zero-network-call verification method, the auto-wrap rule, and the error-signature
    detection patterns — all non-obvious and expensive to re-derive).
14. Write a changelog entry via `/to-changelog`.
15. Remove the originating `.context/TODO.md` item via `/to-todo`, if one exists.

## Test-Loop Design
- **`run`:** reset by deleting `e2e/render-plantuml/result` if present, start the built `madaview`
  binary against `e2e/render-plantuml/data/` as root, launch a Playwright page, navigate to the
  fixture file, wait for `div.plantuml svg` to appear, capture the rendered `<article>` innerHTML
  and every fired request URL, write both plus run metadata (madaview version/commit, fixture
  root, port, timestamp) to the result directory — identical shape to `render-mermaid/run`.
- **`verify`:** reads `dom.html` and `requests.json` from the result dir; checks (1) each of the
  three alias fences rendered a real `<svg>` inside its `div.plantuml`, (2) the invalid-syntax
  fence shows plain-text `PlantUML render error:` and neither an `<svg>` nor PlantUML's own
  native error-card text, (3) no request URL falls outside `http://localhost:`, (4) a GFM heading
  placed after the invalid block in the fixture is still present in the DOM (proves one bad
  diagram doesn't take down the rest of the document) — reports good/unexpected/ambiguous with
  root cause traced through `dom.html`/`requests.json`, matching the existing scenarios' report
  format.
- **Scenarios:** `render-plantuml` → all three fence aliases (`plantuml`, `puml`, `uml`) hydrate to
  real `<svg>` elements with zero external network requests; an invalid-syntax fence in the same
  document shows a plain-text inline error (not a crash, not raw code, not PlantUML's own noisy
  error card) while the rest of the document keeps rendering normally.

## Evaluation Criteria
- **Good:** `go test ./...` passes with the new placeholder-emission tests; the `render-plantuml`
  e2e scenario passes; a markdown file with any of the three fence aliases renders a correct
  diagram in-browser with zero network requests (verifiable by opening the dev server, viewing a
  fixture file, and checking the browser Network tab shows no requests beyond the app's own JS/CSS
  bundle); an intentionally broken PlantUML fence shows a clean one-line error message and the
  rest of the page still renders.
- **Ambiguous:** none identified beyond what the direction doc already accepts (bundle size growth,
  partial diagram-type fidelity for edge-case diagram types) — no new ambiguous zones surfaced
  during grilling.
- **Bad:** any network request observed during PlantUML hydration; any fence alias silently
  falling back to a chroma-highlighted code block instead of a diagram; an invalid diagram
  crashing the page, blanking the pane, or blocking the rest of the document from rendering;
  PlantUML's own noisy native error card (with the version-nag banner) leaking into the rendered
  page instead of the plain-text error.
