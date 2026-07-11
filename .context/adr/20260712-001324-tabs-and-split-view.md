# ADR: Tabs and Split View

**Date:** 2026-07-12

## Context
Madaview's current viewing model (`App.tsx` + `ContentPane.tsx`) has the URL route as the
sole source of truth: one file visible at a time, navigating via `<Link>` fully replaces
whatever was on screen. Direction
`.context/direction/20260712-000044-tabs-and-split-view.md` supersedes this with an
editor-style workspace: multiple markdown files open as tabs, up to two panes side-by-side,
tabs kept mounted (never re-fetch, never lose scroll), and the URL still tracking whatever
the user is currently focused on for deep-linking/back-forward. Frontend-only; no backend
or API changes.

## Decision
Introduce a single new state-owning layer — a "workspace" — sitting between the existing
routing/UI layer and the existing fetch layer (`api.ts`). It owns all tab/pane/focus
structure behind a small interface (`openFile`, `focusTab`, `focusPane`, `closeTab`,
`splitRight`) and is the only thing that talks to `window.history`. Everything else
(`Sidebar`, the per-tab content renderer, the new tab-strip/pane/divider components) reads
from or calls into this layer and stays ignorant of the invariants it enforces (max 2
panes, dedup-on-open, neighbor-activation-on-close, pane-collapse, URL sync).

Key architectural bet: "tabs stay mounted, scroll never resets" is achieved for free by
toggling `display:none` on a tab's content root instead of conditionally
mounting/unmounting it — browsers natively preserve `scrollTop` on a `display:none` element
as long as it stays in the DOM. This avoids inventing a manual scroll-position save/restore
channel entirely.

Resolved via grilling (all diverge from or sharpen the direction doc in places it left
ambiguous):
- **Workspace state survives navigating to `/settings` and back.** `WorkspaceProvider`
  wraps the whole app shell above `<Routes>`, not just the `/view` routes, so `Workspace`
  never unmounts on a Settings round-trip. Only an actual page reload resets it — this is
  the literal reading of the direction's "no persistence... reload resets" language, and
  avoids a one-click nav mistake silently wiping the user's session.
- **Split-right is a per-tab icon**, not a single pane-level toolbar button acting only on
  the active tab. Matches the direction's literal wording ("an explicit split right action
  on a tab").
- **Tab labels are the path basename only** — available instantly on open, zero coupling to
  fetch state. Keeps the tab-strip/pane layer (structure) and `ContentPane` (content)
  subdomains fully orthogonal: neither needs to know about the other's state.
- **Browser Back/Forward (`popstate`) resets the workspace to a fresh single-pane/single-tab
  view matching the landed-on URL** — i.e. every `popstate` is treated exactly like a fresh
  page load. No attempt to reconstruct a historical multi-tab/split layout, since none was
  ever persisted.
- **`history.pushState` fires on every focused-pane active-tab change, literally, with no
  dedup/replaceState logic for rapid same-session tab flips.** Accepted growth of the
  history stack for heavy tab-switchers as a minor, explicitly-chosen consequence of
  following the direction's wording exactly.
- **Sidebar clicks made while on `/settings` navigate back to `/view` in addition to
  updating workspace state** (an edge case that only exists because `WorkspaceProvider` now
  lives above the route switch) — matches today's baseline where every sidebar click was a
  `<Link>` that always jumped to the viewer.

## Design

### State layer — `web/src/useWorkspace.ts` (new)
One file (~120-150 lines), the deep module for this feature: types, a pure reducer, a
`WorkspaceProvider` component, and a `useWorkspace()` hook. Not split into separate
`reducer.ts`/`hook.ts` files — the project is small enough (~dozen source files, each
~50-150 lines) that one cohesive file matches its scale.

```ts
interface Tab { id: string; path: string }
interface Pane { id: string; tabs: Tab[]; activeTabId: string }
interface WorkspaceState { panes: Pane[]; focusedPaneId: string | null }
```

Reducer actions:
- `OPEN_FILE { path }` — operates on the focused pane. If a tab with that path already
  exists there, just activates it (no dedup across panes — split intentionally creates
  cross-pane duplicates). Otherwise appends a new tab and activates it.
- `FOCUS_TAB { paneId, tabId }` — sets that pane's `activeTabId`, sets `focusedPaneId`.
- `FOCUS_PANE { paneId }` — sets `focusedPaneId` only (e.g. click inside a pane's content
  area without changing its active tab).
