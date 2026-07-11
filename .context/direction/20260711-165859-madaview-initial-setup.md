# Madaview — Initial Project Setup

## Goal
Ship a cross-platform (Windows/macOS/Linux) markdown viewer that end users can download as a single pre-built binary and run immediately with zero build step, zero runtime install (no Node/Python/JVM needed), and zero configuration required for the basic case. The UI is web technology (React) but the distribution model is a native local binary that starts an HTTP server and is viewed in a browser.

## Failure Criteria
- User has to install a language runtime, package manager, or run a build command before they can view a markdown file. That breaks the core "just download and run" promise.
- Binary works on the maintainer's machine but fails to cross-compile/run cleanly on one of the three target OSes (Windows/macOS/Linux) — cross-platform support is a hard requirement, not aspirational.
- Markdown renders incorrectly or not at all for GFM basics (tables, task lists, fenced code, syntax highlighting), Mermaid diagrams, or KaTeX math — these three are explicitly in scope for v1.
- Server exposes the full filesystem instead of respecting the root-folder boundary, creating an unintended security hole over LAN.

## Ambiguous Zone
- OS security warnings (macOS Gatekeeper quarantine, Windows SmartScreen) on first run are acceptable friction for v1, not a failure — as long as there's a documented one-line workaround (right-click > Open on macOS, "More info > Run anyway" on Windows). This is not "zero friction" but is accepted as a reasonable v1 tradeoff.
- LAN-accessible binding (0.0.0.0) is an intentional convenience tradeoff, not a security bug — served content is still scoped to the chosen root folder and below, so exposure is bounded, not unbounded.
- Live-editing markdown is out of scope for v1 but not permanently ruled out — viewing only is the committed initial scope.

## Direction
Build **madaview**: a Go-based local HTTP server binary that embeds a prebuilt React + Tailwind frontend (via `go:embed`), serves a browsable view of markdown files starting from a root folder (defaults to current working directory, overridable via CLI arg, with a settings option in the browser UI), and renders GitHub-flavored markdown plus Mermaid diagrams and KaTeX math. The server binds to `0.0.0.0` so other devices on the same LAN can also view it. Cross-platform binaries (windows/amd64, darwin/amd64+arm64, linux/amd64) are built and published via GitHub Actions to GitHub Releases on each version tag. The repo is public/open source. Code signing/notarization is explicitly deferred; unsigned-binary OS warnings are documented as a known first-run step rather than solved now.

## Constraints
- Backend: Go, compiled to a single self-contained binary per OS/arch, static frontend assets embedded via `go:embed`.
- Frontend: React + Tailwind CSS, built with a dev-time bundler (e.g. Vite); this build step runs only during CI/release, never required of the end user.
- Markdown scope (v1): GitHub-flavored markdown, Mermaid code-fence diagrams, KaTeX math rendering. Viewing only — no in-app editing.
- File access: server-side directory browsing rooted at a chosen folder (not client-side file picker), root resolved by CLI arg > browser-UI setting > current working directory default. Navigation is restricted to the root folder and its subfolders — no traversal to arbitrary system paths.
- Network: binds to `0.0.0.0` (LAN-visible), not localhost-only.
- Distribution: GitHub Actions builds and cross-compiles on tag push, publishes binaries as GitHub Release assets for windows/amd64, darwin/amd64, darwin/arm64, linux/amd64.
- Repo: public, open source.
- Code signing/notarization: not implemented in v1; unsigned-binary security warnings are handled via documentation, not tooling.

## Out of Scope
- Electron packaging (rejected in favor of a lighter native Go binary).
- Single static HTML file distribution (rejected — needed folder browsing beyond what file:// + manual file picking supports).
- Local-server-with-manual-build-per-release workflow (rejected in favor of GitHub Actions automation).
- Live markdown editing (viewing only for v1).
- Code signing / notarization for macOS and Windows (deferred, not solved in this setup).
- Localhost-only binding (rejected — LAN access chosen instead).
