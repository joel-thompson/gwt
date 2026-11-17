// Package git provides repository and worktree queries through the Git executable.
// It captures all output; callers own user-facing messages.
package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repository is anchored to a checkout, independent of the process working directory.
// Open it again after moving or removing that checkout.
type Repository struct {
	path      string
	mainPath  string
	commonDir string
}

// Open discovers the checkout containing dir, which may be a subdirectory of a
// main or linked worktree. Bare repositories have no checkout and are rejected.
func Open(ctx context.Context, dir string) (*Repository, error) {
	if dir == "" {
		return nil, errors.New("open repository: directory must not be empty")
	}
	path, err := run(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, err
	}
	repo := &Repository{path: strings.TrimSuffix(string(path), "\n")}
	common, err := run(ctx, repo.path, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return nil, err
	}
	repo.commonDir = strings.TrimSuffix(string(common), "\n")
	worktrees, err := repo.Worktrees(ctx)
	if err != nil {
		return nil, err
	}
	if len(worktrees) == 0 {
		return nil, errors.New("open repository: git reported no worktrees")
	}
	// Git lists the main worktree first, including when called from a linked one.
	repo.mainPath = worktrees[0].Path
	return repo, nil
}

func (r *Repository) Path() string { return r.path }

func (r *Repository) MainPath() string { return r.mainPath }

// CommonDir is the shared Git directory, not a linked worktree's private metadata.
func (r *Repository) CommonDir() string { return r.commonDir }

// Worktree describes one entry in git worktree list. Branch is a local branch
// name without refs/heads/, or empty for detached HEADs and bare repositories.
type Worktree struct {
	Path        string
	Head        string
	Branch      string
	Bare        bool
	Detached    bool
	Locked      bool
	LockReason  string
	Prunable    bool
	PruneReason string
}

func (r *Repository) Worktrees(ctx context.Context) ([]Worktree, error) {
	out, err := run(ctx, r.path, "worktree", "list", "--porcelain", "-z")
	if err != nil {
		return nil, err
	}
	var worktrees []Worktree
	var current Worktree
	for _, field := range strings.Split(string(out), "\x00") {
		if field == "" {
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = Worktree{}
			}
			continue
		}
		key, value, _ := strings.Cut(field, " ")
		switch key {
		case "worktree":
			current.Path = value
		case "HEAD":
			current.Head = value
		case "branch":
			current.Branch = strings.TrimPrefix(value, "refs/heads/")
		case "bare":
			current.Bare = true
		case "detached":
			current.Detached = true
		case "locked":
			current.Locked, current.LockReason = true, value
		case "prunable":
			current.Prunable, current.PruneReason = true, value
		}
	}
	return worktrees, nil
}

// HasLocalBranch checks an exact local branch name, not a tag, revision, or remote ref.
// A missing branch is false, nil; execution and repository failures remain errors.
func (r *Repository) HasLocalBranch(ctx context.Context, branch string) (bool, error) {
	_, err := run(ctx, r.path, "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	if err == nil {
		return true, nil
	}
	var commandErr *CommandError
	if errors.As(err, &commandErr) && commandErr.ExitCode == 1 && ctx.Err() == nil {
		return false, nil
	}
	return false, err
}

// CommandError retains Git diagnostics and wraps the execution failure.
// ExitCode is -1 if Git could not start or was terminated by a signal.
type CommandError struct {
	Dir      string
	Args     []string
	ExitCode int
	Stderr   string
	cause    error
}

func (e *CommandError) Error() string {
	message := fmt.Sprintf("git %s in %q: %v", strings.Join(e.Args, " "), e.Dir, e.cause)
	if detail := strings.TrimSpace(e.Stderr); detail != "" {
		message += ": " + detail
	}
	return message
}

func (e *CommandError) Unwrap() error { return e.cause }

func run(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = filepath.Clean(dir)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		exitCode := -1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
		if ctx.Err() != nil {
			err = errors.Join(ctx.Err(), err)
		}
		return nil, &CommandError{
			Dir: dir, Args: append([]string(nil), args...), ExitCode: exitCode,
			Stderr: stderr.String(), cause: err,
		}
	}
	return stdout.Bytes(), nil
}
