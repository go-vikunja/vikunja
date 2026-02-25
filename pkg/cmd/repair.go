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

package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(repairCmd)
}

var repairCmd = &cobra.Command{
	Use:   "repair",
	Short: "Repair and fix data integrity issues",
	Long: `The repair command provides subcommands to detect and fix various
data integrity issues in your Vikunja installation.

Available repair operations:
  task-positions   - Fix duplicate task positions in project views
  projects         - Fix orphaned projects with missing parents
  file-mime-types  - Detect and set MIME types for files
  orphan-positions - Remove orphaned task position records

Most subcommands support --dry-run to preview changes without applying them.`,
}
