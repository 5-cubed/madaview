# ADR: Madaview Initial Project Setup

**Date:** 2026-07-11

## Context
Madaview is a cross-platform (Windows/macOS/Linux) markdown viewer distributed as a single
self-contained Go binary with zero runtime install and zero build step for end users. The UI is
React, but distribution is a native binary that embeds the built frontend via `go:embed`, starts
an HTTP server, and is viewed in a browser. Full context, goal, and failure criteria live in
`.context/direction/20260711-165859-madaview-initial-setup.md`. The repo is currently empty —
this ADR covers the initial architecture and the first implementation pass.

## Decision

### Rendering pipeline
- Markdown is parsed to HTML **server-side in Go** using `goldmark` + the official GFM extension
  (tables, task lists, strikethrough, autolinks).
- Fenced code blocks are syntax-highlighted server-side via `goldmark-highlighting` (wrapping
  `chroma`, pure Go, no external deps) — consistent with "Go renders everything," no extra JS
  bundle needed for this.
- Mermaid diagrams and KaTeX math are **never rendered server-side and never call an external
  rendering service**. Goldmark emits placeholder markup (`<div class="mermaid">...</div>`,
  `<span class="katex-math">...</span>`) containing the raw diagram/math source; the bundled
  frontend ships `mermaid.js` and `katex.js` and hydrates these placeholders fully client-side,
  in-browser, with zero network calls. This is a hard requirement: markdown content must never
  leave the browser for diagram/math rendering (privacy/security constraint).
- PlantUML is explicitly **deferred** — not in v1 scope. It typically needs a JVM or a server
  round-trip, which conflicts with both "zero runtime install" and "no rendering server."
- Rendered HTML is passed through `bluemonday` (Go HTML sanitizer) before being sent to the
  browser, stripping `<script>` and inline event handlers. Necessary because the server is
  LAN-visible — one viewer's markdown file must not be able to run arbitrary JS in another
  viewer's browser.

### Root-folder boundary and safety
- Root resolution priority: CLI arg > last-used UI setting > current working directory default.
  CLI arg only determines the **initial** root at startup — the browser UI can still switch root
  live afterward via `POST /api/root`, regardless of how the server was launched (no lockout).
- Root is swapped at runtime in-process (atomic/mutex-guarded reference), re-validating the new
  path exists and is a directory — no restart required.
- The last-used root is persisted as JSON in the OS user-config directory
  (`os.UserConfigDir()`, e.g. `%AppData%\madaview`, `~/Library/Application Support/madaview`,
  `~/.config/madaview`) so it works regardless of where the binary is placed, including read-only
  install locations.
- Path-traversal guard: every requested path is joined with root, `filepath.Clean`ed, and
  resolved. **Symlink policy**: a symlink is allowed if it physically resides under the root
  folder, even if its resolved target points outside root — the root owner's placement of the
  symlink is treated as an explicit choice to expose that target. A symlink whose *link itself*
  sits outside root is never reachable in the first place (already excluded by the root
  boundary).

### API and frontend contract
- Go embeds a built React SPA (`internal/assets`, `go:embed` of `web/dist`) and serves a narrow
  JSON API:
  - `GET /api/tree?path=...` — lazy, single-level (non-recursive) folder listing; sidebar
    fetches deeper levels on expand. Avoids slow/huge responses on large doc trees.
  - `GET /api/file?path=...` — pre-rendered, sanitized HTML body + metadata (title, mtime) for
    a markdown file.
  - `GET /api/status` — `{version, currentRoot, rootSource, goVersion, os/arch}` for the settings
    UI and for debugging/verifying live server state.
  - `POST /api/root` — runtime root switch.
- Routing uses the Go 1.22+ standard library `http.ServeMux` (method-specific patterns) — no
  third-party router needed for this route count.
- Frontend UX: persistent sidebar file tree (lazy-loaded) + content pane, React Router for
  deep-linkable per-file URLs (e.g. `/view/docs/guide.md`).
- Server binds `0.0.0.0` on default port **4800** (overridable via `--port`).

### Package structure
```
cmd/madaview/        main.go — CLI flags (--root, --port, -v/--verbose), wiring, startup
internal/server/      HTTP routing + handlers (/api/tree, /api/file, /api/status, /api/root) + static SPA serving
internal/markdown/    goldmark pipeline: GFM + chroma highlighting + Mermaid/KaTeX placeholder emission + bluemonday sanitization
internal/rootfs/      root resolution, symlink-aware traversal guard, directory listing, atomic runtime root swap
internal/config/      OS user-config dir read/write for persisted root
internal/assets/      go:embed of web/dist
web/                  React + Vite + Tailwind + npm source, builds to web/dist
e2e/                  Playwright-based test-loop scenarios
```
Each `internal/` package is a deep module: narrow interface (e.g. `rootfs.Resolve(path)`,
`rootfs.SetRoot(path)`), all traversal/symlink/validation complexity hidden inside.

