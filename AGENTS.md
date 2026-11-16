# gwt

A Go + Cobra rewrite of my git worktree helper, `~/src/dotfiles/bin/gwt-worktree`.
[FEATURES.md](FEATURES.md) is the feature list and the cutover plan. Until every
parity item there is ticked, the zsh script stays in use; don't change it from here.

## Specs and notes live in this repo

For this project, do not use `~/ai_specs/` effort folders. Specs, plans, tickets,
ADRs, and notes all go in this repo. Plans go in `agents/plans/` using
[agents/plan-template.md](agents/plan-template.md).

## Design guidelines

`agents/` is copied from `~/src/circleci-cli/agents/` (clig.dev-based). Read the
matching file before writing the code it governs:

| If you are about to… | Read first |
|---|---|
| Finish any command | [agents/checklist.md](agents/checklist.md) |
| Write any command | [agents/02-basics.md](agents/02-basics.md) |
| Write `Short`, `Long`, or `Example` | [agents/03-help-and-documentation.md](agents/03-help-and-documentation.md) |
| Print output | [agents/04-output.md](agents/04-output.md) |
| Return an error or pick an exit code | [agents/05-errors.md](agents/05-errors.md) |
| Add a flag or argument | [agents/06-arguments-and-flags.md](agents/06-arguments-and-flags.md) |
| Add or restructure a subcommand | [agents/08-subcommands.md](agents/08-subcommands.md) |
| Write or change a test | [agents/14-testing.md](agents/14-testing.md) |

Those files name circleci-cli paths (`clikit/errors`, `internal/cmd/root/usage_test.go`,
`internal/conventions`, `internal/ui`, ...) that don't exist here. Follow the rule, and
create the equivalent package here when a feature first needs it.

## Rules specific to gwt

- stdout carries only the directory for the shell wrapper to `cd` into. Every other
  message goes to stderr. The wrapper breaks if anything else reaches stdout.
- Every command gets `Use`, `Short`, `Long` and `Example` (heredoc, 3+ examples).
- Shell out to `git` rather than adding a git library, to match the script's behavior.

## Layout

```
cmd/gwt/main.go        Entry point; calls root.Execute.
internal/cmd/root/     Root command. One package per subcommand under internal/cmd/.
internal/<domain>/     Business logic (e.g. worktree paths, git calls), no Cobra.
agents/                Design guidelines and plans.
```

## Commands

For local development, use `task dev -- <args>` (e.g. `task dev -- --help`).
This runs the CLI without installing it or replacing the existing zsh script.
Keep the install task commented out until feature parity.

```
task build    # dist/gwt
task test
task lint
task check    # lint + test
```
