// Package commands wires the cobra command tree. Each subcommand lives in a
// sibling file; root.go owns shared flags, the global error handler, and the
// JSON output toggle.
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/output"
)

// Globals carries flags shared across subcommands. The pointer is bound onto
// the root command's persistent flags; subcommands read it via PostRun.
type Globals struct {
	JSON    bool
	Verbose bool
}

var globals Globals

// Root builds the cobra command tree.
func Root(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "veans",
		Short:         "veans — a beans-shaped CLI for Vikunja",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}
	root.PersistentFlags().BoolVar(&globals.JSON, "json", false, "emit JSON output")
	root.PersistentFlags().BoolVar(&globals.Verbose, "verbose", false, "verbose logging to stderr")

	root.AddCommand(newVersionCmd(version))
	root.AddCommand(newInitCmd())
	root.AddCommand(newListCmd())
	root.AddCommand(newShowCmd())

	return root
}

// Execute runs the cobra tree and converts errors into the structured output
// envelope. It returns the desired exit code.
func Execute(version string) int {
	cmd := Root(version)
	if err := cmd.Execute(); err != nil {
		output.EmitError(globals.JSON, err, os.Stderr)
		return 1
	}
	return 0
}

func newVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the veans version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
}