- `CLOSE_TAB { paneId, tabId }` — removes the tab. If it was active, activates the left
  neighbor (or right, if it was first). If the pane is now empty, the pane itself is removed
  from `panes` and, if it was focused, `focusedPaneId` moves to the remaining pane (or
  `null` if none remain).
- `SPLIT_RIGHT { fromPaneId, tabId }` — no-op if `panes.length` is already 2. Otherwise
  creates a new pane with one new tab (same `path`, new `id`), appends it, and focuses it
  (so "replace it via the sidebar" in the direction targets the new pane immediately).
- `RESET_FROM_URL { path }` — used only by the `popstate` handler; rebuilds to a single
  pane/single tab (or an empty workspace if `path` is `''`), discarding everything else.

Dev-only invariant assertions run after every transition (see Observability).

`WorkspaceProvider`:
- Seeds initial state from `location.pathname` via a lazy `useState` initializer (runs
  once, no extra effect/ref-guard needed).
- Registers a `popstate` listener dispatching `RESET_FROM_URL` with the new
  `window.location.pathname`.
- Runs an effect on `state` that, when the focused pane's active tab's path changes,
  calls `window.history.pushState(null, '', `/view/${path}`)` directly — bypassing
  react-router's `navigate` so `Workspace` doesn't remount on every tab switch. When there
  are no panes at all, pushes `/`.
- Exposes `openFile(path)`: dispatches `OPEN_FILE`, and if `location.pathname` doesn't
  already start with `/view/`, also calls react-router's `navigate('/view/' + path)` so a
  sidebar click made from `/settings` returns the user to the viewer.

### Content layer — `ContentPane.tsx` (refactored)
Changes from route-driven (`useParams`) to prop-driven: `{ path: string; visible: boolean }`.
Its internal fetch-on-path-change and hydrate-on-file-change effects are otherwise
unchanged from today. One instance is mounted per open tab (across both panes) and kept
alive for the tab's lifetime; `visible` only toggles a `display` style on its own root —
it is never conditionally unmounted while its tab exists. This is what makes "instant
switch, no re-fetch, no scroll reset" fall out of the existing fetch-cache logic for free,
with zero new state needed for scroll position.

### UI layer — new components under `web/src/components/`
- `TabStrip.tsx` — renders one pane's tabs: label (path basename), click-to-focus,
  close (×), and a split-right icon per tab. Only reads `{ id, path }` per tab — never
  touches `FileContent`. Carries `data-pane-id` / `data-tab-id` / `data-active` attributes
  for e2e observability.
- `Pane.tsx` — composes `TabStrip` + all of that pane's `ContentPane` instances
  (visibility-toggled, never conditionally unmounted). Applies a focus-highlight border
  when it's `focusedPaneId`; carries `data-focused`. Any click inside a pane calls
  `focusPane`/`focusTab`.
- `PaneDivider.tsx` — draggable divider between two panes. Ratio lives in local
  `useState` inside `Workspace.tsx` (not in the reducer — it's pure presentational session
  state, not a structural invariant), defaults to 0.5, resets naturally on remount.
- `Workspace.tsx` — reads `useWorkspace()`, renders 1-2 `Pane`s (+ `PaneDivider` when 2
  exist), replaces the two `<ContentPane/>` route elements in `App.tsx`.

### Wiring — `App.tsx`, `Sidebar.tsx`
- `App.tsx`: `WorkspaceProvider` now wraps the entire app shell (header, `Sidebar`,
  `Routes`) — not just the `/view` routes — so it survives a `/settings` round-trip. Route
  elements for `/view/*` and `/` become `<Workspace/>`.
- `Sidebar.tsx`: file entries stop rendering `<Link to={...}>`; they call
  `useWorkspace().openFile(entry.path)` on click instead. This is the only way to express
  "always targets the last-focused pane" and "dedup within that pane" — a plain route link
  can't carry pane-targeting information.

## Action Sequence
1. Add `Tab` / `Pane` / `WorkspaceState` types in `web/src/useWorkspace.ts`.
2. Implement the pure reducer (`OPEN_FILE`, `FOCUS_TAB`, `FOCUS_PANE`, `CLOSE_TAB`,
   `SPLIT_RIGHT`, `RESET_FROM_URL`) with dev-only invariant assertions, in the same file.
3. Implement `WorkspaceProvider` + `useWorkspace()`: initial-state-from-URL, the
   `popstate` listener, and the pushState-on-focused-active-tab-change effect.
4. Add the "navigate to `/view` on `openFile` if not already there" behavior to
   `WorkspaceProvider` (covers the Settings-round-trip sidebar-click edge case).
