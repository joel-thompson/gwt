package acceptance_test

import (
	"testing"

	"github.com/joel-thompson/gwt/internal/testing/binary"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/golden"
)

func TestVersionFlag(t *testing.T) {
	path := binary.Build(t)
	result := binary.RunCLI(t, path, "--version")

	assert.Check(t, cmp.Equal(result.ExitCode, 0))
	assert.Check(t, golden.String(result.Stdout, t.Name()+".stdout.txt"))
	assert.Check(t, golden.String(result.Stderr, t.Name()+".stderr.txt"))
}
