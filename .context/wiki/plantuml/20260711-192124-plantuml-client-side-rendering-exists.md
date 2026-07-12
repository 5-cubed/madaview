# PlantUML has an official client-side-only renderer

The initial ADR (`.context/adr/20260711-173722-madaview-initial-setup.md`) deferred PlantUML out of v1 scope, reasoning that it "typically needs a JVM or a server round-trip," conflicting with madaview's "zero runtime install" and "content never leaves the browser" constraints.

That reasoning was based on an incomplete picture. The real, current npm package is **`@plantuml/core`** (not `plantuml.js`/`plantuml-core` as first assumed — those names don't exist on the registry) — an official PlantUML build that runs **entirely client-side in the browser**: TeaVM compiles the Java PlantUML engine to JS, and Viz.js/WASM handles the Graphviz diagram-layout step. No JVM, no server round-trip, no network calls (confirmed empirically: a real diagram render fires zero requests beyond the app's own bundled JS).

This means PlantUML can follow the exact same client-side hydration pattern already used for Mermaid and KaTeX — it does not actually conflict with either constraint. The deferral has been reversed and the feature is implemented; see `.context/adr/20260711-193529-plantuml-rendering.md` for the committed decision, [[20260711-232133-plantuml-core-api-shape]] for the package's actual API, and [[20260711-232150-plantuml-error-card-detection]] for how invalid-syntax errors are surfaced.
