# gwt

Create, list, and remove git worktrees. A Go rewrite of
`~/src/dotfiles/bin/gwt-worktree`; see [FEATURES.md](FEATURES.md) for status.

## Local development

Run the CLI in dev mode without installing it:

```sh
task dev -- --help
task dev -- <command> <args>
```

Arguments after `--` are forwarded to the CLI. This leaves the existing zsh
script untouched; installation stays disabled until feature parity.

## Git layer

`internal/git` shells out to the installed `git` executable; it does not use a Git
library or a shell. `git.Open(ctx, dir)` discovers the checkout containing `dir`,
including from inside a linked worktree. The returned repository provides:

- `Path()`, `MainPath()`, and `CommonDir()` for checkout and shared metadata paths.
- `Worktrees(ctx)` for structured listings, including detached, locked, and
  prunable worktrees. Branch names omit the `refs/heads/` prefix.
- `HasLocalBranch(ctx, name)` for exact local branch checks, not revision lookup.

Bare repositories cannot be opened directly because they have no checkout.
All Git output is captured, never printed. Execution failures return a
`*git.CommandError` with the arguments, directory, exit status, and Git stderr;
wrapped execution errors and context cancellation remain available through
`errors.Is` / `errors.As`.

Tests use temporary real repositories, with Git configuration isolated from the
user's settings. Run them with `go test ./internal/git`.

## Manual testing sandbox

A local sandbox is available at `~/testing/worktree-testing` (outside this repo):

- `app-main/` and `service-master/`: Git repos with `main` and `master` default
  branches, local bare origins, branches `existing-feature` and `feature/nested`,
  and tag `v1.2.3`.
- Fixtures include a fake `.env`, a `.aigate.yml` symlink target, tracked,
  ignored, and untracked `.claude/` files, a `.claude/worktrees/` exclusion
  fixture, and real `.codegraph/codegraph.db` indexes.
- `worktrees/` and `ai-specs/`: isolated destinations for CLI testing.

Build with `task build` from this repo, then run from either fixture repo:

```sh
export GWT_ROOT=~/testing/worktree-testing/worktrees
export AI_SPECS=~/testing/worktree-testing/ai-specs
/Users/Joel/src/gwt/dist/gwt --help
```

Use the development binary rather than the installed `gwt` shell function.
The sandbox's `README.md` has more testing details. These fixtures support the
planned behavior in [FEATURES.md](FEATURES.md); not all features are implemented yet.
