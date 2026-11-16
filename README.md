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
