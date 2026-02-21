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
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(repairFileMimeTypesCmd)
}

var repairFileMimeTypesCmd = &cobra.Command{
	Use:   "repair-file-mime-types",
	Short: "Detect and set MIME types for all files that have none",
	Long: `Scans all files in the database that have no MIME type set,
detects the type from the stored file content, and updates the database.

This is useful after upgrading from a version that did not store MIME types
on file creation. Only files with an empty or NULL mime column are affected.`,
	PreRun: func(_ *cobra.Command, _ []string) {
		initialize.FullInitWithoutAsync()
	},
	Run: func(_ *cobra.Command, _ []string) {
		s := db.NewSession()
		defer s.Close()

		if err := s.Begin(); err != nil {
			log.Errorf("Failed to start transaction: %s", err)
			return
		}
		defer func() {
			_ = s.Rollback()
		}()

		result, err := files.RepairFileMimeTypes(s)
		if err != nil {
			log.Errorf("Failed to repair file MIME types: %s", err)
			return
		}

		if err := s.Commit(); err != nil {
			log.Errorf("Failed to commit changes: %s", err)
			return
		}

		log.Infof("Repair complete:")
		log.Infof("  Files scanned: %d", result.Total)
		log.Infof("  Files updated: %d", result.Updated)

		if len(result.Errors) > 0 {
			log.Errorf("Errors encountered (%d):", len(result.Errors))
			for _, e := range result.Errors {
				log.Errorf("  - %s", e)
			}
		}

		if result.Total == 0 {
			log.Infof("No files with missing MIME types found - all files are healthy!")
		}
	},
}
