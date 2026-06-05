// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package commands wires the cobra command tree. Each subcommand lives in a
// sibling file. The agent-facing commands (list/show/create/update/claim/api)
// emit JSON unconditionally; only init and login speak human prose.
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/output"
)

// Root builds the cobra command tree.
func Root(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "veans",
		Short:         "veans — a beans-shaped CLI for Vikunja",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}

	root.AddCommand(newVersionCmd(version))
	root.AddCommand(newInitCmd())
	root.AddCommand(newListCmd())
	root.AddCommand(newShowCmd())
	root.AddCommand(newCreateCmd())
	root.AddCommand(newUpdateCmd())
	root.AddCommand(newClaimCmd())
	root.AddCommand(newPrimeCmd())
	root.AddCommand(newAPICmd())
	root.AddCommand(newLoginCmd())

	return root
}

// Execute runs the cobra tree and converts errors into the structured output
// envelope. Errors land on stderr as JSON `{code, error}` and the process
// exits non-zero — both agent-facing and human-facing commands share this
// shape so callers can branch on `code` regardless of which command they ran.
func Execute(version string) int {
	cmd := Root(version)
	if err := cmd.Execute(); err != nil {
		output.EmitError(err, os.Stderr)
		return 1
	}
	return 0
}

func newVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the veans version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(version)
		},
	}
}
