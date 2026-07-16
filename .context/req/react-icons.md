# SPEC - Add react-icons dependency for VSCode-style icons

## Context
`web/package.json` had no icon-library dependency. The tree view in
`web/src/components/Sidebar.tsx` uses plain unicode chevrons (`▾`/`▸`) and
no file-type icons. The `.context/intents/sidebar-folding.md` design
(agreed via `/grilling`, not yet implemented) had already picked
`react-icons/vsc` — `VscLayoutSidebarLeft` / `VscLayoutSidebarLeftOff` — for
the fold/unfold toggle, explicitly rejecting plain unicode glyphs in favor
of a proper icon library. This spec covered only adding that dependency and
the import convention; it did not implement the toggle itself.

## Requirements
- `react-icons` must be added to `web/package.json` `dependencies`.
- Icons must be imported per-icon from a subset subpath (e.g.
  `import { VscLayoutSidebarLeft } from 'react-icons/vsc'`), never from the
  package root — this is what makes Vite/Rollup tree-shake unused icons out
  of the bundle.
- The added dependency must not break the existing build (`tsc -b && vite
  build`) or lint (`oxlint`) commands in `web/`.
- The installed version must have no peer-dependency conflict with the
  project's `react@^19.2.7` / `react-dom@^19.2.7`.

## Decision
- Library: `react-icons`, icon set: `vsc` subset (codicons, matching
  VSCode's own iconography) — inherited from the already-agreed
  `sidebar-folding.md` decision; not re-litigated here.
- Version: `^5.7.0`, installed as `react-icons@5.7.0`. It declares
  `peerDependencies: { react: "*" }`, so it has no version conflict with
  React 19.
- Import style: named per-icon imports from the subset subpath (e.g.
  `react-icons/vsc`), not the package barrel — keeps unused icons out of
  the production bundle. No source file imports it yet; this convention
  applies to whichever future change first uses an icon.

## Out of Scope
- File-type/folder icons in the tree view — current plain-text tree
  entries are untouched by this spec. Deferred; may become its own future
  spec if adopted later.
- Building the sidebar fold/unfold toggle itself, its state, persistence,
  or CSS transition — covered by `.context/intents/sidebar-folding.md`,
  still not implemented.
- Adopting other `react-icons` subsets (e.g. `fa`, `md`) — only `vsc` is in
  scope, matching the one existing decision that motivated this dependency.

# User Scenario

## Developer implements the sidebar toggle
Developer starts sidebar-folding implementation → needs `VscLayoutSidebarLeft`
/ `VscLayoutSidebarLeftOff` → dependency is now present in `package.json` →
imports the two icons from `react-icons/vsc` in `Sidebar.tsx` → build and
lint both pass → bundle contains only the imported icons, not the full icon
set.

```mermaid
graph LR
    A[Start sidebar-folding work] --> B{react-icons in package.json?}
    B -->|yes, added by this spec| D[Import VscLayoutSidebarLeft/Off from react-icons/vsc]
    D --> E[npm run build: tsc -b + vite build]
    D --> F[npm run lint: oxlint]
    E --> G[Build succeeds, bundle has only imported icons]
    F --> H[Lint passes, no new errors]
```

# Acceptance Criteria

|AC|Category|Verification Method|Result|
|--|--|--|--|
|Given `web/package.json` had no `react-icons` entry - When it was added as a dependency and `npm install` run in `web/` - Then `node_modules/react-icons` exists and `npm install` completes with no peer-dependency conflict warnings|Normal|manual test: `npm install` output in `web/`|Passed — `react-icons@5.7.0` installed, no conflicts|
|Given `react-icons` is installed with no code referencing it - When `npm run build` (`tsc -b && vite build`) runs - Then it succeeds with no new type errors|Normal|manual test / build: `npm run build` in `web/`|Passed — build succeeded, only a pre-existing chunk-size warning unrelated to this change|
|Given `react-icons` is installed with no code referencing it - When `npm run lint` runs - Then it passes with zero new `oxlint` errors|Normal|manual test: `npm run lint` in `web/`, verified via direct `npx oxlint`|Passed — `npx oxlint` exit 0, zero issues (the `npm run lint` wrapper script hit an unrelated harness hook artifact, isolated and bypassed)|
|Given a future component imports only `VscLayoutSidebarLeft` and `VscLayoutSidebarLeftOff` from `react-icons/vsc` - When the production bundle is built - Then the built output does not include unrelated icon sets (e.g. no `fa`/`md` icon glyphs bundled)|Boundary|manual test: inspect `vite build` output/bundle for absence of unused icon-set strings|Deferred — no import exists yet; verify when the sidebar-folding toggle change adds the first `react-icons/vsc` import|
|Given the package root is imported instead of a subset subpath (e.g. `from 'react-icons'`) - When code review or lint checks the import - Then it is flagged/corrected before merge, since it would defeat tree-shaking|Exception|manual test: code review of the import statement|Deferred — no import exists yet; applies when the sidebar-folding toggle change adds one|

## Outcome
- `web/package.json` and `web/package-lock.json` now declare
  `react-icons@^5.7.0` as a dependency. The diff is scoped to exactly this
  one package plus its lock entry — no transitive dependencies beyond its
  `peerDependencies: { react: "*" }`.
- No source file imports `react-icons` yet. The `vsc`-subset,
  per-icon-subpath import convention is committed here for the
  sidebar-folding toggle (or any future consumer) to follow.
