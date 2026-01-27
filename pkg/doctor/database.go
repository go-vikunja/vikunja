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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
)

// CheckDatabase returns database connectivity checks.
func CheckDatabase() CheckGroup {
	dbType := config.DatabaseType.GetString()

	results := []CheckResult{
		checkDatabaseConnection(),
		checkDatabaseVersion(dbType),
	}

	return CheckGroup{
		Name:    fmt.Sprintf("Database (%s)", dbType),
		Results: results,
	}
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
