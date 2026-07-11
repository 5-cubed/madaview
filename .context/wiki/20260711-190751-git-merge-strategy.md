# Git merge strategy for feature branches

- No GitHub remote is configured yet, so there's no PR flow for now — an agent merges its own finished feature branch into `main` itself, from the main worktree.
- Before merging, the full suite must be green: `go test ./...` plus the full Playwright `e2e/` run — every time, not scoped to just the changed package. This matters because concurrent agents land on `main` independently, so a change in one package can silently break another feature's assumptions.
- Merge with `git merge --no-ff feature/<slug>` — preserves every logical commit from the branch plus a merge commit marking where the feature landed. Never squash, never rebase-then-fast-forward.
- If the merge conflicts (e.g. two features both touched `go.mod` or `internal/server`), the merging agent resolves it inline itself using context from both diffs — it doesn't stop and hand it back by default.
- Immediately after a successful merge: `git worktree remove` and `git branch -d` the finished feature branch. Worktrees and branches are disposable; the merge commit on `main` is the permanent record.

Revisit this once a GitHub remote exists — the intent is to layer in real PRs + required review later, this local self-merge flow is a stopgap for the pre-remote phase.
