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
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/migration"
	"github.com/spf13/cobra"
)

func init() {
	migrateCmd.AddCommand(migrateListCmd)
	migrationRollbackCmd.Flags().StringVarP(&rollbackUntilFlag, "name", "n", "", "The id of the migration you want to roll back until.")
	_ = migrationRollbackCmd.MarkFlagRequired("name")
	migrateCmd.AddCommand(migrationRollbackCmd)
	rootCmd.AddCommand(migrateCmd)
}

// TODO: add args to run migrations up or down, until a certain point etc
// Rollback until
// list -> Essentially just show the table, maybe with an extra column if the migration did run or not
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run all database migrations which didn't already run.",
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		initialize.LightInit()
	},
	Run: func(_ *cobra.Command, _ []string) {
		migration.Migrate(nil)
	},
}

var migrateListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show a list with all database migrations.",
	Run: func(_ *cobra.Command, _ []string) {
		migration.ListMigrations()
	},
}

var rollbackUntilFlag string

var migrationRollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Roll migrations back until a certain point.",
	Run: func(_ *cobra.Command, _ []string) {
		migration.Rollback(rollbackUntilFlag)
	},
}
