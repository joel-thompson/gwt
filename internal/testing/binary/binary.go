// Package binary builds and runs gwt for end-to-end tests.
package binary

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

// Build compiles the current source into a test-owned directory.
func Build(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "gwt")
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", path, "github.com/joel-thompson/gwt/cmd/gwt")
	out, err := cmd.CombinedOutput()
	assert.NilError(t, err, "build gwt: %s", out)
	return path
}

// Result captures the process's public outputs.
type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// RunCLI runs a binary with isolated working and data directories and no stdin.
func RunCLI(t *testing.T, path string, args ...string) Result {
	t.Helper()
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"HOME="+dir,
		"GWT_ROOT="+filepath.Join(dir, "worktrees"),
		"AI_SPECS="+filepath.Join(dir, "specs"),
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	assert.NilError(t, ctx.Err(), "gwt %v timed out: %s", args, stderr.String())
	if err != nil {
		var exitErr *exec.ExitError
		assert.Assert(t, errors.As(err, &exitErr), "run gwt: %v", err)
	}
	return Result{
		Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: cmd.ProcessState.ExitCode(),
	}
}
