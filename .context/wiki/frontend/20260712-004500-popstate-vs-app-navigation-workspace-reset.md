# popstate vs. in-app navigation: different workspace-reset semantics

`web/src/useWorkspace.ts`'s `WorkspaceProvider` treats these as two
deliberately different things, even though both can move the URL between
`/view/*` and `/settings`:

- **In-app navigation** (clicking the Settings link, or a sidebar file
  click while on `/settings`) uses react-router's `navigate()`. The
  workspace reducer state is untouched — `WorkspaceProvider` lives above
  `<Routes>`, so `Workspace` just unmounts/remounts its DOM while the
  underlying tab/pane state survives intact.
- **Browser back/forward** (`popstate`) is *always* treated as a fresh
  page load: the `popstate` handler dispatches `RESET_FROM_URL`, which
  rebuilds a single pane/single tab (or empty workspace) from
  `window.location.pathname`, discarding whatever multi-tab/split state
  existed before. This is intentional (never persisted, never
  reconstructed) — not a bug.

**Why this matters for testing:** an e2e/manual test that returns from
`/settings` via `page.goBack()` (a `popstate`) will observe a reset
workspace, while the same return via clicking a sidebar file or the app's
own nav will observe the *same* tabs/panes as before. These are different,
both-correct code paths — picking the wrong one in a test will produce a
false failure that looks like a persistence bug but is actually a test
picking `popstate` when it meant to exercise the "app-internal round trip"
behavior, or vice versa.
