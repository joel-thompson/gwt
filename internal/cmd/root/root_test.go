package root

import (
	"bytes"
	"testing"

	"gotest.tools/v3/assert"
)

func TestVersionFlag(t *testing.T) {
	cmd := NewCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--version"})

	assert.NilError(t, cmd.Execute())
	assert.Equal(t, out.String(), "gwt version dev\n")
}
