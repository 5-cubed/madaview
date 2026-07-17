# Intent: Sidebar resize (drag-to-width)

**Date:** 2026-07-17
**Status:** Agreed via `/grilling` — not yet spec'd (`.context/req`) or built.

## Context

`web/src/components/Sidebar.tsx` currently has two fixed widths: `w-64`
(256px) open, `w-10` (40px) folded via the toggle button
(`.context/req/sidebar-fold-unfold.md`, `.context/archi/sidebar-fold-unfold.md`).
There is no way to choose a width in between. Separately,
`web/src/components/PaneDivider.tsx` already implements a draggable divider
for the split-view feature — ratio-based (0–1), clamped to 0.15–0.85 of a
flex container, mousedown/mousemove/mouseup, `cursor-col-resize`, no
persistence, no keyboard/touch support. This intent adds a second, distinct
drag affordance: a pixel-width resize handle on the sidebar's trailing edge.

## Decisions

- **Persist to `localStorage['madaview:sidebar-width']`.** Follows the
  fold-state precedent (`madaview:sidebar-folded`) rather than PaneDivider's
  session-only ratio — the fold-unfold spec already established "restore at
  unfold" as this sidebar's philosophy, and width should behave the same way.
- **Bounds: 180px–600px, default 256px.** Fixed pixel range (not a ratio of
  window width like PaneDivider) — the sidebar is a fixed-purpose tree panel,
  not a flexible content pane, so it needs bounds that stay usable regardless
  of window size. 256px matches today's existing default.
- **Standalone drag-handle logic in `Sidebar.tsx`, not a generalized
  `PaneDivider`.** Mirrors PaneDivider's mousedown/mousemove/mouseup +
  `cursor-col-resize` pattern but reports pixel width instead of a ratio.
  Keeps fold state's existing "no lifting out of `Sidebar.tsx`" convention
  intact and avoids reshaping `PaneDivider`'s ratio API (and `Workspace.tsx`'s
  usage of it) to serve a second, different caller.
- **Handle hidden/inactive while folded.** Matches the folded rail's existing
  design intent (icon-only, toggle-button-only — per the fold-unfold spec's
  rejection of a "draggable rail"). No auto-unfold-on-drag.
- **No snap-to-fold at the minimum width.** Resize and fold stay two
  independent, explicit actions — consistent with the fold spec's philosophy
  of deliberate state changes (toggle-button-only, no implicit triggers).
- **Suppress the `transition-[width] duration-200` during an active drag.**
  Left on, it makes width visually lag the cursor by 200ms. Track an
  `isDragging` flag; drop the transition class while dragging, restore it
  after, so fold/unfold keeps its animation and drag tracks 1:1.
- **Open-state width applied via inline `style={{ width }}`,** not a Tailwind
  class — drag values are arbitrary pixels, not steppable utility classes.
  Matches `Workspace.tsx`'s existing `style={{ flexBasis: ... }}` pattern for
  continuously-variable, user-controlled sizing.
- **`Sidebar()` returns a Fragment: `<nav>` + handle `<div>` as flex
  siblings**, rendered only when unfolded. `Sidebar` and the routed content
  are already flex siblings in `App.tsx`, so this lets the handle sit
  visually at the sidebar/content border without any `App.tsx` changes —
  mirrors how `PaneDivider` sits between `Workspace`'s two `Pane` divs.
- **Invalid stored width (missing/non-numeric) → default 256px; numeric but
  out-of-bounds → clamp into [180, 600].** Distinguishes "garbage" from "a
  still-usable value that's merely outside a since-changed range," rather
  than collapsing both cases to the same fallback.
- **Write to `localStorage` only on drag-end (mouseup), not per-mousemove.**
  React state updates live during the drag for rendering; the persisted
  write is a single commit at release, avoiding hundreds of blocking
  `localStorage.setItem` calls per drag gesture.
- **Mouse-only: no keyboard resize, no touch support, no double-click
  reset.** Matches `PaneDivider`'s exact current scope — no precedent
  elsewhere in this app for keyboard-driven resize or touch dragging, and
  double-click reset wasn't requested.

## Out of Scope

- Keyboard-driven resize (arrow keys on the handle).
- Touch/pointer-event support.
- Double-click-to-reset-to-default.
- Snap-to-fold when dragged to the minimum width.
- Cross-browser-tab live sync of width (same as fold state — no `storage`
  event listener).
- Generalizing `PaneDivider` into a shared ratio-or-pixel drag primitive.

## Next Steps

Write `.context/req/sidebar-resize.md` and
`.context/archi/sidebar-resize.md` following the `sidebar-fold-unfold`
pair's format, then implement per this project's plan/build pipeline.
