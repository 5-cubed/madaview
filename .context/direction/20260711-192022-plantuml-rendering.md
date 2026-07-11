# PlantUML Diagram Rendering

## Goal
Add PlantUML diagram rendering to madaview, reversing the initial ADR's deferral of PlantUML. Success looks like: a markdown file containing a ` ```plantuml `, ` ```puml `, or ` ```uml ` fenced code block renders as an actual diagram in the browser, using the exact same architectural pattern already proven for Mermaid and KaTeX — client-side hydration in-browser, zero network calls, no server-side rendering, no JVM/runtime dependency. This preserves every hard constraint from the original ADR (zero runtime install, zero build step for end users, markdown content never leaves the browser) while closing the PlantUML gap.

## Failure Criteria
- Any network request fires during PlantUML hydration (violates the "content never leaves the browser" privacy constraint — same bar Mermaid/KaTeX are already held to).
- The binary requires installing a JVM, Java, or any other runtime the end user has to set up themselves.
- A malformed/invalid PlantUML diagram crashes the page, blanks the content pane, or takes down rendering of the rest of the document.
- `plantuml`, `puml`, or `uml` fences are inconsistently recognized (e.g. only one alias works while docs commonly use another).
- The new e2e scenario (`render-plantuml`) is skipped or weakened to "renders something" without actually asserting an `<svg>` (or equivalent real diagram output) appears and zero external requests were made.

## Ambiguous Zone
- Bundle size growth from including the TeaVM-compiled PlantUML engine + WASM Graphviz layout engine is accepted as a known, non-blocking tradeoff — this is a deliberate cost of the chosen approach, not a failure, as long as the binary remains a single self-contained download.
- Diagram types that plantuml.js/plantuml-core doesn't fully support (if any edge-case diagram type has partial fidelity) are acceptable v1 gaps as long as the common diagram types (sequence, class, use-case, activity, state, component) render correctly and unsupported cases fail visibly via the inline error state rather than silently or by crashing.

## Direction
Reverse the "PlantUML explicitly deferred" decision from `.context/adr/20260711-173722-madaview-initial-setup.md`. Bundle `plantuml.js`/`plantuml-core` (TeaVM-compiled PlantUML engine + Viz.js/WASM Graphviz for diagram layout) into the frontend build, following the exact pattern already established for Mermaid and KaTeX:
- `internal/markdown`'s goldmark pipeline recognizes ` ```plantuml `, ` ```puml `, and ` ```uml ` fences (all three treated as aliases of the same PlantUML renderer) and emits a placeholder `<div>` containing the raw diagram source, same as it already does for `mermaid` fences.
- The frontend hydration pass (already scanning for Mermaid/KaTeX placeholders post-render) is extended to also find and hydrate PlantUML placeholders in-browser, entirely client-side, with zero network calls.
- On invalid/unparseable PlantUML syntax, the hydration logic catches the render error and shows a readable inline error message in place of the diagram (matching the Mermaid failure-visibility pattern) — the rest of the document continues to render normally.
- Bundle size growth is accepted; no lazy-loading/code-splitting is required for this pass.

## Constraints
- Rendering approach is locked to client-side-only (plantuml.js/plantuml-core via TeaVM + WASM Graphviz) — no server-side Go rendering, no shelling out to a JVM/`plantuml.jar`, no external rendering service (plantuml.com, Kroki, or otherwise).
- All existing ADR constraints remain in force for this feature too: zero runtime install, zero build step for end users, markdown content never leaves the browser during diagram/math rendering.
- Fence aliases: `plantuml`, `puml`, and `uml` must all be recognized as PlantUML.
- Error handling: invalid diagrams render an inline error state, not a raw code-block fallback and not a crash.

## Out of Scope
- Server-side PlantUML rendering (Go shelling out to `plantuml.jar`, requiring a bundled JRE) — rejected, breaks "zero runtime install."
- External rendering service calls (plantuml.com, self-hosted Kroki, etc.) — rejected, breaks "content never leaves the browser."
- Lazy-loading/code-splitting the PlantUML WASM engine to reduce base bundle size — not pursued this pass; bundle size growth is accepted as-is.
- Silent fallback to raw code block on render error — rejected in favor of a visible inline error state.
