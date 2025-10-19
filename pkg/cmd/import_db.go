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
	"os"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/services"

	"github.com/spf13/cobra"
)

var (
	importDBSQLiteFile string
	importDBFilesDir   string
	importDBDryRun     bool
	importDBQuiet      bool
)

func init() {
	rootCmd.AddCommand(importDBCmd)
	importDBCmd.Flags().StringVarP(&importDBSQLiteFile, "sqlite-file", "s", "", "Path to the SQLite database file to import (required)")
	importDBCmd.Flags().StringVarP(&importDBFilesDir, "files-dir", "f", "", "Path to the files directory to migrate (optional)")
	importDBCmd.Flags().BoolVarP(&importDBDryRun, "dry-run", "d", false, "Perform a dry run without making any changes")
	importDBCmd.Flags().BoolVarP(&importDBQuiet, "quiet", "q", false, "Suppress progress output")
	_ = importDBCmd.MarkFlagRequired("sqlite-file")
}

var importDBCmd = &cobra.Command{
	Use:   "import-db",
	Short: "Import data from a SQLite database file (primary migration path)",
	Long: `Import data from a SQLite database file into the current Vikunja instance.

This command is the primary migration path for moving from one Vikunja instance
to another, especially when migrating from SQLite to PostgreSQL or MySQL.

The import process:
1. Validates the SQLite file and target database
2. Reads all data from the SQLite database
3. Transforms data to match the current schema
4. Imports data in a single transaction (all-or-nothing)
5. Optionally migrates files from the old instance

Examples:
  # Import SQLite database only
  vikunja import-db --sqlite-file=/path/to/vikunja.db

  # Import with files
  vikunja import-db --sqlite-file=/path/to/vikunja.db --files-dir=/path/to/files

  # Dry run to test without making changes
  vikunja import-db --sqlite-file=/path/to/vikunja.db --dry-run

  # Quiet mode (no progress output)
  vikunja import-db --sqlite-file=/path/to/vikunja.db --quiet
`,
	PreRun: func(_ *cobra.Command, _ []string) {
		// Initialize database and config
		initialize.FullInitWithoutAsync()
	},
	Run: func(_ *cobra.Command, _ []string) {
		// Validate SQLite file exists
		if _, err := os.Stat(importDBSQLiteFile); err != nil {
			log.Criticalf("SQLite file not found or not accessible: %s", importDBSQLiteFile)
			return
		}

		// Validate files directory if provided
		if importDBFilesDir != "" {
			if info, err := os.Stat(importDBFilesDir); err != nil {
				log.Criticalf("Files directory not found or not accessible: %s", importDBFilesDir)
				return
			} else if !info.IsDir() {
				log.Criticalf("Files path is not a directory: %s", importDBFilesDir)
				return
			}
		}

		// Get database engine
		engine := db.GetEngine()

		// Create service registry and import service
		registry := services.NewServiceRegistry(engine)
		importService := registry.SQLiteImport()

		// Prepare import options
		opts := services.ImportOptions{
			SQLiteFile: importDBSQLiteFile,
			FilesDir:   importDBFilesDir,
			DryRun:     importDBDryRun,
			Quiet:      importDBQuiet,
		}

		// Display import configuration
		if !importDBQuiet {
			log.Info("========================================")
			log.Info("  Vikunja Database Import")
			log.Info("========================================")
			log.Infof("SQLite File: %s", importDBSQLiteFile)
			if importDBFilesDir != "" {
				log.Infof("Files Directory: %s", importDBFilesDir)
			} else {
				log.Info("Files Directory: (none - files will not be migrated)")
			}
			if importDBDryRun {
				log.Info("Mode: DRY RUN (no changes will be made)")
			} else {
				log.Info("Mode: LIVE IMPORT")
			}
			log.Info("========================================")
			log.Info("")
		}

		// Perform import
		report, err := importService.ImportFromSQLite(opts)
		if err != nil {
			log.Criticalf("Import failed: %v", err)
			if len(report.Errors) > 0 {
				log.Error("Errors encountered during import:")
				for _, e := range report.Errors {
					log.Errorf("  - %s", e)
				}
			}
			return
		}

		// Display import report
		if !importDBQuiet {
			log.Info("")
			log.Info("========================================")
			log.Info("  Import Report")
			log.Info("========================================")

			if importDBDryRun {
				log.Info("DRY RUN COMPLETED - No changes were made")
			} else {
				if report.Success {
					log.Info("✓ Import completed successfully!")
				} else {
					log.Warning("⚠ Import completed with warnings")
				}
			}

			log.Infof("Duration: %s", report.Duration)
			log.Info("")
			log.Info("Entity Counts:")
			log.Infof("  Users:              %d", report.Counts.Users)
			log.Infof("  Teams:              %d", report.Counts.Teams)
			log.Infof("  Team Members:       %d", report.Counts.TeamMembers)
			log.Infof("  Projects:           %d", report.Counts.Projects)
			log.Infof("  Tasks:              %d", report.Counts.Tasks)
			log.Infof("  Labels:             %d", report.Counts.Labels)
			log.Infof("  Task-Label Links:   %d", report.Counts.TaskLabels)
			log.Infof("  Comments:           %d", report.Counts.Comments)
			log.Infof("  Attachments:        %d", report.Counts.Attachments)
			log.Infof("  Buckets:            %d", report.Counts.Buckets)
			log.Infof("  Saved Filters:      %d", report.Counts.SavedFilters)
			log.Infof("  Subscriptions:      %d", report.Counts.Subscriptions)
			log.Infof("  Project Views:      %d", report.Counts.ProjectViews)
			log.Infof("  Project Backgrounds:%d", report.Counts.ProjectBackgrounds)
			log.Infof("  Link Shares:        %d", report.Counts.LinkShares)
			log.Infof("  Webhooks:           %d", report.Counts.Webhooks)
			log.Infof("  Reactions:          %d", report.Counts.Reactions)
			log.Infof("  API Tokens:         %d", report.Counts.APITokens)
			log.Infof("  Favorites:          %d", report.Counts.Favorites)

			if importDBFilesDir != "" {
				log.Info("")
				log.Info("File Migration:")
				log.Infof("  Files Processed:    %d", report.Counts.Files)
				log.Infof("  Files Copied:       %d", report.Counts.FilesCopied)
				log.Infof("  Files Failed:       %d", report.Counts.FilesFailed)

				if report.FilesError != nil {
					log.Warningf("  Files Error:        %v", report.FilesError)
				}
			}

			if len(report.Errors) > 0 {
				log.Info("")
				log.Warning("Errors encountered:")
				for _, e := range report.Errors {
					log.Warningf("  - %s", e)
				}
			}

			log.Info("========================================")

			// Calculate totals
			totalEntities := report.Counts.Users + report.Counts.Teams +
				report.Counts.Projects + report.Counts.Tasks +
				report.Counts.Labels + report.Counts.Comments +
				report.Counts.Attachments

			if !importDBDryRun && report.Success {
				log.Info("")
				log.Infof("✓ Successfully imported %d entities in %s", totalEntities, report.Duration)
				if importDBFilesDir != "" && report.Counts.FilesCopied > 0 {
					log.Infof("✓ Migrated %d files", report.Counts.FilesCopied)
				}
			} else if importDBDryRun {
				log.Info("")
				log.Infof("Dry run validated - ready to import %d entities", totalEntities)
			}
		}

		// Exit with appropriate code
		if !report.Success {
			os.Exit(1)
		}
	},
}
