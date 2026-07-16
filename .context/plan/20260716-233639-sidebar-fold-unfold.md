# Plan: Sidebar fold state + persistence

**ADR:** `.context/adr/20260716-233115-sidebar-fold-unfold.md`

> Note: this app has no unit-test framework (`web/package.json` has no
> jest/vitest) — per the ADR's Test-Loop Design, verification is E2E-only
> via the `e2e/{scenario}/run`+`verify` harness. Each RED step below extends
> `e2e/sidebar-fold-unfold/run` and `verify`, run via `./test`, in place of a
> unit test file.

## Action Sequence
1. (RED) Create `e2e/sidebar-fold-unfold/data/` fixture: `readme.md`,
   `notes.md`, and a `guide/` subfolder containing `intro.md` (enough files
   to exercise a folder expand/collapse). Create
   `e2e/sidebar-fold-unfold/run`, `verify`, `test` (mirroring
   `e2e/theme-switching/`'s structure: `test` calls
   `runThenVerify(scenarioDir)` from `../lib/test-runner.mjs`). `run` uses
   `resetResultDir`, `startServer`, `nextPort`, `writeMetadata`, `writeJSON`
   from `../lib/harness.mjs` and `withPage` from `../lib/browser.mjs` to
   start the server against the fixture, load the page fresh (no
   `madaview:sidebar-folded` key set), read `nav[data-testid="sidebar"]`'s
   `data-folded` attribute and `getComputedStyle(nav).width`, and write them
   to `result/interaction.json` as `{ initial: { dataFolded, width } }`.
   `verify` reads `interaction.json` via `readResultJSON` and asserts
   `initial.dataFolded === "false"` and `initial.width === "256px"` (the
   `w-64` pixel value), calling `report('unexpected', ...)` on failure.
   Run `./e2e/sidebar-fold-unfold/test`; confirm it fails because
   `data-testid="sidebar"` / `data-folded` do not exist on `<nav>` in
   `web/src/components/Sidebar.tsx` yet.
2. (GREEN) In `web/src/components/Sidebar.tsx`, add a private
   `getStoredFolded(): boolean` function (`localStorage.getItem('madaview:sidebar-folded') === 'true'`,
   no try/catch), and change `Sidebar()` to
   `const [folded, setFolded] = useState(() => getStoredFolded())`. Add
   `data-testid="sidebar"` and `data-folded={folded}` to the `<nav>`
   element. Run `./e2e/sidebar-fold-unfold/test`; confirm step 1 now
   passes.
3. (RED) Extend `run` to, after the initial capture, click
   `[data-testid="sidebar-toggle"]` and capture a `folded` entry in
   `interaction.json`: `nav`'s `data-folded` attribute,
   `getComputedStyle(nav).width`, `getComputedStyle(nav).transitionDuration`,
   and `getComputedStyle(page.locator('[data-testid="sidebar-tree"]'))`'s
   `display`. Extend `verify` to assert `folded.dataFolded === "true"`,
   `folded.width === "40px"`, `folded.transitionDuration === "0.2s"`, and
   `folded.treeDisplay === "none"`. Run `./e2e/sidebar-fold-unfold/test`;
   confirm it fails (no `sidebar-toggle` button or `sidebar-tree` wrapper
   exist yet).
4. (GREEN) In `Sidebar.tsx`, add a private `setStoredFolded(folded: boolean)`
   function (`localStorage.setItem('madaview:sidebar-folded', String(folded))`,
   no try/catch). Add a toggle `<button data-testid="sidebar-toggle"
   aria-label={folded ? 'Expand sidebar' : 'Collapse sidebar'}
   onClick={...}>` at the top of `<nav>`, rendering
   `VscLayoutSidebarLeftOff`/`VscLayoutSidebarLeft` from `react-icons/vsc`
   (folded ? off-icon : open-icon); its click handler computes `next =
   !folded`, calls `setFolded(next)` and `setStoredFolded(next)`. Change
   `<nav>`'s className to switch between `w-64 p-3` and `w-10 p-1`, plus an
   always-on `transition-[width] duration-200 ease-in-out`. Wrap the
   existing `<TreeLevel path="" />` in `<div data-testid="sidebar-tree"
   className={folded ? 'hidden' : ''}>`. Run
   `./e2e/sidebar-fold-unfold/test`; confirm step 3 now passes.
5. (RED) Extend `run` to, before folding, click the `guide/` folder row
   (`page.getByText('guide/')`) to expand it and capture the resulting
   number of visible tree rows (`page.locator('[data-testid="sidebar-tree"] li').count()`)
   as `expandedRowCount`. Record `requests.length` at this point as
   `requestsBeforeToggle`. Then fold (click `sidebar-toggle`), then unfold
   (click `sidebar-toggle` again), and capture: `unfolded.width` (expect
   back to `256px`), `unfolded.treeDisplay` (expect not `"none"`),
   `unfolded.rowCount` (`sidebar-tree li` count, expect equal to
   `expandedRowCount` — `intro.md` still visible under `guide/`), and
   `requestsAfterUnfold` (expect `requests.length === requestsBeforeToggle`,
   i.e. no new `/api/tree` call fired for `guide/` on unfold). Extend
   `verify` to assert all four. Run `./e2e/sidebar-fold-unfold/test`. If it
   fails, the fix is: ensure the `sidebar-tree` wrapper added in step 4 is
   never conditionally unmounted (only its `className` toggles) — adjust
   `Sidebar.tsx` accordingly and re-run until green. If it already passes,
   no further code change is needed; this confirms step 4's
   className-only-toggle correctly preserves `TreeNode`'s mounted
   `expanded` state.
6. (RED) Extend `run` to, after step 5's sequence, call `page.reload()`,
   then immediately (before any interaction) capture `nav`'s `data-folded`
   attribute and `getComputedStyle(nav).width` as `afterReload`. Extend
   `verify` to assert `afterReload.dataFolded === "true"` and
   `afterReload.width === "40px"` (the fold from step 5 persisted and
   renders correctly on the very first paint, since the page was left
   folded at the end of step 5). Run `./e2e/sidebar-fold-unfold/test`. If
   it fails, the fix is in `getStoredFolded()` or the `useState` lazy
   initializer added in step 2 (it isn't running synchronously before
   first render, or the key name doesn't match what step 4's
   `setStoredFolded` writes) — correct `Sidebar.tsx` accordingly and re-run
   until green. If it already passes, no further code change is needed.
7. (RED) Extend `run` to, after step 6's reload, set
   `localStorage.setItem('madaview:sidebar-folded', 'garbage')` via
   `page.evaluate`, reload again, and capture `data-folded`/width as
   `afterGarbage`. Extend `verify` to assert `afterGarbage.dataFolded ===
   "false"` and `afterGarbage.width === "256px"` (a stored value other than
   the exact string `"true"` falls back to open). Run
   `./e2e/sidebar-fold-unfold/test`. If it fails, fix the equality check in
   `getStoredFolded()` (`Sidebar.tsx`) so only the exact string `"true"`
   resolves to folded; anything else must resolve to `false`. Re-run until
   green. If it already passes, no further code change is needed.
8. Manually verify the ADR's two-browser-tab boundary case (not
   e2e-automated, per the ADR's Test-Loop Design): open the app in two
   browser tabs, fold the sidebar in tab A, confirm tab B's sidebar stays
   in its prior state until tab B is manually reloaded.

## Closeout
- [ ] Refactor
- [ ] Test
