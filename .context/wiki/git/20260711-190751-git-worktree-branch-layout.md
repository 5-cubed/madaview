# Git worktree & branch layout

Each feature is developed by an agent in its own `git worktree`, never directly in the main checkout.

- The main checkout (`madaview-web/`) stays permanently on `main`. It is never used for feature work — only for claiming features (see feature-claiming entry) and merging finished branches.
- Feature worktrees live as siblings to the main checkout, not nested inside it: `madaview-web-worktrees/<slug>/`.
- Branch naming: `feature/<slug>` (e.g. `feature/mermaid-hydration`).
- Worktree directory naming drops the `feature/` prefix since every worktree is a feature by definition: `madaview-web-worktrees/<slug>/`, not `madaview-web-worktrees/feature/<slug>/`.
- No hard cap on how many agents/worktrees run concurrently — judged per session based on how independent the features look (e.g. two features both touching `internal/server` are riskier to run in parallel than one touching `internal/rootfs` and one adding an e2e scenario).
