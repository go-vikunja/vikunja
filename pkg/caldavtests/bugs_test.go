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

package caldavtests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBugs contains tests that reproduce specific bugs reported by users.
// Each test references the GitHub issue it reproduces.
// These tests are expected to FAIL until the bug is fixed.
//
// To add a new bug reproduction test:
// 1. Create a new t.Run with the issue number in the name
// 2. Reproduce the exact CalDAV request sequence from the bug report
// 3. Assert what the correct behavior SHOULD be (not what it currently does)
// 4. The test will fail until the bug is fixed — this is expected and good

func TestBugs(t *testing.T) {
	// Template for adding bug reproductions:
	//
	// t.Run("GitHub_Issue_NNNN_short_description", func(t *testing.T) {
	//     e := setupTestEnv(t)
	//
	//     // Reproduce the steps from the issue...
	//     vtodo := NewVTodo("issue-NNNN", "...").Build()
	//     rec := caldavPUT(t, e, "/dav/projects/36/issue-NNNN.ics", vtodo)
	//
	//     // Assert the expected (correct) behavior
	//     assert.Equal(t, 201, rec.Code)
	// })

	// A client that still holds the href from before the task moved to another
	// project completes it there. The stale collection must not gain a copy.
	t.Run("GitHub_Issue_3482_completed_task_duplicated_into_stale_collection", func(t *testing.T) {
		e := setupTestEnv(t)

		const uid = "issue-3482"
		created := caldavPUT(t, e, "/dav/projects/36/"+uid+".ics", NewVTodo(uid, "Task G").Build())
		assert.Equal(t, http.StatusCreated, created.Code)

		stale := caldavPUT(t, e, "/dav/projects/38/"+uid+".ics", NewVTodo(uid, "Task G").Status("COMPLETED").Build())
		assert.Equal(t, http.StatusNotFound, stale.Code)

		assert.Equal(t, http.StatusNotFound, caldavGET(t, e, "/dav/projects/38/"+uid+".ics").Code)
		assert.Equal(t, http.StatusOK, caldavGET(t, e, "/dav/projects/36/"+uid+".ics").Code)

		report := caldavREPORT(t, e, "/dav/projects/38/", ReportCalendarQuery)
		assert.NotContains(t, report.Body.String(), uid)
	})
}
