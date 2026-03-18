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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientDAVx5Flow(t *testing.T) {
	t.Run("Full DAVx5 sync flow", func(t *testing.T) {
		e := setupTestEnv(t)

		// Step 1: Discover principal
		// DAVx5 sends PROPFIND to the server root or well-known URL
		rec := caldavPROPFIND(t, e, "/dav/", "0", PropfindCurrentUserPrincipal)
		assert.True(t, rec.Code == 207 || rec.Code == 301,
			"Step 1: PROPFIND /dav/ should return 207 or redirect. Got %d", rec.Code)

		// Step 2: Get calendar-home-set from principal
		rec = caldavPROPFIND(t, e, "/dav/principals/user15/", "0", PropfindCalendarHomeSet)
		assertResponseStatus(t, rec, 207)
		assert.Contains(t, rec.Body.String(), "calendar-home-set",
			"Step 2: Principal should advertise calendar-home-set")

		// Step 3: List calendars
		rec = caldavPROPFIND(t, e, "/dav/projects", "1", PropfindCalendarCollectionProperties)
		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)
		assert.GreaterOrEqual(t, len(ms.Responses), 2,
			"Step 3: Should list calendars")

		// Step 4: Check CTag for a specific calendar
		rec = caldavPROPFIND(t, e, "/dav/projects/36", "0", PropfindCalendarCollectionProperties)
		assertResponseStatus(t, rec, 207)

		// Step 5: Full sync — calendar-query to get all task ETags
		rec = caldavREPORT(t, e, "/dav/projects/36", ReportCalendarQuery)
		assertResponseStatus(t, rec, 207)
		ms = parseMultistatus(t, rec)
		assert.Greater(t, len(ms.Responses), 0,
			"Step 5: calendar-query should return tasks")

		// Collect hrefs for multiget
		var hrefs []string
		for _, r := range ms.Responses {
			if strings.HasSuffix(r.Href, ".ics") {
				hrefs = append(hrefs, r.Href)
			}
		}

		// Step 6: Multiget to fetch specific tasks
		if len(hrefs) > 0 {
			body := ReportCalendarMultiget(hrefs[:1]...) // Just fetch first task
			rec = caldavREPORT(t, e, "/dav/projects/36", body)
			assertResponseStatus(t, rec, 207)
			ms = parseMultistatus(t, rec)
			assert.Len(t, ms.Responses, 1,
				"Step 6: multiget should return requested task")
		}

		// Step 7: Push a local change via PUT
		vtodo := NewVTodo("davx5-sync-test", "DAVx5 Synced Task").
			Due(time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)).
			Build()
		rec = caldavPUT(t, e, "/dav/projects/36/davx5-sync-test.ics", vtodo)
		assert.Equal(t, http.StatusCreated, rec.Code,
			"Step 7: PUT should create the task")
	})
}

func TestClientThunderbirdFlow(t *testing.T) {
	t.Run("Thunderbird discovery and initial sync", func(t *testing.T) {
		e := setupTestEnv(t)

		// Step 1: Thunderbird starts with OPTIONS to check DAV support
		rec := caldavOPTIONS(t, e, "/dav/")
		assert.Equal(t, http.StatusOK, rec.Code,
			"Step 1: OPTIONS should succeed")
		davHeader := rec.Header().Get("DAV")
		assert.NotEmpty(t, davHeader,
			"Step 1: Should have DAV header")

		// Step 2: PROPFIND on well-known for principal
		rec = caldavRequest(t, e, "PROPFIND", "/.well-known/caldav", PropfindCurrentUserPrincipal, map[string]string{
			"Depth": "0",
		})
		assert.True(t, rec.Code == 207 || rec.Code == 301 || rec.Code == 302,
			"Step 2: well-known should respond. Got %d", rec.Code)

		// Step 3: PROPFIND principal for calendar-home-set
		rec = caldavPROPFIND(t, e, "/dav/principals/user15/", "0", PropfindCalendarHomeSet)
		assertResponseStatus(t, rec, 207)

		// Step 4: Thunderbird checks current-user-privilege-set to know if it can write
		// RFC 3744 §5.4 (rfc3744.txt line 1158)
		rec = caldavPROPFIND(t, e, "/dav/projects/36", "0", PropfindCurrentUserPrivilegeSet)
		// This may return 207 with or without the property — document the behavior
		assert.True(t, rec.Code == 207 || rec.Code == 200,
			"Step 4: PROPFIND for privileges should not error. Got %d", rec.Code)

		// Step 5: List calendars
		rec = caldavPROPFIND(t, e, "/dav/projects", "1", PropfindCalendarCollectionProperties)
		assertResponseStatus(t, rec, 207)

		// Step 6: Sync via calendar-query
		rec = caldavREPORT(t, e, "/dav/projects/36", ReportCalendarQuery)
		assertResponseStatus(t, rec, 207)
	})
}

