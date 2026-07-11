# isolated-adr-runner can't honor an ADR's in-the-moment user-confirmation gate

`.claude/agents/isolated-adr-runner.md` defaults to running an ADR's entire Action
Sequence straight through with no confirmation between steps, on the reasoning that any
mistake is "contained" in its isolated worktree. That reasoning breaks down when an ADR
contains a step that explicitly requires pausing for live user go-ahead before a
hard-to-reverse, publicly-visible action (e.g. `git push origin <tag>` that triggers a
release pipeline and creates a public GitHub Release).

Two structural reasons the agent can't fulfill that gate itself:
1. Git worktrees share the same `.git` and the same `origin` remote as the main
   checkout. A push issued from inside the isolated worktree still hits the real,
   shared remote — the isolation only covers the working directory/branch, not the
   remote. "Contained" is false for anything that pushes to `origin`.
2. The agent's toolset (`Bash, Read, Glob, Edit, Write`) has no way to actually ask the
   user a question — there is no `AskUserQuestion` tool available to it.

**Working pattern:** when delegating such an ADR to `isolated-adr-runner`, scope the
prompt to only the steps before the gate (e.g. the file edit + commit), explicitly
instruct it to stop there and not touch tags/`origin`/releases, then handle the gate
(ask the user) and all remaining steps from the primary session/checkout — merging the
agent's isolated-branch commit onto `main` first if needed (fast-forward is safe/cheap
to check for and do without a separate confirmation, distinct from the gated action
itself).

Related: [[20260711-232551-isolated-adr-runner-skips-todo-claiming]] (a different
structural gap in the same agent — TODO.md claiming — with the same underlying cause:
things that are true "in isolation" aren't true for shared/remote state).
