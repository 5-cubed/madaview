# Tabs and Split View

## Goal
Replace madaview's single "one URL, one visible file" viewing model with a lightweight editor-style workspace: users can have multiple markdown files open as tabs, view up to two of those tabs side-by-side in a split, and close tabs/panes cleanly — without losing the app's core simplicity or its existing shareable-URL behavior for the common single-file case.

## Failure Criteria
- Opening a file from the sidebar ever discards another already-open tab's content/scroll state instead of adding/focusing a tab.
- Clicking a file that's already open in the focused pane creates a duplicate tab instead of focusing the existing one.
- Switching between already-open tabs re-fetches from the server or resets scroll position (breaks the "kept mounted" contract).
- A third pane can be created, or "split right" remains clickable/effective when 2 panes already exist.
- Closing the last tab in a pane leaves a visibly empty pane instead of collapsing back to single-pane full width.
- Closing a non-active tab changes the active tab, the URL, or the scroll position of the tab the user was actually looking at.
- The browser URL fails to reflect the focused pane's active tab (breaks back/forward and copy-link-to-share for the single-pane case, which is today's baseline behavior).
- Reloading the page silently tries to restore prior tabs/splits (contradicts the explicit no-persistence decision) or crashes instead of cleanly resetting to a single tab.
- Sidebar clicks open into the wrong pane when 2 panes are open (must target whichever pane was last focused, not always pane 1).

## Ambiguous Zone
- Tab reordering (drag within a strip) and cross-pane tab moving are explicitly deferred, not rejected forever — acceptable that v1 requires closing and reopening a file in the other pane if a user wants it moved.
- The pane-resize divider ratio is in-memory only, reset on reload, consistent with the broader no-persistence decision — not a separate persistence tier.
- Keyboard shortcuts for closing tabs (e.g. Cmd/Ctrl+W) and tab-strip overflow behavior (many tabs open) are implementation details left to /planning, not a blocking decision here — default to a scrollable/overflow tab strip and mouse-driven close (×) if not otherwise specified.
- Visual focus indication for "which pane is focused" (for sidebar-click targeting) is an implementation detail, not a modeled decision — any clear affordance (e.g. border highlight) satisfies this direction.

## Direction
Build a tab + split-view workspace on top of the existing React frontend:

- **Opening**: every sidebar file click opens a new tab in the currently focused pane, or focuses the existing tab if that file is already open there. No preview-tab/pinning model.
- **Tabs stay mounted**: all open tabs in a pane keep their fetched HTML and scroll position cached while inactive (hidden, not unmounted) — switching tabs is instant and never re-fetches or resets scroll.
- **Closing**: each tab has a close (×) affordance. Closing the active tab activates the immediate left neighbor (or the right neighbor if closing the first tab). Closing a non-active tab has no effect on the currently active tab, URL, or scroll state. Closing the last tab in a pane collapses that pane entirely; the remaining pane expands to full width.
- **Split view**: an explicit "split right" action on a tab creates a second pane (max 2 panes total — the action is unavailable once 2 exist) containing that same file duplicated as its own independent tab, so the user can immediately scroll it independently or replace it via the sidebar. Splits are vertical (side-by-side) only — no horizontal/stacked splits, no nesting beyond 2 panes.
- **Pane resize**: the divider between the two panes is user-draggable to adjust the width ratio; the ratio lives only in component state and resets to 50/50 on reload (no persistence).
- **Focus and sidebar targeting**: interacting with a pane (its tab strip or content) focuses it. Sidebar clicks always target the last-focused pane.
- **URL sync**: the browser URL (`/view/*`) always reflects the focused pane's active tab, updated via `history.pushState` on tab/pane-focus changes, preserving back/forward navigation and copy-link sharing for whatever the user is currently looking at in the focused pane. A freshly loaded/shared URL always opens as a single tab in a single pane — it never tries to reconstruct a prior multi-tab/split layout.
- **No persistence**: tabs, split state, active-tab-per-pane, and pane-resize ratio are all pure in-memory React state. A page reload (or the server restarting) resets to exactly one pane with one tab, driven by whatever the URL says at load time — identical to today's baseline behavior.

This directly supersedes the current model in `App.tsx`/`ContentPane.tsx` where the URL route is the sole source of truth and only one file can be visible at a time.

## Constraints
- Frontend-only change: `web/src/App.tsx`, `web/src/components/ContentPane.tsx`, `web/src/components/Sidebar.tsx`, `web/src/types.ts` are the expected touch points. No backend/API changes are anticipated — `GET /api/tree` and the file-content endpoint behind `fetchFile` are unaffected.
- Max 2 panes, vertical split only, draggable-but-unpersisted resize ratio.
- No persistence layer (no localStorage, no server-side storage) for tabs/splits/resize ratio.
- URL must continue to reflect a single "current file" (the focused pane's active tab) for deep-linking/back-forward, matching today's `/view/*` contract for the single-tab case.
- Existing markdown-only sidebar filtering ([[20260711-191722-sidebar-markdown-only-filter]]) and single-file-listing-surface behavior ([[20260711-191747-single-file-listing-surface]]) are unaffected — this feature only changes how many/which files can be *viewed* at once, not what's listed.

## Out of Scope
- Tab drag-to-reorder within a strip (deferred).
- Cross-pane drag-to-move a tab (deferred).
- More than 2 panes, or horizontal/nested splits (rejected for v1 — vertical 2-pane only).
- Persisting tabs/splits/resize ratio across reload or server restart (rejected — pure in-memory session state).
- VSCode-style single-click "preview tab" with pinning (rejected — every click is a real, persistent tab).
- Backend/API changes (rejected — this is a client-side viewing-model change only).