func TestClientTasksOrgSubtasks(t *testing.T) {
	t.Run("Tasks.org subtask sync: child-only RELATED-TO", func(t *testing.T) {
		// Tasks.org behavior:
		// - Child tasks include RELATED-TO;RELTYPE=PARENT:<parent-uid>
		// - Parent tasks have NO RELATED-TO at all
		// - Tasks may arrive in any order
		// - On re-sync, parent is sent again without RELATED-TO

		e := setupTestEnv(t)

		// Round 1: Initial sync — parent first, then children
		parent := NewVTodo("tasks-org-parent", "Buy groceries").Build()
		rec := caldavPUT(t, e, "/dav/projects/36/tasks-org-parent.ics", parent)
		require.Equal(t, 201, rec.Code)

		child1 := NewVTodo("tasks-org-child-1", "Buy milk").
			RelatedToParent("tasks-org-parent").Build()
		rec = caldavPUT(t, e, "/dav/projects/36/tasks-org-child-1.ics", child1)
		require.Equal(t, 201, rec.Code)

		child2 := NewVTodo("tasks-org-child-2", "Buy eggs").
			RelatedToParent("tasks-org-parent").Build()
		rec = caldavPUT(t, e, "/dav/projects/36/tasks-org-child-2.ics", child2)
		require.Equal(t, 201, rec.Code)

		// Verify parent shows children
		rec = caldavGET(t, e, "/dav/projects/36/tasks-org-parent.ics")
		body := rec.Body.String()
		assert.Contains(t, body, "tasks-org-child-1")
		assert.Contains(t, body, "tasks-org-child-2")

		// Round 2: Re-sync — parent updated (title change), still no RELATED-TO
		parentUpdated := NewVTodo("tasks-org-parent", "Buy groceries (updated list)").Build()
		rec = caldavPUT(t, e, "/dav/projects/36/tasks-org-parent.ics", parentUpdated)
		require.True(t, rec.Code >= 200 && rec.Code < 300)

		// Verify children are still linked after parent re-sync
		rec = caldavGET(t, e, "/dav/projects/36/tasks-org-parent.ics")
		body = rec.Body.String()
		assert.Contains(t, body, "Buy groceries (updated list)",
			"Parent title should be updated")
		assert.Contains(t, body, "tasks-org-child-1",
			"Child 1 relation should survive parent re-sync")
		assert.Contains(t, body, "tasks-org-child-2",
			"Child 2 relation should survive parent re-sync")

		// Round 3: Complete child via PUT with STATUS:COMPLETED
		child1Done := NewVTodo("tasks-org-child-1", "Buy milk").
			RelatedToParent("tasks-org-parent").
			Status("COMPLETED").
			Completed(time.Now().UTC()).
			Build()
		rec = caldavPUT(t, e, "/dav/projects/36/tasks-org-child-1.ics", child1Done)
		require.True(t, rec.Code >= 200 && rec.Code < 300)

		// Verify child is completed
		rec = caldavGET(t, e, "/dav/projects/36/tasks-org-child-1.ics")
		assert.Contains(t, rec.Body.String(), "STATUS:COMPLETED")
	})

	t.Run("Tasks.org subtask sync: children arrive before parent", func(t *testing.T) {
		e := setupTestEnv(t)

		// Children arrive first (reverse order)
		child := NewVTodo("tasks-rev-child", "Subtask").
			RelatedToParent("tasks-rev-parent").Build()
		rec := caldavPUT(t, e, "/dav/projects/36/tasks-rev-child.ics", child)
		require.Equal(t, 201, rec.Code)

		// Parent arrives later — no RELATED-TO
		parent := NewVTodo("tasks-rev-parent", "Main Task").Build()
		rec = caldavPUT(t, e, "/dav/projects/36/tasks-rev-parent.ics", parent)
		require.Equal(t, 201, rec.Code)

		// Verify bidirectional relations
		rec = caldavGET(t, e, "/dav/projects/36/tasks-rev-parent.ics")
		assert.Contains(t, rec.Body.String(), "SUMMARY:Main Task",
			"Parent should have real title, not DUMMY")
		assert.Contains(t, rec.Body.String(), "tasks-rev-child",
			"Parent should show child relation")

		rec = caldavGET(t, e, "/dav/projects/36/tasks-rev-child.ics")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=PARENT:tasks-rev-parent")
	})
}
