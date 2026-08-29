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

package doctor

import (
	"fmt"
	"os"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
)

// CheckDatabase returns database connectivity checks.
func CheckDatabase() CheckGroup {
	dbType := config.DatabaseType.GetString()
	group := CheckGroup{Name: fmt.Sprintf("Database (%s)", dbType)}

	if dbType == "sqlite" {
		if config.DatabasePath.GetString() == db.DatabasePathMemory {
			group.Results = append(group.Results, CheckResult{
				Name:   "Database file",
				Passed: true,
				Value:  "memory (ephemeral, nothing to verify)",
			})
			return group
		}

		fileCheck := checkSqliteFile()
		group.Results = append(group.Results, fileCheck)
		if !fileCheck.Passed {
			// Connecting would create the database file, and a diagnostic that
			// reports "connection OK" against a database it just made is a lie.
			return group
		}
	}

	if _, err := db.CreateDBEngine(); err != nil {
		group.Results = append(group.Results, CheckResult{
			Name:   "Connection",
			Passed: false,
			Error:  err.Error(),
		})
		return group
	}

	group.Results = append(group.Results,
		checkDatabaseConnection(),
		checkDatabaseVersion(dbType),
	)

	if dbType == "postgres" {
		group.Results = append(group.Results, checkParadeDB()...)
	}

	return group
}

func checkSqliteFile() CheckResult {
	result := CheckResult{Name: "Database file"}

	path, err := db.ResolvedDatabasePath()
	if err != nil {
		result.Error = err.Error()
		return result
	}

	info, err := os.Stat(path)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	if info.IsDir() {
		result.Error = fmt.Sprintf("%s exists but is not a file", path)
		return result
	}

	result.Passed = true
	result.Value = path
	return result
}

func checkDatabaseConnection() CheckResult {
	s := db.NewSession()
	defer s.Close()

	if err := s.Ping(); err != nil {
		return CheckResult{
			Name:   "Connection",
			Passed: false,
			Error:  err.Error(),
		}
	}

	return CheckResult{
		Name:   "Connection",
		Passed: true,
		Value:  "OK",
	}
}

func checkDatabaseVersion(dbType string) CheckResult {
	s := db.NewSession()
	defer s.Close()

	var versionQuery string
	switch dbType {
	case "sqlite":
		versionQuery = "SELECT sqlite_version()"
	case "mysql":
		versionQuery = "SELECT version()"
	case "postgres":
		versionQuery = "SELECT version()"
	default:
		return CheckResult{
			Name:   "Server version",
			Passed: false,
			Error:  fmt.Sprintf("unknown database type: %s", dbType),
		}
	}

	results, err := s.QueryString(versionQuery)
	if err != nil {
		return CheckResult{
			Name:   "Server version",
			Passed: false,
			Error:  err.Error(),
		}
	}

	if len(results) == 0 || len(results[0]) == 0 {
		return CheckResult{
			Name:   "Server version",
			Passed: false,
			Error:  "could not retrieve version",
		}
	}

	// Get the first value from the result map
	var version string
	for _, v := range results[0] {
		version = v
		break
	}

	return CheckResult{
		Name:   "Server version",
		Passed: true,
		Value:  version,
	}
}

var paradeDBIndexes = []string{
	"idx_tasks_paradedb",
	"idx_projects_paradedb",
	"idx_time_entries_paradedb",
}

// checkParadeDB reports whether the pg_search extension is installed and, if so,
// whether the bm25 indexes Vikunja relies on exist. A missing extension is not a
// failure — Vikunja falls back to substring search.
func checkParadeDB() []CheckResult {
	s := db.NewSession()
	defer s.Close()

	var version string
	installed, err := s.Table("pg_extension").
		Where("extname = ?", "pg_search").
		Cols("extversion").
		Get(&version)
	if err != nil {
		return []CheckResult{{
			Name:   "ParadeDB",
			Passed: false,
			Error:  err.Error(),
		}}
	}

	if !installed {
		return []CheckResult{{
			Name:   "ParadeDB",
			Passed: true,
			Value:  "not installed (using substring search)",
		}}
	}

	results := []CheckResult{{
		Name:   "ParadeDB",
		Passed: true,
		Value:  "pg_search " + version,
	}}

	var existing []string
	err = s.Table("pg_indexes").
		In("indexname", paradeDBIndexes).
		Cols("indexname").
		Find(&existing)
	if err != nil {
		return append(results, CheckResult{
			Name:   "ParadeDB indexes",
			Passed: false,
			Error:  err.Error(),
		})
	}

	if len(existing) < len(paradeDBIndexes) {
		found := make(map[string]bool, len(existing))
		for _, name := range existing {
			found[name] = true
		}
		var missing []string
		for _, name := range paradeDBIndexes {
			if !found[name] {
				missing = append(missing, name)
			}
		}
		return append(results, CheckResult{
			Name:   "ParadeDB indexes",
			Passed: false,
			Error:  fmt.Sprintf("missing: %s (restart Vikunja to create them)", strings.Join(missing, ", ")),
		})
	}

	return append(results, CheckResult{
		Name:   "ParadeDB indexes",
		Passed: true,
		Value:  fmt.Sprintf("%d present", len(paradeDBIndexes)),
	})
}
