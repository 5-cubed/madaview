# Git commit conventions

- Header uses Conventional Commits format: `type(scope): summary` (e.g. `feat(rootfs): add symlink-aware traversal guard`), scope maps to the `internal/<package>` structure.
- Body must detail the why, what, and how in depth — not just restate the header.
- Commit granularly: one commit per logical step within a feature branch (e.g. add the guard, then add the e2e scenario, then wire it into handlers), not one giant commit at the end. Each commit should build/test cleanly on its own.
- Do not add a `Co-Authored-By` trailer to commits in this repo — this overrides Claude Code's default behavior of appending one.