### Observability
- `log/slog` structured logging to stdout (no external logging dependency): one line per HTTP
  request (method, path, status, duration), one line per path-safety rejection (requested path,
  resolved path, reason), one line per root change (old root, new root, source: cli/ui/default).
- `-v`/`--verbose` CLI flag raises log level to Debug (symlink resolution steps, render timing).
- `GET /api/status` exposes live server state for both the UI and manual/test-loop debugging.

### Build and embed bootstrap
- `go:embed` requires `web/dist` to exist with content at compile time, but the frontend build
  only runs via `npm` in CI/dev. To avoid a fresh clone's `go build`/`go vet`/`go test` hard-
  failing before anyone has built the frontend: add a `Makefile` `build` target that runs
  `npm ci && npm run build` in `web/` then `go build ./cmd/madaview`, and check in a minimal
  placeholder `web/dist/index.html` (e.g. "run make build") so the embed directive always has
  something to embed.

### CI/CD
- Separate CI workflow: on every push/PR to main, run `go test ./...` and the Playwright e2e
  suite on a single Linux runner for fast feedback.
- Release workflow: on tag push `v*`, first re-run the full test suite (Go unit tests +
  Playwright E2E) as a gate — a tag must represent a working release. Only on pass, cross-compile
  the matrix (windows/amd64, darwin/amd64, darwin/arm64, linux/amd64) and smoke-test each binary
  by launching it and curling `/api/status` **on its native OS runner** (windows-latest,
  macos-latest for arm64 execution + amd64 build-only since no Intel Mac runner exists,
  ubuntu-latest), then publish all artifacts to GitHub Releases.
- Repo: `github.com/5-cubed/madaview`, public, MIT license.

## Design
See package structure and API contract above. Key data flow: browser requests `/api/tree` →
`internal/rootfs` lists one level under current root (validated, symlink-resolved) → JSON to
client. Browser requests `/api/file` → `internal/rootfs` validates path → file bytes → 
`internal/markdown` (goldmark + GFM + chroma + placeholder emission) → `bluemonday` sanitize →
HTML string → JSON response → client injects into content pane → client hydrates Mermaid/KaTeX
placeholders in-browser.

## Action Sequence
1. Init Go module (`github.com/5-cubed/madaview`), directory skeleton (`cmd/madaview`,
   `internal/{server,markdown,rootfs,config,assets}`, `web/`, `e2e/`), `LICENSE` (MIT), `.gitignore`,
   README stub.
2. Scaffold `web/` frontend: Vite + React + TypeScript + Tailwind via npm, React Router, minimal
   sidebar+content-pane shell wired to placeholder data (no live API yet).
3. Implement `internal/rootfs`: root resolution (CLI arg > UI setting > cwd default), path-join +
   clean + symlink-aware traversal guard (allow symlinks physically under root regardless of
   target), single-level directory listing, atomic runtime root swap.
4. Implement `internal/config`: `os.UserConfigDir()`-based JSON read/write of the persisted root
   path.
5. Implement `internal/markdown`: goldmark pipeline (GFM extension + goldmark-highlighting/chroma
   + custom Mermaid/KaTeX fence-to-placeholder rendering) + bluemonday sanitization policy.
6. Implement `internal/server`: `http.ServeMux` routes (`GET /api/tree`, `GET /api/file`,
   `GET /api/status`, `POST /api/root`) plus static SPA serving from `internal/assets`; `slog`
   structured logging for requests, path-safety rejections, and root changes.
7. Implement `internal/assets`: `go:embed` of `web/dist`.
8. Add build orchestration: `Makefile` `build` target (`npm ci && npm run build` in `web/`, then
   `go build ./cmd/madaview`), plus checked-in placeholder `web/dist/index.html` so a bare
   `go build`/`go vet`/`go test` succeeds on a fresh clone before the frontend is ever built.
9. Implement `cmd/madaview/main.go`: CLI flags (`--root`, `--port` default 4800,
   `-v`/`--verbose`), wiring config + rootfs + markdown + server, binds `0.0.0.0:<port>`.
10. Implement frontend API client + sidebar tree component (lazy per-folder fetch via
    `/api/tree`) + content pane (renders sanitized HTML from `/api/file`) + settings UI (root
    override via `/api/root`, displays `/api/status`).
