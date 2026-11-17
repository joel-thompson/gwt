package git_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joel-thompson/gwt/internal/git"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
)

func TestOpen(t *testing.T) {
	main := newRepository(t)
	linked := filepath.Join(filepath.Dir(main), "linked checkout")
	gitRun(t, main, "worktree", "add", "-b", "feature/nested", linked)

	for _, checkout := range []string{main, linked} {
		t.Run(filepath.Base(checkout), func(t *testing.T) {
			nested := filepath.Join(checkout, "nested directory")
			assert.NilError(t, os.Mkdir(nested, 0o755))
			repo, err := git.Open(t.Context(), nested)
			assert.NilError(t, err)
			assert.Check(t, cmp.Equal(repo.Path(), checkout))
			assert.Check(t, cmp.Equal(repo.MainPath(), main))
			assert.Check(t, cmp.Equal(repo.CommonDir(), filepath.Join(main, ".git")))
		})
	}

	t.Run("relative directory", func(t *testing.T) {
		cwd, err := os.Getwd()
		assert.NilError(t, err)
		relative, err := filepath.Rel(cwd, main)
		assert.NilError(t, err)
		repo, err := git.Open(t.Context(), relative)
		assert.NilError(t, err)
		assert.Check(t, cmp.Equal(repo.Path(), main))
	})
}

func TestOpenFailures(t *testing.T) {
	main := newRepository(t)
	bare := filepath.Join(filepath.Dir(main), "bare")
	gitRun(t, main, "init", "--bare", bare)

	for _, tc := range []struct {
		name string
		dir  string
	}{
		{"outside repository", filepath.Dir(main)},
		{"missing directory", filepath.Join(main, "missing")},
		{"bare repository", bare},
		{"empty directory argument", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, err := git.Open(t.Context(), tc.dir)
			assert.Check(t, repo == nil)
			assert.Check(t, err != nil)
		})
	}

	t.Run("git diagnostics", func(t *testing.T) {
		_, err := git.Open(t.Context(), filepath.Dir(main))
		var commandErr *git.CommandError
		assert.Assert(t, errors.As(err, &commandErr))
		assert.Check(t, commandErr.ExitCode != 0)
		assert.Check(t, cmp.Contains(commandErr.Stderr, "not a git repository"))
		assert.Check(t, cmp.Equal(commandErr.Dir, filepath.Dir(main)))
		assert.Check(t, cmp.ErrorContains(err, strings.TrimSpace(commandErr.Stderr)))
		var exitErr *exec.ExitError
		assert.Check(t, errors.As(err, &exitErr))
	})

	t.Run("git not installed", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())
		_, err := git.Open(t.Context(), main)
		var commandErr *git.CommandError
		assert.Assert(t, errors.As(err, &commandErr))
		assert.Check(t, cmp.Equal(commandErr.ExitCode, -1))
		assert.Check(t, cmp.ErrorIs(err, exec.ErrNotFound))
	})
}

func TestWorktrees(t *testing.T) {
	main := newRepository(t)
	parent := filepath.Dir(main)
	linked := filepath.Join(parent, "linked space\tand\nnewline")
	detached := filepath.Join(parent, "detached")
	prunable := filepath.Join(parent, "prunable")
	gitRun(t, main, "worktree", "add", "-b", "feature/nested", linked)
	gitRun(t, main, "worktree", "lock", "--reason", "keep this\ncheckout", linked)
	gitRun(t, main, "worktree", "add", "--detach", detached)
	gitRun(t, main, "worktree", "add", "-b", "gone", prunable)
	assert.NilError(t, os.RemoveAll(prunable))

	repo, err := git.Open(t.Context(), linked)
	assert.NilError(t, err)
	worktrees, err := repo.Worktrees(t.Context())
	assert.NilError(t, err)
	head := strings.TrimSpace(gitRun(t, main, "rev-parse", "HEAD"))
	byPath := make(map[string]git.Worktree)
	for _, worktree := range worktrees {
		byPath[worktree.Path] = worktree
	}
	t.Run("lists all worktrees", func(t *testing.T) {
		assert.Check(t, cmp.Len(worktrees, 4))
	})

	t.Run("main worktree is first", func(t *testing.T) {
		assert.Assert(t, len(worktrees) > 0)
		assert.Check(t, cmp.Equal(worktrees[0].Path, main))
		assert.Check(t, cmp.DeepEqual(byPath[main], git.Worktree{
			Path: main, Head: head, Branch: "main",
		}))
	})

	t.Run("linked worktree locked with reason", func(t *testing.T) {
		assert.Check(t, cmp.DeepEqual(byPath[linked], git.Worktree{
			Path: linked, Head: head, Branch: "feature/nested", Locked: true, LockReason: "keep this\ncheckout",
		}))
	})

	t.Run("detached worktree", func(t *testing.T) {
		assert.Check(t, cmp.DeepEqual(byPath[detached], git.Worktree{
			Path: detached, Head: head, Detached: true,
		}))
	})

	t.Run("prunable worktree", func(t *testing.T) {
		assert.Check(t, byPath[prunable].Prunable)
		assert.Check(t, byPath[prunable].PruneReason != "")
	})

	t.Run("lock without reason", func(t *testing.T) {
		gitRun(t, main, "worktree", "lock", detached)
		worktrees, err := repo.Worktrees(t.Context())
		assert.NilError(t, err)
		found := false
		for _, worktree := range worktrees {
			if worktree.Path == detached {
				found = true
				assert.Check(t, worktree.Locked)
				assert.Check(t, cmp.Equal(worktree.LockReason, ""))
			}
		}
		assert.Check(t, cmp.Equal(found, true), "expected detached worktree %q", detached)
	})
}

