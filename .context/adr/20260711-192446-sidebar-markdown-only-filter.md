# ADR: Sidebar File Tree Markdown-Only Filter

**Date:** 2026-07-11

## Context
`/api/tree` (`internal/rootfs/list.go` `List()` → `internal/server/tree.go` `handleTree` →
`web/src/components/Sidebar.tsx`) currently returns every filesystem entry at a given level,
unfiltered. This ships non-markdown filenames (`.env`, `package.json`, images) to the browser and
lets a user expand into folders that contain no markdown anywhere below them — a dead end. Full
goal and failure criteria live in
`.context/direction/20260711-191722-sidebar-markdown-only-filter.md`. Confirmed via search
(`.context/wiki/20260711-191747-single-file-listing-surface.md`) that `/api/tree` is the only
directory-listing surface in the app, so this is the only place the filter needs to apply.

## Decision
Filter entirely server-side, inside `rootfs.List()`'s existing implementation — no new package,
no signature change. `List()` remains `(reqPath string) ([]Entry, error)`; `internal/server/tree.go`
and `Sidebar.tsx` require zero changes, since the API now returns exactly what should be shown.

Recognized markdown extensions: `.md`, `.markdown`, `.mdx`, matched case-insensitively (see
`.context/wiki/20260711-191747-recognized-markdown-extensions.md`).

### Filtering rule
- A file entry is kept only if its extension matches one of the recognized markdown extensions
  (case-insensitive).
- A directory entry is kept only if its subtree contains at least one markdown file anywhere
  below it (recursive check, not just one level).

### Recursion mechanism
The recursive "does this subtree contain markdown" check recurses by calling `List()` itself
again on each subdirectory, rather than writing a second, independent filesystem walk. This
reuses `Resolve()`'s path-safety logic and the existing symlink-aware `IsDir` stat logic with zero
duplication — the recursive check can never diverge from `List()`'s safety behavior because it
*is* `List()`. The added I/O cost (re-stat per level, no caching) is an accepted v1 tradeoff per
the direction doc's Ambiguous Zone.

