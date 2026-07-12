---
name: e2e-runtime-dirs-must-self-provision
description: e2e scenario data dirs that are entirely gitignored content don't exist on a fresh checkout; run scripts must mkdir them, not assume they exist
metadata:
  type: project
---

`e2e/symlink-inside-root-followed/data/root/` holds only one file,
`linked.md`, which is gitignored and created at runtime by the scenario's
`run` script (a symlink, checked in only via `.gitignore` comment — avoids
symlink portability issues across checkouts/platforms). Git does not track
empty directories, so on a fresh checkout `data/root/` doesn't exist as a
directory at all — only `data/outside/secret.md` (a real tracked file) is
present.

The `run` script called `symlink(target, linkPath)` directly without ever
creating `data/root` first. This passed locally because a leftover
`data/root/` directory persisted on disk from earlier manual runs
(untracked by git, but present on the filesystem) — masking that the
script never provisioned its own directory. It failed in CI with `ENOENT`
on the `symlink()` syscall because CI always starts from a truly fresh
checkout.

**Why:** any e2e scenario `data/` subdirectory whose *only* contents are
gitignored/runtime-generated files is invisible to git and will not exist
on a fresh clone or CI runner, even though it appears to exist on a dev
machine that has run the scenario before.

**How to apply:** any e2e `run` script that creates files at runtime inside
a directory must `await mkdir(dir, { recursive: true })` first, rather than
assuming the directory exists — matches the existing convention in
`e2e/root-override-ui-runtime/run`, which does
`mkdir(configHome, { recursive: true })` for its own runtime-created
directory. Don't rely on a tracked placeholder file (e.g. `.gitkeep`) as
the primary fix — self-provisioning in the script is more robust and
consistent with how the rest of the harness (`e2e/lib/harness.mjs`) already
works. See [[20260711-230407-symlink-e2e-ci-fix]] for the direction that
fixed this.
