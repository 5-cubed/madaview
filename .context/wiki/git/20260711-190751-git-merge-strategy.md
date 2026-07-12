# Git merge strategy for feature branches

- A GitHub remote (`origin` → `github.com/5-cubed/madaview.git`) now exists as of 2026-07-11, superseding the earlier "no remote yet" assumption below — but no PR flow has actually started being used: this session pushed a small fix straight to `main` (no branch, no PR) at the user's direct request, and that was accepted without pushback. Treat direct-to-`main` (or self-merged feature branches, per below) as still the live convention until the user explicitly asks for a PR-based flow.
- For feature branches specifically, an agent merges its own finished feature branch into `main` itself, from the main worktree — no PR ceremony.
- Before merging, the full suite must be green: `go test ./...` plus the full Playwright `e2e/` run — every time, not scoped to just the changed package. This matters because concurrent agents land on `main` independently, so a change in one package can silently break another feature's assumptions.
- Merge with `git merge --no-ff feature/<slug>` — preserves every logical commit from the branch plus a merge commit marking where the feature landed. Never squash, never rebase-then-fast-forward.
- If the merge conflicts (e.g. two features both touched `go.mod` or `internal/server`), the merging agent resolves it inline itself using context from both diffs — it doesn't stop and hand it back by default.
- Immediately after a successful merge: `git worktree remove` and `git branch -d` the finished feature branch. Worktrees and branches are disposable; the merge commit on `main` is the permanent record.

Revisit again if the user asks for real PRs + required review — the self-merge/direct-push flow was originally meant as a stopgap for the pre-remote phase, but has continued in practice even after the remote appeared.