### Symlink cycle guard
Recursing into subdirectories (including symlinked ones, consistent with the existing policy that
a symlink physically under root is followed even if its target is outside root) can loop forever
if a symlink points back to one of its own ancestor directories — `os.Stat`/`os.ReadDir` will not
`ELOOP` on this case (only true symlink-to-symlink chains hit the OS's own depth limit). Left
unguarded, this is a real hang/crash risk on a LAN-exposed server: anyone able to place a symlink
under the served root could plant a self-referential loop.

Guard: during a single top-level dead-end check's recursive descent, resolve each directory's real
path (`filepath.EvalSymlinks`) and track visited real paths in a set scoped to that descent. If a
subdirectory's real path was already visited earlier in the same descent, treat it as contributing
no new markdown and stop descending into it, rather than recursing again.

### Observability
No new logging is added. Filtering out non-markdown files and dead-end directories is normal,
expected behavior on every request, not an error or rejection — per-entry `slog.Warn` lines would
be noise on every listing. The existing request-level access log in `tree.go` (method, path,
status, duration) remains sufficient to see that requests are happening; a debug-level filter
count can be added later if it's ever actually needed for debugging.

## Design
- `internal/rootfs/list.go`:
  - `isMarkdownName(name string) bool` — unexported helper, lowercases `filepath.Ext(name)` and
    compares against `.md`, `.markdown`, `.mdx`.
  - `subtreeHasMarkdown(reqPath string, visited map[string]struct{}) (bool, error)` — unexported
    helper. Resolves `reqPath`'s real path via `filepath.EvalSymlinks`; if already in `visited`,
    returns `false, nil` (cycle, stop); otherwise adds it to `visited`, calls `r.List(reqPath)`
    (pre-filter internal call, or filtered — either way markdown files surface `true`
    immediately), and for each directory entry recurses. Returns `true` as soon as any markdown
    file is found or any recursive call returns `true`.
  - `List()` filters its `dirEntries` loop: skip non-directory entries that fail
    `isMarkdownName`; skip directory entries where `subtreeHasMarkdown` returns `false`.
- `internal/server/tree.go`, `web/src/components/Sidebar.tsx`: unchanged.

## Action Sequence
1. Add unexported `isMarkdownName(name string) bool` helper (case-insensitive `.md`/`.markdown`/`.mdx` match) in `internal/rootfs`.
2. Add unexported recursive `subtreeHasMarkdown()` helper that recurses via `List()` itself, tracking visited real paths (via `filepath.EvalSymlinks`) to guard against symlink cycles.
3. Wire both into `List()`: keep markdown files, keep directories where `subtreeHasMarkdown` is true, drop everything else.
4. Update/extend `internal/rootfs/list_test.go`: non-md file exclusion, dead-end dir exclusion, case-insensitive match (`README.MD`), deep-nested markdown kept, symlink-cycle doesn't hang.
5. Update `internal/server/tree_test.go`: fix `TestTree_ListsSingleLevel` fixture (add `docs/nested.md` so `docs/` legitimately survives filtering), add a payload-level test asserting excluded names never appear anywhere in the raw JSON body.
6. Create `e2e/sidebar-tree-markdown-only-filter/{run,verify,data/root}` following the `path-traversal-rejected` structural pattern (API-only fetch, no browser needed).
7. Run `go test ./...` and the e2e suite; confirm green.
8. Manual live check: `make build`, run the binary against a real mixed-content folder, `curl /api/tree` and eyeball the Sidebar in a browser.
9. Write domain truths/decisions to the wiki via `/to-wiki` (recursion-via-`List()`-with-cycle-guard decision).
10. Write a changelog entry via `/to-changelog`.
11. Remove the originating `.context/TODO.md` item via `/to-todo`, if one exists (none currently does).

## Test-Loop Design
- **`run`:** `e2e/sidebar-tree-markdown-only-filter/run` resets its `result/` dir, builds a fixture
  tree under `data/root/`:
  - `guide.md` (kept), `README.MD` (mixed-case, kept)
  - `.env`, `package.json`, `image.png` at top level (excluded — non-markdown)
  - `docs/nested.md` (kept — proves recursive keep, and that `nested.md` doesn't leak to the
    parent level)
  - `empty-assets/image.png` (excluded — dead-end dir, no markdown anywhere below)
  - `deeply/nested/dir/deep.md` (kept — proves multi-level recursion reaches deep files)

  Starts the server against this root, fetches `GET /api/tree?path=` and
  `GET /api/tree?path=docs`, capturing the raw JSON response bodies (text, not just parsed) plus
  server logs. Writes `result/responses.json` (per-path status + raw body) and
  `result/metadata.json` (scenario name, root, port).
- **`verify`:** reads `responses.json`, asserts:
  - Root-level response's parsed entries are exactly `{guide.md, README.MD, docs, deeply}` — no
    `.env`, `package.json`, `image.png`, or `empty-assets`.
  - `docs`-level response is exactly `{nested.md}`.
  - The *raw* body text of every captured response never contains the substrings `.env`,
    `package.json`, `image.png`, or `empty-assets` — catches the "still present in the payload"
    failure mode, not just wrong parsed entries.
  - Reports `good` / `unexpected` with the specific mismatch and a pointer to `responses.json`.
- **Scenario:** `sidebar-tree-markdown-only-filter` → root and `docs` listings contain only
  markdown files and non-dead-end directories; excluded names never appear in either response
  body.

## Evaluation Criteria
- `go test ./...` passes, including new/updated cases in `internal/rootfs/list_test.go` and
  `internal/server/tree_test.go`.
- The new e2e scenario's `verify` reports `good`.
- Manual live check: `make build && ./madaview --root <mixed-content-folder>`, then
  `curl localhost:4800/api/tree` — the raw JSON text visually contains no non-markdown filenames
  and no dead-end folder names anywhere in the response, only markdown files and folders that
  contain markdown below them. Opening the Sidebar in a browser shows the same filtered set, with
  no folder that expands into nothing.
- A symlink loop planted under root (e.g. `dir/link -> dir`) does not hang or crash the server —
  `subtreeHasMarkdown`'s cycle guard returns `false` for the repeated path instead of recursing
  forever.
