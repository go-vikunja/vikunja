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
	rootCmd.AddCommand(deleteOrphanTaskPositions)
}

var deleteOrphanTaskPositions = &cobra.Command{
	Use:   "delete-orphan-task-positions",
	Short: "Removes all task positions for tasks or project views which don't exist anymore.",
	PreRun: func(_ *cobra.Command, _ []string) {
		initialize.FullInitWithoutAsync()
	},
	Run: func(_ *cobra.Command, _ []string) {

		s := db.NewSession()
		defer s.Close()

		count, err := models.DeleteOrphanedTaskPositions(s)
		if err != nil {
			log.Errorf("Could not delete orphaned task positions: %s", err)
			return
		}

		if count == 0 {
			log.Infof("No orphaned task positions found.")
			return
		}

		log.Infof("Successfully deleted %d orphaned task positions.", count)
	},
}