11. Bundle `mermaid.js` and `katex.js` into the frontend build; add hydration logic that scans
    rendered content for Mermaid/KaTeX placeholders post-render and initializes them in-browser
    only (no network calls).
12. Write `e2e/` Playwright test-loop scenarios (see Test-Loop Design below), each with its own
    `run`/`verify`/`test` and fixture data under `e2e/<scenario>/data/`.
13. Write `.github/workflows/ci.yml`: on push/PR to main, run `go test ./...` and the Playwright
    e2e suite on Linux.
14. Write `.github/workflows/release.yml`: on tag push `v*`, gate on full test suite, then
    cross-compile matrix + native-runner smoke test (launch binary, curl `/api/status`), then
    publish to GitHub Releases.
15. Write README quickstart (download binary, run, open browser) including the documented
    one-line workaround for macOS Gatekeeper / Windows SmartScreen first-run warnings.
16. Write a changelog entry via `/to-changelog`.
17. Remove the originating `.context/TODO.md` item via `/to-todo`, if one exists.

## Test-Loop Design
- **`run`:** for each scenario, reset by deleting any prior output dir, then start the actual
  built `madaview` binary (or Playwright-launched browser session against it) pointed at a
  scenario-specific fixture root under `e2e/<scenario>/data/`, execute the scenario's actions
  (HTTP calls and/or Playwright browser interactions), and write outputs to a result directory:
  captured HTTP responses / DOM snapshots, server stdout logs, metadata (madaview version/commit,
  fixture root path, config used, timestamp, OS/arch).
- **`verify`:** reads the result directory's outputs and metadata, compares actual results
  against each scenario's expected condition (below), and reports good / unexpected / ambiguous
  with root cause traced through the captured logs.
- **Scenarios:**
  1. `render-gfm-basics` → a fixture file with tables, task lists, and a fenced code block
     renders with correct `<table>`, `<input type=checkbox disabled>`, and highlighted `<pre>`
     markup.
  2. `render-mermaid` → a fixture file with a mermaid fence hydrates to an actual `<svg>` element
     inside the placeholder div, with no network requests fired during hydration.
  3. `render-katex` → inline and block math hydrate to rendered KaTeX glyphs (`.katex` class
     present, no network requests).
  4. `path-traversal-rejected` → `../` and absolute-path escape attempts against `/api/file` and
     `/api/tree` return 403/404 with no file content in the response body.
  5. `symlink-inside-root-followed` → a symlink physically under the fixture root, pointing to a
     target outside root, still serves its target's content successfully.
  6. `root-override-cli` → launching with `--root <fixture>` is reflected in `GET /api/status`
     (`currentRoot`, `rootSource: "cli"`).
  7. `root-override-ui-runtime` → `POST /api/root` live-switches the server without restart,
     `/api/status` reflects the new root immediately, and the value persists to the OS config
     dir for the next launch.
  8. `cross-platform-smoke` → CI matrix job per native runner: build, launch the binary, curl
     `/api/status`, expect HTTP 200 with valid JSON.

## Evaluation Criteria
- **Good:** All 8 test-loop scenarios pass on every CI run. A user can download a release binary
  for their OS, run it with no prior install of Node/Python/JVM/etc., and immediately view GFM
  tables/task-lists/code, Mermaid diagrams, and KaTeX math correctly in a browser pointed at
  `http://localhost:4800` (or LAN IP). Attempting to browse outside the configured root
  (`../`, absolute paths) always fails closed (403/404), verified by scenario 4. Changing root via
  CLI or the UI is reflected in `/api/status` within the same request/response cycle (scenarios 6
  and 7).
- **Ambiguous:** OS first-run security warnings (Gatekeeper/SmartScreen) appear but the
  documented one-line workaround resolves them — acceptable per the direction doc's Ambiguous
  Zone, not a failure. A `darwin/amd64` binary builds successfully but cannot be executed in CI
  (no Intel Mac runner) — build-only verification is accepted as ambiguous-but-not-blocking for
  that one target.
- **Bad:** Any scenario in the test-loop fails on a tagged release commit (the release workflow's
  test gate exists specifically to prevent shipping this). A binary requires installing a
  runtime or running a build command before first use. Mermaid/KaTeX rendering issues an external
  network request. A path-traversal or non-root-boundary symlink escape serves content outside
  the configured root when the symlink itself is not under root. Cross-platform smoke test fails
  on any of the three runnable OSes (windows, macos-arm64, linux).
