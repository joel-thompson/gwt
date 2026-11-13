// Package root wires the top-level gwt command.
package root

import (
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

// Version is set at build time via -ldflags "-X .../root.Version=...".
var Version = "dev"

// NewCmd returns the root command. Subcommands are added here as they land.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gwt",
		Short: "Create, list, and remove git worktrees",
		Long: heredoc.Doc(`
			gwt manages git worktrees for the current repository. Worktrees live in
			$GWT_ROOT (default ~/src/worktrees) and are named <repo>-<branch>.

			Messages go to stderr. stdout is reserved for the directory the caller
			should cd into, so a shell wrapper can follow it.
		`),
		Example: heredoc.Doc(`
			# Show the version
			gwt --version

			# Show help for gwt
			gwt --help

			# Show help for a subcommand
			gwt help <command>
		`),
		Version:       Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetOut(os.Stderr)
	cmd.SetErr(os.Stderr)
	return cmd
}

// Execute runs the root command and returns the process exit code.
func Execute() int {
	if err := NewCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	return 0
}
