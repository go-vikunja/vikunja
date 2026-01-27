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

// CheckResult represents the result of a single diagnostic check.
type CheckResult struct {
	Name   string
	Passed bool
	Value  string   // e.g., "vikunja (uid=1000)" or "OK"
	Error  string   // only populated if Passed is false
	Lines  []string // additional lines to display (e.g., list of CORS origins)
}

// CheckGroup represents a category of checks with a header.
type CheckGroup struct {
	Name    string        // e.g., "Database (sqlite)"
	Results []CheckResult
}
