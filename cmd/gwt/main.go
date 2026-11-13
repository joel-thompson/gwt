// Command gwt manages git worktrees under ~/src/worktrees.
package main

import (
	"os"

	"github.com/joel-thompson/gwt/internal/cmd/root"
)

func main() {
	os.Exit(root.Execute())
}
