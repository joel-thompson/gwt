# Features

The Go rewrite of my worktree helper. Today the work is done by
`~/src/dotfiles/bin/gwt-worktree` (zsh), wrapped by the `gwt` / `gwtr` zsh
functions in `~/src/dotfiles/.zshrc_all` (lines 70–82), which run the script
and `cd` into the path it prints. `gwtl` is an alias for `git worktree list`.

Everything below is a parity target with that script unless marked new. Tick
items off as they land.

## Cutover plan

Keep using the zsh script until every parity item below is done. Once it is:

1. Install the Go binary (`task install`).
2. Point the `gwt` / `gwtr` functions in `~/src/dotfiles/.zshrc_all` at it
   (or replace them with `eval "$(gwt shell-init zsh)"`).
3. Update `~/.claude/CLAUDE.md`, which tells agents to run
   `~/src/dotfiles/bin/gwt-worktree add|rm|list`.
4. Delete `bin/gwt-worktree` from dotfiles after a week of daily use with no
   fallbacks.

## Output contract

- [ ] Messages go to stderr; stdout carries only the directory the caller
      should `cd` into (a process cannot change its parent shell's directory).
- [ ] Non-zero exit on any failure; nothing on stdout when it fails.

## Configuration

- [ ] `GWT_ROOT` env var for where worktrees live (default `~/src/worktrees`).
- [ ] `AI_SPECS` env var for the cross-repo specs folder (default `~/ai_specs`).
- [ ] Worktree path is `$GWT_ROOT/<basename of main checkout>-<branch>`, the same
      from the main checkout or from inside any linked worktree.

## add `<branch> [start-point]`

- [ ] Fail with a clear message outside a git repo, or with no branch given.
- [ ] Create `$GWT_ROOT` if missing.
- [ ] Existing local branch: check it out into the new worktree.
- [ ] Start point given: create the branch from it with `--no-track`.
- [ ] Neither: `git fetch origin`, then branch from the default remote branch
      with `--no-track`. Abort if the fetch fails.
- [ ] Default branch resolution: `origin/HEAD`, then `origin/main`, then
      `origin/master`; if none, tell the user to pass a start point.
- [ ] Note when the main checkout has uncommitted changes (they stay put).
- [ ] Copy untracked/ignored files under `.claude/` into the worktree, skipping
      `.claude/worktrees/`, and report how many.
- [ ] Symlink `.aigate.yml` from the main checkout if present, and add
      `/.aigate.yml` to the repo's `info/exclude` so `git worktree remove`
      doesn't refuse.
- [ ] Create `$AI_SPECS` if missing.
- [ ] Copy `.env` (copy, not link) if present and not already in the worktree.
- [ ] Copy `.codegraph/codegraph.db` via `sqlite3 .backup`, then run
      `codegraph sync -q` in the worktree.
- [ ] Print the worktree path on stdout.

## rm `<branch>`

- [ ] Fail with a clear message outside a git repo, or with no branch given.
- [ ] If the caller is inside the worktree being removed, step out to the main
      checkout first and print that path on stdout so the wrapper follows.
- [ ] `git worktree remove` (no force).
- [ ] Delete the local branch: `-d` if merged, else `-D` and print the SHA with
      a `git branch <branch> <sha>` recovery command. Report if it can't be
      deleted (e.g. checked out elsewhere).

## list

- [ ] Equivalent of `git worktree list`.

## Help

- [ ] Usage text with the examples the script shows today (`gwt my-feature`,
      `gwt my-feature HEAD`, `gwt my-feature v1.2.3`).

## Shell integration (new, needed for cutover)

- [ ] `gwt shell-init zsh` prints `gwt` / `gwtr` / `gwtl` wrapper functions that
      run the binary and `cd` into its stdout, replacing the hand-written ones in
      `.zshrc_all`. The binary name and the wrapper name collide, so the wrapper
      must call the binary by absolute path or `command gwt`.
