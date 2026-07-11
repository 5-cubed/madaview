# isolated-adr-runner worktree cleanup belongs to the caller, not the agent

`.claude/agents/isolated-adr-runner.md` runs with `isolation: worktree` and its completion
criterion requires the report to include the worktree path/branch "so the caller can locate
and review the result." That means the worktree must still exist when the agent finishes —
if the agent deleted its own worktree as its last step, there would be nothing left to
review or merge.

**Decision (confirmed via `/create-agent` grilling on 2026-07-12):** the agent itself never
runs `git worktree remove` / `git worktree prune`. Cleanup happens from the primary/caller
session, and only after the caller has reviewed and merged (or otherwise finished with) the
agent's branch:

```bash
git worktree remove <path>   # once merged/reviewed and no longer needed
git worktree prune
```

If the branch itself turns out to have no unique commits over `main` (e.g. the agent made no
changes, or its commit was already fast-forward merged), the local branch ref can be deleted
too, but that's a separate, more destructive decision — don't bundle it into routine worktree
cleanup without confirming.

Related: [[20260712-002000-isolated-adr-runner-cannot-honor-user-confirmation-gates]] and
[[20260711-232551-isolated-adr-runner-skips-todo-claiming]] — same root pattern: things that
are true "in isolation" (the agent's own worktree) aren't things the agent can safely act on
for shared/reviewable state.
