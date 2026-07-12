# @plantuml/core's actual API shape

`@plantuml/core` (npm, MIT, zero deps) exports two functions from its main `plantuml.js` ES module:

- `render(lines, targetId, opts?)` — writes SVG directly into a DOM element by id.
- `renderToString(lines, onSuccess, onError)` — delivers the SVG as a string via `onSuccess(svg)`; errors go to `onError(message)`.

`lines` is a **string array** (one PlantUML source line per element), not a raw multi-line string — this differs from Mermaid's `render(id, source)`, which takes a single string. Madaview builds this array with `source.split('\n')`.

`viz-global.js` (the Graphviz/WASM layout engine) must be imported as a side effect **before** `renderToString`/`render` is called — `import '@plantuml/core/viz-global.js'` above `import { renderToString } from '@plantuml/core'`. This works fine under a real Vite production build via ES-import bundling, despite the package README recommending a classic `<script>` tag (its UMD wrapper does a `typeof exports` check that could misbehave under a bundler — verified not to in practice). The WASM binary is embedded as a base64 data URI inside `viz-global.js` (Emscripten inline-wasm mode), so it's instantiated locally and never fetched over the network.

The package ships **zero TypeScript type declarations**. Madaview hand-writes an ambient module declaration at `web/src/plantuml-core.d.ts` covering just these two entry points.

See [[20260711-232150-plantuml-error-card-detection]] for how rendering failures actually surface.
