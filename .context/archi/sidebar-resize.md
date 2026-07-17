# Architecture: Sidebar resize (drag-to-width)

**Date:** 2026-07-17

## Static View

`web/src/components/Sidebar.tsx` is a single component file containing:

- **`clampWidth(width: number): number`** — `Math.min(600, Math.max(180, width))`.
  Single source of truth for the `[180, 600]` bound, used both by
  `getStoredWidth` and live during drag.
- **`getStoredWidth(): number`** — reads `localStorage` key
  `'madaview:sidebar-width'`. Missing key (`null`) or a non-numeric string
  returns the default, `256`; a valid number returns `clampWidth(number)`, so
  an out-of-range stored value clamps into range rather than falling back to
  the default. `raw === null` is checked explicitly before `Number(raw)` —
  `Number(null)` is `0`, not `NaN`, so without that check a missing key would
  incorrectly clamp to `180` instead of defaulting to `256`.
- **`setStoredWidth(width: number): void`** — writes width to `localStorage`
  as a string. Called exactly once per drag, from `mouseup`.
- **`Sidebar()` component** — exports, manages both fold toggle and resize.
  - `const [width, setWidth] = useState(() => getStoredWidth())` — lazy
    initializer, same synchronous-before-first-paint pattern as `folded`.
  - `const [isDragging, setIsDragging] = useState(false)` and
    `const navRef = useRef<HTMLElement>(null)`.
  - `handleResizeStart()` — sets `isDragging(true)` on the handle's
    `mousedown`.
  - A `useEffect` gated on `isDragging` (deps: `[isDragging, width]`)
    attaches `mousemove`/`mouseup` listeners to `window` only while a drag
    is active:
    - `mousemove`: `clampWidth(e.clientX - navRef.current.getBoundingClientRect().left)`,
      then `setWidth(...)`. Absolute, rect-relative — recomputed fresh every
      move, matching `PaneDivider`'s calculation shape, so no drag-start
      delta state is needed.
    - `mouseup`: `setStoredWidth(width)` then `setIsDragging(false)`. Because
      `width` is a `useEffect` dependency, the listeners are re-attached on
      every `mousemove`-triggered re-render, so `mouseup`'s closure always
      reads the latest `width` — needed so `setIsDragging(false)` triggers
      the re-render that restores the transition class immediately, with no
      other interaction required.
  - `<nav ref={navRef}>` — `style={folded ? undefined : { width }}` (inline,
    not a Tailwind class, since drag values are arbitrary pixels);
    `transition-[width] duration-200 ease-in-out` is present only when
    `!isDragging`, so drag tracks the cursor 1:1 while fold/unfold keeps its
    animation.
  - Return changed from a single `<nav>` to a Fragment: `<nav>` followed by
    a resize-handle `<div data-testid="sidebar-resize-handle" role="separator"
    aria-orientation="vertical">`, rendered only when `!folded`.
    `className="w-1 shrink-0 cursor-col-resize bg-[var(--border)]
    hover:bg-[var(--accent)]"` — byte-for-byte `PaneDivider`'s classes.
    `onMouseDown={handleResizeStart}`.

## Dynamic View

### Scenario: Drag to resize, reload

```
User → Sidebar open at 256px (or a previously-persisted width)
  ↓
User presses down on the resize handle
  ↓
handleResizeStart() → setIsDragging(true)
  → transition-[width] class drops from <nav>
  ↓
User drags right
  ↓
mousemove fires repeatedly:
  raw = e.clientX - navRef.current.getBoundingClientRect().left
  clamped = clampWidth(raw)
  setWidth(clamped)
  → <nav> style width updates 1:1 with the cursor, no transition lag
  ↓
User drags past 600px
  ↓
clampWidth(raw) caps at 600 → width never visibly exceeds 600px
  ↓
User releases the mouse at width=400
  ↓
mouseup fires:
  setStoredWidth(400) → localStorage['madaview:sidebar-width'] = "400"
  setIsDragging(false) → transition-[width] class restored
  ↓
User reloads the page
  ↓
getStoredWidth() reads "400" → Number("400")=400, not NaN → clampWidth(400)=400
  ↓
Sidebar renders at 400px on first paint (no flash)
```

### Scenario: Invalid or out-of-range stored value

```
localStorage['madaview:sidebar-width'] is missing, or "garbage"
  ↓
getStoredWidth(): raw === null OR Number(raw) is NaN
  → returns 256 (default)
  ↓
Sidebar renders at 256px

---

localStorage['madaview:sidebar-width'] is "900" (valid number, out of range)
  ↓
getStoredWidth(): Number("900")=900, not NaN → clampWidth(900) → 600
  ↓
Sidebar renders at 600px, not the default
```

