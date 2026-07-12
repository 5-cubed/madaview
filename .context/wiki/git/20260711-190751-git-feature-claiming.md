# Feature claiming via TODO.md

`.context/TODO.md` is the single source of truth for which features are unclaimed, in progress, or done — this is how concurrent agents avoid picking the same feature.

Because a git worktree can't have two checkouts of the same branch, an agent can't edit `main`'s TODO.md while sitting in its own feature worktree. So claiming happens through the dedicated main worktree:

1. Agent switches to the main checkout (always on `main`).
2. Edits `.context/TODO.md`, tagging the item it's taking inline: `[in-progress: feature/<slug>, claimed <date>]`.
3. Commits that claim directly to `main` (no branch/PR ceremony for this bookkeeping commit).
4. Creates `feature/<slug>` and its sibling worktree from that updated `main`.

If an agent abandons a feature partway through: revert to the main worktree, remove the in-progress tag (item goes back to unclaimed backlog), delete the abandoned branch and worktree. No trace is left — the item is available for another agent to pick up fresh.