5. Refactor `ContentPane.tsx` to accept `{ path, visible }` props instead of `useParams`,
   toggling `display` on its own root instead of conditional unmount.
6. Create `web/src/components/TabStrip.tsx` (tab list, close, per-tab split-right icon,
   `data-*` test hooks).
7. Create `web/src/components/Pane.tsx` (composes `TabStrip` + N `ContentPane` instances,
   focus-highlight styling, `data-focused`).
8. Create `web/src/components/PaneDivider.tsx` (draggable resize, local ratio state).
9. Create `web/src/components/Workspace.tsx` (composes 1-2 `Pane`s + optional
   `PaneDivider`).
10. Update `App.tsx`: move `WorkspaceProvider` above `Routes` to wrap the whole shell; swap
    the `/view/*` and `/` route elements to `Workspace`.
11. Update `Sidebar.tsx`: replace file `<Link>` with a click handler calling
    `useWorkspace().openFile(path)`.
12. Add `e2e/tabs-and-split-view/` scenario (`data/`, `run`, `verify`, `test`) per the
    Test-Loop Design below.
13. Update `.context/wiki/` via `/to-wiki` if new truths were found or existing entries are
    stale (skip otherwise).
14. Write a changelog entry via `/to-changelog`.
15. Remove the originating `.context/TODO.md` item via `/to-todo`, if one exists.

## Test-Loop Design
No existing e2e scenario covers multi-tab/split-pane UI interaction — the existing
scenarios are single-file-render or API-only. New scenario required.

- **`run`:** Reset `e2e/tabs-and-split-view/result/`. Start the server against a
  `data/` fixture root of 3 markdown files (`a.md` long enough to scroll — ~200 lines —
  `b.md`, `c.md`, all short). Drive Playwright through, capturing a checkpoint after each
  step into `result/interactions.json` (tab count per pane, pane count, focused pane id,
  active tab id, `window.location.pathname`, request count, `scrollTop` of the active
  content root):
  1. Open `a.md` and `b.md` via sidebar clicks → 1 pane, 2 tabs.
  2. Scroll `a.md`'s tab down, switch to `b.md`, switch back to `a.md` → assert `scrollTop`
     unchanged and no new `/api/file` request fired.
  3. Re-click `a.md` in the sidebar (already open) → assert tab count still 2 (no dup).
  4. Click split-right on the active tab → assert 2 panes, pane 2 has one tab duplicating
     the same path, and split-right icons are now absent/disabled in both panes.
  5. With pane 2 focused, click `c.md` in the sidebar → assert it opens in pane 2, not
     pane 1.
  6. Close a non-active tab → assert active tab id, URL, and `scrollTop` all unchanged.
  7. Close the active tab → assert the left-neighbor tab activates.
  8. Close the last tab in a pane → assert that pane is gone and the remaining pane is
     full width (only one pane element in the DOM).
  9. Navigate to `/settings` and back → assert tabs/panes from steps 1-8 are still intact.
  10. Reload the page at whatever URL is current → assert exactly one pane with one tab
      matching the URL, no crash, no restored multi-tab state.
  11. Navigate back via `history.back()` after step 1's tab-open → assert workspace resets
      to a fresh single-pane/single-tab matching the landed-on URL (not a restored
      2-tab layout).
  Write `server.log`, `metadata.json` (scenario name, root, port) alongside
  `interactions.json`.

- **`verify`:** Reads `interactions.json` and `metadata.json`; checks each checkpoint
  against its expected value from the list above. Each checkpoint maps 1:1 to a bullet in
  the direction's Failure Criteria. Reports good/unexpected/ambiguous per checkpoint, with
  root cause traced through `server.log` for any unexpected network activity.

- **Scenarios:**
  - `tabs-and-split-view` → all 11 checkpoints pass (see above).

## Evaluation Criteria
- **Good:** `e2e/tabs-and-split-view` reports all checkpoints good, mapping 1:1 onto the
  direction's 9 Failure Criteria bullets with none triggered. Manual smoke in a real
  browser: open 3 files as tabs, split, drag the divider, close tabs down to zero in one
  pane, reload — behavior matches the Direction section exactly, no console errors, no
  network calls fire on switching back to an already-loaded tab.
- **Ambiguous:** Divider drag ratio is exercised only via manual/visual check, not asserted
  pixel-exact in the e2e scenario — acceptable per the direction's own Ambiguous Zone
  (resize ratio is explicitly not a modeled invariant beyond "draggable, resets on reload").
- **Bad:** Any single Failure Criteria bullet is observed to occur, in either the e2e
  scenario or manual smoke testing.
