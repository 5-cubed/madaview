# Subtree markdown check: why the cycle guard needs a shared readDir, not mutual List() recursion

`internal/rootfs/list.go`'s dead-end-directory filter (keep a directory only
if its subtree contains markdown somewhere below it) cannot be implemented
by having `List()` and the recursive check call each other through the
*public* `List()` method with a fresh visited-map created on every call.
That defeats the symlink-cycle guard: each mutual-recursion hop resets the
map, so a self-referencing symlink (e.g. `dir/link -> dir`) recurses
forever instead of being caught.

The working design:
- An unexported `readDir(reqPath)` holds the raw `Resolve` + `stat` +
  `ReadDir` + symlink-aware `IsDir` logic — single source, no duplication
  between the two call sites that need it.
- `List(reqPath)` calls `readDir` once, creates **one** fresh visited map
  per top-level call, and applies filtering (markdown-only files,
  `subtreeHasMarkdown` for directories) using that single shared map across
  all sibling directory checks in the same call.
- `subtreeHasMarkdown(reqPath, visited)` resolves the real path via
  `filepath.EvalSymlinks`, checks/adds it to the shared `visited` set, then
  calls `readDir` (not `List`) and recurses into subdirectories by calling
  *itself* directly (not `List`), threading the same map instance all the
  way down the descent.

Verified by a symlink self-loop test (`dir/link -> dir`) asserting `List`
returns within a few seconds instead of hanging.
