# Sidebar file tree: markdown-only filter

## Goal
The sidebar/file tree (`GET /api/tree` → `Sidebar.tsx`) shows only markdown files and the folders needed to reach them — never non-markdown files, and never dead-end folders that contain no markdown anywhere below them.

## Failure Criteria
- Non-markdown filenames (e.g. `.env`, `package.json`, images) still appear in the sidebar or are still present anywhere in the `/api/tree` HTTP response payload — filtering that only hides items client-side but still ships the full listing over the network does not satisfy this.
- A folder with zero markdown files anywhere in its subtree still shows up in the tree, letting a user expand into a dead end with nothing to click.
- A markdown file with a valid extension (`.md`, `.markdown`, `.mdx`) fails to show because of case mismatch (e.g. `README.MD`).
- Existing tests (`tree_test.go`, `list_test.go`) are left asserting the old unfiltered behavior instead of being updated to match.

## Ambiguous Zone
- The recursive "does this subtree contain markdown" check is real added I/O cost that scales with subtree size/depth, rerun on every lazy per-level expand (no caching layer). Accepted as a deliberate tradeoff for v1 — correctness of "no dead-end folders" was chosen over this cost. Revisit only if it becomes a measured performance problem, not preemptively.
- No other file-listing surface exists in the app today (confirmed by search: no search feature, no breadcrumb component, no recent-files list) — `/api/tree` is the only place this filter needs to apply.

## Direction
Filter server-side: `internal/rootfs/list.go` `List()` (or a wrapper it calls) excludes any file whose extension isn't `.md`, `.markdown`, or `.mdx` (case-insensitive), and excludes any directory whose subtree contains no such file anywhere below it (recursive check, not just one level). `/api/tree` therefore never transmits non-markdown filenames or dead-end folders in its response at all — this is not a client-side rendering filter. `Sidebar.tsx` needs no filtering logic of its own since the API already returns exactly what should be shown.

## Constraints
- Extensions matched: `.md`, `.markdown`, `.mdx` — case-insensitive.
- Filtering happens in the Go backend (`internal/rootfs` / `internal/server`), not in `web/src/components/Sidebar.tsx`.
- The recursive subtree check must respect the same path-safety/symlink boundaries `List()` already enforces (root-folder confinement, symlink target resolution) — no new traversal exposure introduced.
- `TestTree_ListsSingleLevel` in `internal/server/tree_test.go` currently asserts an unfiltered 1-md-file + 1-dir listing returns 2 entries; this test (and any `internal/rootfs/list_test.go` coverage) must be updated to reflect markdown-only + subtree-aware filtering.

## Out of Scope
- Client-side filtering in `Sidebar.tsx` (rejected — non-md filenames must never leave the server, not just go unrendered).
- Caching/memoizing the recursive subtree check (deferred — only revisit if performance becomes a measured problem).
- Any change to other views — confirmed none currently list directory contents besides the sidebar.