func TestUnbornRepository(t *testing.T) {
	main := newRepository(t)
	unborn := filepath.Join(filepath.Dir(main), "unborn")
	gitRun(t, main, "init", "-b", "main", unborn)
	repo, err := git.Open(t.Context(), unborn)
	assert.NilError(t, err)
	worktrees, err := repo.Worktrees(t.Context())
	assert.NilError(t, err)
	assert.Assert(t, cmp.Len(worktrees, 1))
	assert.Check(t, cmp.Equal(worktrees[0].Branch, "main"))
	exists, err := repo.HasLocalBranch(t.Context(), "main")
	assert.NilError(t, err)
	assert.Check(t, !exists)
}

func TestHasLocalBranch(t *testing.T) {
	main := newRepository(t)
	gitRun(t, main, "branch", "feature/nested")
	gitRun(t, main, "tag", "tag-only")
	gitRun(t, main, "update-ref", "refs/remotes/origin/remote-only", "HEAD")
	repo, err := git.Open(t.Context(), main)
	assert.NilError(t, err)

	for _, tc := range []struct {
		branch string
		want   bool
	}{
		{"main", true},
		{"feature/nested", true},
		{"missing", false},
		{"tag-only", false},
		{"origin/remote-only", false},
		{"HEAD", false},
		{"main~1", false},
		{"--help", false},
		{"", false},
	} {
		t.Run(tc.branch, func(t *testing.T) {
			exists, err := repo.HasLocalBranch(t.Context(), tc.branch)
			assert.NilError(t, err)
			assert.Check(t, cmp.Equal(exists, tc.want))
		})
	}

	t.Run("repository failure is not absence", func(t *testing.T) {
		assert.NilError(t, os.RemoveAll(filepath.Join(main, ".git")))
		exists, err := repo.HasLocalBranch(t.Context(), "main")
		assert.Check(t, !exists)
		assert.Check(t, err != nil)
	})
}

func TestCanceledContext(t *testing.T) {
	main := newRepository(t)
	repo, err := git.Open(t.Context(), main)
	assert.NilError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = git.Open(ctx, main)
	assert.Check(t, cmp.ErrorIs(err, context.Canceled))
	_, err = repo.Worktrees(ctx)
	assert.Check(t, cmp.ErrorIs(err, context.Canceled))
	_, err = repo.HasLocalBranch(ctx, "main")
	assert.Check(t, cmp.ErrorIs(err, context.Canceled))
}

func newRepository(t *testing.T) string {
	t.Helper()
	// Keep the developer's Git environment, configuration, hooks, and templates out of fixtures.
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(key, "GIT_") {
			t.Setenv(key, "")
			assert.NilError(t, os.Unsetenv(key))
		}
	}
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_TERMINAL_PROMPT", "0")
	t.Setenv("LC_ALL", "C")
	parent, err := filepath.EvalSymlinks(t.TempDir())
	assert.NilError(t, err)
	main := filepath.Join(parent, "main checkout")
	gitRun(t, parent, "-c", "init.templateDir=", "init", "-b", "main", main)
	gitRun(t, main, "config", "user.name", "Test User")
	gitRun(t, main, "config", "user.email", "test@example.invalid")
	gitRun(t, main, "-c", "commit.gpgsign=false", "commit", "--allow-empty", "-m", "initial")
	return main
}

func gitRun(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	assert.NilError(t, err, "git %v: %s", args, out)
	return string(out)
}
