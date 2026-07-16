# Architecture: Sidebar fold state + persistence

**Date:** 2026-07-16

## Static View

`web/src/components/Sidebar.tsx` is a single component file containing:

- **`getStoredFolded(): boolean`** (line 7–9) — reads `localStorage` key
  `'madaview:sidebar-folded'`, returns `true` if value equals exact string
  `"true"`, otherwise `false`. No try/catch; failure mode matches
  `theme.ts`'s pattern.
- **`setStoredFolded(folded: boolean): void`** (line 11–13) — writes folded
  state to `localStorage` as string (`String(boolean)`).
- **`Sidebar()` component** (line 15–43) — exports, manages fold toggle UI.
  - `const [folded, setFolded] = useState(() => getStoredFolded())` — lazy
    initializer synchronously reads `localStorage` before first render, so
    DOM renders correct width on first paint (no flash).
  - `handleToggle()` callback — flips `folded`, calls `setStoredFolded(next)`
    to persist immediately (no effect needed).
  - `<nav>` — width/padding switch via conditional className
    (`w-64 p-3` when open, `w-10 p-1` when folded), plus unconditional
    `transition-[width] duration-200 ease-in-out` for animation.
    `data-testid="sidebar"` and `data-folded={folded}` for e2e observability.
  - `<button data-testid="sidebar-toggle">` — renders fold/unfold toggle icon
    (`VscLayoutSidebarLeft` when open, `VscLayoutSidebarLeftOff` when
    folded), icon size 16px, aria-label changes with state.
  - `<div data-testid="sidebar-tree" className={folded ? 'hidden' : ''}>` —
    wrapper around `<TreeLevel>` that applies `display:none` when folded
    (CSS-only toggle, never unmounted).
- **`TreeLevel()` and `TreeNode()` components** (lines 46–114) — unchanged
  from prior implementation; kept mounted during fold/unfold so
  `TreeNode`'s per-node `expanded` state in `useState` survives the toggle.

## Dynamic View

### Scenario: First visit, fold, reload

```
User → Load app (no localStorage key)
  ↓
getStoredFolded() reads missing key
  → returns false
  ↓
Sidebar renders: folded=false, width=256px, tree visible
  ↓
User clicks toggle button
  ↓
handleToggle() invoked
  → setFolded(true) updates React state
  → setStoredFolded(true) writes to localStorage
  ↓
Sidebar re-renders: folded=true, width animates 256px→40px over 200ms, tree hidden
  ↓
User reloads page
  ↓
getStoredFolded() reads localStorage key='madaview:sidebar-folded' value='true'
  → returns true
  ↓
Sidebar renders: folded=true, width=40px, tree hidden
  (no flash of open state because useState initializer ran before first paint)
```

### Scenario: Fold/unfold round-trip preserves tree expand-state

```
User → Sidebar open, user expands a folder in tree
  ↓
TreeNode(entry={folder}) → setExpanded(true)
  → DOM shows nested <TreeLevel> with sub-entries
  ↓
User clicks toggle to fold
  ↓
handleToggle() → setFolded(true)
  ↓
Sidebar re-renders with className={folded ? 'hidden' : ''}
  → TreeLevel/TreeNode stay mounted (never unmounted)
  → TreeNode's expanded state (useState) survives in memory
  → tree DOM hidden by CSS display:none
  ↓
User clicks toggle to unfold
  ↓
handleToggle() → setFolded(false)
  ↓
Sidebar re-renders, className removes 'hidden'
  ↓
TreeNode's expanded state (still in useState) and mounted DOM are now visible
  → folder is still expanded, no new /api/tree request fired
```

### Scenario: Corrupted localStorage value falls back to open

```
User's localStorage['madaview:sidebar-folded'] is set to 'garbage' (or deleted, or '1')
  ↓
getStoredFolded() reads value
  → value !== 'true' → returns false
  ↓
Sidebar renders: folded=false, width=256px, tree visible
  (same as first visit with no key)
```

## Observability

- **`data-testid="sidebar"` and `data-folded={folded}`** on `<nav>` — e2e test
  can read fold state via Playwright's
  `page.locator('[data-testid="sidebar"]').getAttribute('data-folded')`.
- **`data-testid="sidebar-toggle"` and `aria-label`** on toggle button — e2e
  can click toggle and verify aria-label changes (doubles as a selector).
- **`data-testid="sidebar-tree"` on wrapper div** — e2e can read tree
  visibility via `window.getComputedStyle(...).display` and row count via
  `document.querySelectorAll('[data-testid="sidebar-tree"] li').length`.
- **Browser DevTools** — `localStorage.getItem('madaview:sidebar-folded')`
  returns the persisted string directly; `getStoredFolded()` can be called
  in console to verify the parsing logic.
- **CSS transition** — computed style `transitionDuration` reads as `'0.2s'`
  (200ms).
- No new console logging needed; failure modes (localStorage disabled,
  missing DOM elements) are observable via Playwright assertions and DevTools.

## Test-Loop Design

### Scenario: `e2e/sidebar-fold-unfold/`

**Fixture:** `e2e/sidebar-fold-unfold/data/` contains `readme.md`, `notes.md`,
`guide/intro.md` (one nested folder to exercise mount preservation).

**`run`:** Starts server against fixture, drives Playwright through:
1. Load page → capture initial state (width, data-folded)
2. Click toggle → capture folded state (width, transition-duration,
   localStorage value, tree visibility)
3. Click toggle again → capture unfolded state (width, row count, request
   count to verify no re-fetch)
4. Reload page → capture state after reload (width, data-folded)
5. Corrupt localStorage and reload → capture fallback state (width,
   data-folded)

Writes `result/interaction.json` with all captures, `metadata.json`, and
`server.log`.

**`verify`:** Reads `interaction.json` and asserts:
- Initial: `width=256px`, `data-folded="false"`
- Folded: `width=40px`, `data-folded="true"`, `treeDisplay="none"`,
  `storedValue="true"`, `transitionDuration=0.2s`
- Unfolded: `width=256px`, `data-folded="false"`, row count matches pre-fold
  count (tree stayed mounted), request count unchanged (no re-fetch)
- After reload: `width=40px`, `data-folded="true"` (persistence verified)
- After garbage: `width=256px`, `data-folded="false"` (fallback works)

Reports `good` if all checks pass, `unexpected` with problem list otherwise.

## Verification Criteria

Per `.context/req/sidebar-fold-unfold.md` Acceptance Criteria:

- **Normal (toggle works):** e2e checkpoint 1–3: sidebar opens on first
  visit, folds on click, animates 200ms, tree hides but stays mounted,
  localStorage saves.
- **Normal (mount preservation):** e2e checkpoint 4: folder expanded before
  fold remains expanded after unfold, no API call on unfold.
- **Normal (persistence):** e2e checkpoint 5: sidebar renders folded on
  first paint after reload, no flash of open state.
- **Exception (fallback):** e2e checkpoint 6: corrupted localStorage defaults
  to open.
- **Boundary (first visit):** e2e checkpoint 1: no key → renders open.
- **Boundary (200ms timing):** manual test: computed transition-duration
  reads as `0.2s`.
- **Boundary (no cross-tab sync):** manual two-tab test: fold in tab A does
  not affect tab B until B reloads.
