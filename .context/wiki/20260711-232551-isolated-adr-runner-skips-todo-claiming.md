# isolated-adr-runner agent intentionally skips TODO.md claiming

`.claude/agents/isolated-adr-runner.md` executes an ADR's action sequence via `/auto-action` inside `isolation: worktree` — the harness places it directly into its own isolated worktree/branch before it runs at all.

This conflicts with the TODO.md-claiming flow in
[[20260711-190751-git-feature-claiming.md]]: that flow requires editing and
committing `.context/TODO.md` **from the main checkout**, because "an agent
can't edit main's TODO.md while sitting in its own feature worktree." Since
`isolation: worktree` means the agent never sits in main to begin with, it
structurally cannot perform that claim step.

**Decision:** `isolated-adr-runner` does not attempt TODO.md claiming at
all — that responsibility, if needed, belongs to a separate step done from
the main checkout *before* delegating to this agent. This also means the
agent doesn't create the project's specific `madaview-web-worktrees/<slug>`
/ `feature/<slug>` naming from
[[20260711-190751-git-worktree-branch-layout.md]] — it uses whatever
worktree path/branch the harness's generic isolation mechanism assigns,
and reports that path/branch back in its final output so the caller can
find it. This was a deliberate tradeoff (confirmed twice during
`/create-agent` grilling), not an oversight — accepted in exchange for not
having to hand-roll `git worktree add` plumbing inside the agent itself.

Related: [[20260712-002000-isolated-adr-runner-cannot-honor-user-confirmation-gates]]
(a different structural gap in the same agent — honoring an ADR's live
user-confirmation gate before a shared/remote action — with the same root cause).
