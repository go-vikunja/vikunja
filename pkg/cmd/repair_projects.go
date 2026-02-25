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
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"

	"github.com/spf13/cobra"
)

func init() {
	repairProjectsCmd.Flags().Bool("dry-run", false, "Preview repairs without making changes")
	rootCmd.AddCommand(repairProjectsCmd)
}

var repairProjectsCmd = &cobra.Command{
	Use:   "repair-projects",
	Short: "Repair orphaned projects whose parent project no longer exists",
	Long: `Finds projects whose parent_project_id references a project that no longer
exists in the database and re-parents them to the top level (parent_project_id = 0).

This can happen when a parent project is deleted but its sub-projects are not
fully cleaned up, for example after importing from external services like Trello.

Orphaned projects cannot be un-archived, modified, or deleted through the UI
because permission checks fail when traversing the broken parent chain.

Use --dry-run to preview what would be fixed without making changes.`,
	PreRun: func(_ *cobra.Command, _ []string) {
		initialize.FullInitWithoutAsync()
	},
	Run: func(cmd *cobra.Command, _ []string) {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		s := db.NewSession()
		defer s.Close()

		if dryRun {
			log.Infof("Running in dry-run mode - no changes will be made")
		}

		result, err := models.RepairOrphanedProjects(s, dryRun)
		if err != nil {
			log.Errorf("Failed to repair orphaned projects: %s", err)
			return
		}

		if !dryRun {
			if err := s.Commit(); err != nil {
				log.Errorf("Failed to commit changes: %s", err)
				return
			}
		}

		log.Infof("Repair complete:")
		log.Infof("  Orphaned projects found: %d", result.Found)
		log.Infof("  Projects repaired: %d", result.Repaired)

		if result.Found == 0 {
			log.Infof("No orphaned projects found - all parent references are valid!")
		}
	},
}
