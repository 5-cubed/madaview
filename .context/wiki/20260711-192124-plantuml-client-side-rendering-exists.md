# PlantUML has an official client-side-only renderer

The initial ADR (`.context/adr/20260711-173722-madaview-initial-setup.md`) deferred PlantUML out of v1 scope, reasoning that it "typically needs a JVM or a server round-trip," conflicting with madaview's "zero runtime install" and "content never leaves the browser" constraints.

That reasoning was based on an incomplete picture. `plantuml.js` / `plantuml-core` (github.com/plantuml/plantuml.js, github.com/plantuml/plantuml-core) is an official PlantUML build that runs **entirely client-side in the browser**: TeaVM compiles the Java PlantUML engine to JS, and Viz.js/WASM handles the Graphviz diagram-layout step. No JVM, no server round-trip, no network calls.

This means PlantUML can follow the exact same client-side hydration pattern already used for Mermaid and KaTeX — it does not actually conflict with either constraint. The deferral has been reversed; see `.context/direction/20260711-192022-plantuml-rendering.md` for the committed direction to add it.
