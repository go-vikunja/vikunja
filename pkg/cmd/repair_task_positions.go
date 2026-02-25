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
	repairTaskPositionsCmd.Flags().Bool("dry-run", false, "Preview repairs without making changes")
	repairCmd.AddCommand(repairTaskPositionsCmd)
}

var repairTaskPositionsCmd = &cobra.Command{
	Use:   "task-positions",
	Short: "Detect and repair duplicate task positions across all views",
	Long: `Scans all project views for tasks with duplicate position values and repairs them.

Duplicate positions can occur due to race conditions or historical bugs, causing
tasks to appear in the wrong order or jump around when the page is refreshed.

This command will:
1. Scan all project views for duplicate positions
2. Attempt localized repair by redistributing conflicting tasks
3. Fall back to full view recalculation if localized repair fails

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

		result, err := models.RepairTaskPositions(s, dryRun)
		if err != nil {
			log.Errorf("Failed to repair task positions: %s", err)
			return
		}

		if !dryRun {
			if err := s.Commit(); err != nil {
				log.Errorf("Failed to commit changes: %s", err)
				return
			}
		}

		// Print summary
		log.Infof("Repair complete:")
		log.Infof("  Views scanned: %d", result.ViewsScanned)
		log.Infof("  Views repaired: %d", result.ViewsRepaired)
		log.Infof("  Tasks affected: %d", result.TasksAffected)
		if result.FullRecalcViews > 0 {
			log.Infof("  Views requiring full recalculation: %d", result.FullRecalcViews)
		}

		if len(result.Errors) > 0 {
			log.Errorf("Errors encountered (%d):", len(result.Errors))
			for _, e := range result.Errors {
				log.Errorf("  - %s", e)
			}
		}

		if result.ViewsRepaired == 0 && len(result.Errors) == 0 {
			log.Infof("No position conflicts found - all views are healthy!")
		}
	},
}