### Scenario: Fold hides the handle, unfold restores width

```
Sidebar unfolded at a custom width, e.g. 600px
  ↓
User clicks fold toggle → handleToggle() → setFolded(true)
  ↓
Sidebar re-renders: {!folded && <handle>} no longer renders the handle
  → width state is untouched (Sidebar itself stays mounted; only the
    handle's own conditional render changes)
  ↓
User clicks toggle again → setFolded(false)
  ↓
Sidebar re-renders: handle reappears, <nav> style width={width} still 600
  → sidebar returns to its last custom width
```

## Observability

- `<nav data-testid="sidebar" data-folded={folded}>` (unchanged) —
  `getComputedStyle(nav).width` is the reused observation point for current
  width; no separate `data-width` attribute needed.
- `<div data-testid="sidebar-resize-handle">` — mousedown target for e2e's
  `page.mouse.move`/`down`/`up` simulation; `getComputedStyle(handle).cursor`
  confirms `col-resize`.
- `getComputedStyle(nav).transitionDuration` — reused checkpoint (same one
  fold-unfold reads) to confirm the class is absent mid-drag (`"0s"`) and
  present after `mouseup` (`"0.2s"`).
- Browser DevTools: `localStorage.getItem('madaview:sidebar-width')` shows
  the persisted value directly.

## Test-Loop Design

### Scenario: `e2e/sidebar-resize/`

**Fixture:** `e2e/sidebar-resize/data/readme.md` — a single file; width
behavior doesn't depend on tree contents.

**`run`:** Starts server against the fixture, drives Playwright through:
1. Load, then `localStorage.setItem('madaview:sidebar-width', '400')` +
   reload → capture `initial` (width, handle existence, handle cursor).
   Deliberately not `page.addInitScript` for this preset — that persists
   across every later `page.reload()` in the script and would keep
   overwriting the `'garbage'`/`'900'` values steps 5–6 set right before
   their own reloads.
2. Instrument `localStorage.setItem` to count writes to the
   `madaview:sidebar-width` key.
3. Single continuous drag via `page.mouse.move`/`down`/`up`: to a point
   corresponding to 400px (capture `midDrag`: width, transition-duration),
   then past 600px (capture `clampedAtMax`), then past 180px (capture
   `clampedAtMin`), then back to 400px and release (capture `afterDrag`:
   width, transition-duration, stored value, write count).
4. Reload → capture `afterReload`.
5. Set `'garbage'`, reload → capture `afterInvalid`.
6. Set `'900'`, reload → capture `afterOutOfRange`.
7. Click fold toggle → capture `folded` (handle existence). Click again →
   capture `unfolded` (width, handle existence).

Writes `result/interaction.json`, `metadata.json`, `server.log`.

**`verify`:** Reads `interaction.json` and asserts:
- Initial: `width≈400px`, handle exists, `cursor=col-resize`.
- Mid-drag: `width≈400px`, `transitionDuration="0s"` (suppressed).
- Clamped: `width≈600px` at max, `width≈180px` at min.
- After drag: `width≈400px`, `transitionDuration="0.2s"` (restored),
  `storedWidth="400"`, `writeCount===1`.
- After reload: `width≈400px`.
- After invalid: `width≈256px` (default).
- After out-of-range: `width≈600px` (clamped, not default).
- Folded: handle absent. Unfolded: `width≈600px` (restored), handle present.

Reports `good` if all checks pass, `unexpected` with problem list otherwise.

## Verification Criteria

Per `.context/req/sidebar-resize.md` Acceptance Criteria:

- **Boundary (first visit):** e2e checkpoint 1 (with no stored key path) →
  renders at the default, 256px.
- **Normal (drag tracks cursor):** e2e checkpoint 3 (`midDrag`) → width
  tracks the cursor 1:1, no transition lag.
- **Boundary (live clamp):** e2e checkpoint 3 (`clampedAtMax`/`clampedAtMin`)
  → width never visibly exceeds `[180, 600]` mid-drag.
- **Normal (persist once, restore transition):** e2e checkpoint 3
  (`afterDrag`) → exactly one `localStorage` write, transition class back.
- **Normal (reload persistence):** e2e checkpoint 4 → renders at the
  persisted width on first paint.
- **Exception (invalid fallback):** e2e checkpoint 5 → renders at the
  default, 256px.
- **Exception (out-of-range clamp):** e2e checkpoint 6 → renders clamped
  into range, not the default.
- **Boundary (fold hides handle, unfold restores width):** e2e checkpoint 7.
- **Normal (cursor affordance):** e2e checkpoint 1 → `cursor: col-resize`.
