# Evaluate: madaview-initial-setup

ADR: `.context/adr/20260711-173722-madaview-initial-setup.md`

## What ran

- `go test ./...` — all packages pass (`cmd/madaview` has no test files; `internal/assets`, `internal/config`, `internal/markdown`, `internal/rootfs`, `internal/server` all `ok`).
- `make build` — frontend (`npm ci && npm run build`) + `go build -o madaview ./cmd/madaview` — succeeded.
- All 8 test-loop scenarios under `e2e/` run via each scenario's `test` (run + verify), against the real built binary and a real Chromium session (Playwright) where applicable, executed on this machine (darwin/arm64).

## Good

All 8 scenarios reported `GOOD`, with real (not trivially-passing) checks verified by inspecting each `result/`:

1. **render-gfm-basics** — table, disabled task-list checkboxes, and chroma-highlighted `<pre>` all present in captured DOM.
2. **render-mermaid** — mermaid fence hydrated to a real `<svg>` inside `.mermaid`; `requests.json` shows only `http://localhost:<port>/...` traffic — zero external requests during hydration.
3. **render-katex** — inline and block math hydrated to `.katex`/`.katex-mathml` nodes; requests confirmed same-origin only (font files loaded from local `/assets/`, not a CDN).
4. **path-traversal-rejected** — all 8 attempts (`../` traversal, absolute path, URL-encoded traversal, across both `/api/file` and `/api/tree`) returned genuine 404s with no file content leaked (`attempts.json` inspected directly, not just the verify summary).
5. **symlink-inside-root-followed** — a symlink physically under the fixture root, pointing outside root, served its target's content (200, correct title) — matches the ADR's documented symlink policy.
6. **root-override-cli** — `--root` reflected in `/api/status` as `rootSource: "cli"`.
7. **root-override-ui-runtime** — `POST /api/root` live-switched without restart, `/api/status` reflected the new root in the same cycle, and the value persisted to an isolated fake config dir (`HOME`/`XDG_CONFIG_HOME`/`APPDATA` redirected into the scenario's own `result/`, so this never touched the real user's `~/Library/Application Support/madaview`) — a fresh relaunch with no `--root` picked it up automatically.
8. **cross-platform-smoke** — native `go build` (no npm, placeholder frontend) launched and answered `/api/status` 200 with valid JSON on darwin/arm64, the runner this ran on.

Evaluation Criteria's "Good" bar (all 8 scenarios pass, path traversal fails closed, root changes reflected same-cycle) is fully met on this platform.

## Unexpected / Ambiguous / Needs check

- **Git `HEAD` warning on every scenario run** (ambiguous, not a failure). Every `run` printed:
  `fatal: ambiguous argument 'HEAD': unknown revision or path not in the working tree.`
  **Root cause:** `e2e/lib/harness.mjs`'s `safeGitCommit()` shells out to `git rev-parse HEAD` for `metadata.json`'s `commit` field. This repo has zero commits yet (`git log` confirms "your current branch 'main' does not have any commits yet"). The `try/catch` correctly prevents a crash and falls back to `commit: "unknown"` in every `metadata.json` (verified directly) — so no scenario result is actually wrong — but Node's `execSync` lets the child's stderr leak to the parent's stderr by default, so the git error text prints to the console on every single run regardless of the catch. This is purely cosmetic noise, not a data-correctness issue, and will self-resolve once an initial commit exists.
- **Cross-platform-smoke only exercised darwin/arm64 here.** The ADR's Ambiguous Zone explicitly accepts darwin/amd64 as build-only (no Intel Mac runner) and defers windows/linux verification to CI's native-runner matrix — this local run cannot cover those, which is expected and not a gap in this evaluation, just a boundary of what a single local machine can test.
- Gatekeeper/SmartScreen first-run warnings were not exercised (no downloaded/quarantined release binary in this local dev-build flow) — out of scope for a local run, covered by the ADR's documented Ambiguous Zone instead.

## Patterns

- **Single pattern, single root cause:** the only unexpected signal across all 8 scenarios is the repeated git-HEAD stderr noise, and it traces to one cause (no commits yet in this repo) rather than 8 independent issues. It doesn't affect any scenario's actual pass/fail verdict.

## Next steps

- No `/attack` target — no scenario failed, and the one unexpected item is cosmetic log noise tied to repo state, not application behavior.
- Optional cleanup (not a defect): once the initial commit lands, this warning disappears on its own. If it's worth silencing earlier, `safeGitCommit()` in `e2e/lib/harness.mjs` could pass `stdio: ['ignore', 'pipe', 'ignore']` to `execSync` so a missing-HEAD repo doesn't print to the console during test runs.
- windows/linux native-runner verification of `cross-platform-smoke` remains covered by CI only, per the ADR's own design — nothing to do locally.
