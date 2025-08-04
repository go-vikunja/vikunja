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

package webtests

import (
	"encoding/xml"
	"net/http"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/routes/caldav"

	ics "github.com/arran4/golang-ical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCaldav(t *testing.T) {
	t.Run("Delivers VTODO for project", func(t *testing.T) {
		e, _ := setupTestEnv()
		rec, err := newCaldavTestRequestWithUser(t, e, http.MethodGet, caldav.ProjectHandler, &testuser15, ``, nil, map[string]string{"project": "36"})
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), "BEGIN:VCALENDAR")
		assert.Contains(t, rec.Body.String(), "PRODID:-//Vikunja Todo App//EN")
		assert.Contains(t, rec.Body.String(), "X-WR-CALNAME:Project 36 for Caldav tests")
		assert.Contains(t, rec.Body.String(), "BEGIN:VTODO")
		assert.Contains(t, rec.Body.String(), "END:VTODO")
		assert.Contains(t, rec.Body.String(), "END:VCALENDAR")
	})
	t.Run("Import VTODO", func(t *testing.T) {
		const vtodo = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:List 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid
DTSTAMP:20230301T073337Z
SUMMARY:Caldav Task 1
CATEGORIES:tag1,tag2,tag3
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
BEGIN:VALARM
TRIGGER;VALUE=DATE-TIME:20230304T150000Z
ACTION:DISPLAY
END:VALARM
END:VTODO
END:VCALENDAR`

		e, _ := setupTestEnv()
		rec, err := newCaldavTestRequestWithUser(t, e, http.MethodPut, caldav.TaskHandler, &testuser15, vtodo, nil, map[string]string{"project": "36", "task": "uid"})
		require.NoError(t, err)
		assert.Equal(t, 201, rec.Result().StatusCode)
	})
	t.Run("Export VTODO", func(t *testing.T) {
		e, _ := setupTestEnv()
		rec, err := newCaldavTestRequestWithUser(t, e, http.MethodGet, caldav.TaskHandler, &testuser15, ``, nil, map[string]string{"project": "36", "task": "uid-caldav-test"})
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), "BEGIN:VCALENDAR")
		assert.Contains(t, rec.Body.String(), "SUMMARY:Title Caldav Test")
		assert.Contains(t, rec.Body.String(), "DESCRIPTION:Description Caldav Test")
		assert.Contains(t, rec.Body.String(), "DUE:20230301T150000Z")
		assert.Contains(t, rec.Body.String(), "PRIORITY:3")
		assert.Contains(t, rec.Body.String(), "CATEGORIES:Label #4")
		assert.Contains(t, rec.Body.String(), "BEGIN:VALARM")
		assert.Contains(t, rec.Body.String(), "TRIGGER;VALUE=DATE-TIME:20230304T150000Z")
		assert.Contains(t, rec.Body.String(), "ACTION:DISPLAY")
		assert.Contains(t, rec.Body.String(), "END:VALARM")
	})
}

func TestCaldavSubtasks(t *testing.T) {
	const vtodoHeader = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
`
	const vtodoFooter = `
END:VCALENDAR`

	t.Run("Import Task & Subtask", func(t *testing.T) {

		const vtodoParentTaskStub = `BEGIN:VTODO
UID:uid_parent_import
DTSTAMP:20230301T073337Z
SUMMARY:Caldav parent task
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
END:VTODO`

		const vtodoChildTaskStub = `BEGIN:VTODO
UID:uid_child_import
DTSTAMP:20230301T073337Z
SUMMARY:Caldav child task
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=PARENT:uid_parent_import
END:VTODO`

		const vtodoGrandChildTaskStub = `
BEGIN:VTODO
UID:uid_grand_child_import
DTSTAMP:20230301T073337Z
SUMMARY:Caldav grand child task
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=PARENT:uid_child_import
END:VTODO`

		e, _ := setupTestEnv()

		const parentVTODO = vtodoHeader + vtodoParentTaskStub + vtodoFooter
		rec, err := newCaldavTestRequestWithUser(t, e, http.MethodPut, caldav.TaskHandler, &testuser15, parentVTODO, nil, map[string]string{"project": "36", "task": "uid_parent_import"})
		require.NoError(t, err)
		assert.Equal(t, 201, rec.Result().StatusCode)

		const childVTODO = vtodoHeader + vtodoChildTaskStub + vtodoFooter
		rec, err = newCaldavTestRequestWithUser(t, e, http.MethodPut, caldav.TaskHandler, &testuser15, childVTODO, nil, map[string]string{"project": "36", "task": "uid_child_import"})
		require.NoError(t, err)
		assert.Equal(t, 201, rec.Result().StatusCode)

		const grandChildVTODO = vtodoHeader + vtodoGrandChildTaskStub + vtodoFooter
		rec, err = newCaldavTestRequestWithUser(t, e, http.MethodPut, caldav.TaskHandler, &testuser15, grandChildVTODO, nil, map[string]string{"project": "36", "task": "uid_grand_child_import"})
		require.NoError(t, err)
		assert.Equal(t, 201, rec.Result().StatusCode)

		rec, err = newCaldavTestRequestWithUser(t, e, http.MethodGet, caldav.ProjectHandler, &testuser15, ``, nil, map[string]string{"project": "36"})
		require.NoError(t, err)
		assert.Equal(t, 200, rec.Result().StatusCode)

		assert.Contains(t, rec.Body.String(), "UID:uid_parent_import")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=CHILD:uid_child_import")
		assert.Contains(t, rec.Body.String(), "UID:uid_child_import")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=PARENT:uid_parent_import")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=CHILD:uid_grand_child_import")
		assert.Contains(t, rec.Body.String(), "UID:uid_grand_child_import")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=PARENT:uid_child_import")
	})

	t.Run("Import Task & Subtask (Reverse - Subtask first)", func(t *testing.T) {
		e, _ := setupTestEnv()

		const vtodoGrandChildTaskStub = `
BEGIN:VTODO
UID:uid_grand_child_import
DTSTAMP:20230301T073337Z
SUMMARY:Caldav grand child task
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=PARENT:uid_child_import
END:VTODO`

		const grandChildVTODO = vtodoHeader + vtodoGrandChildTaskStub + vtodoFooter
		rec, err := newCaldavTestRequestWithUser(t, e, http.MethodPut, caldav.TaskHandler, &testuser15, grandChildVTODO, nil, map[string]string{"project": "36", "task": "uid_grand_child_import"})
		require.NoError(t, err)
		assert.Equal(t, 201, rec.Result().StatusCode)

		const vtodoChildTaskStub = `BEGIN:VTODO
UID:uid_child_import
DTSTAMP:20230301T073337Z
SUMMARY:Caldav child task
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=PARENT:uid_parent_import
RELATED-TO;RELTYPE=CHILD:uid_grand_child_import
END:VTODO`

		const childVTODO = vtodoHeader + vtodoChildTaskStub + vtodoFooter
		rec, err = newCaldavTestRequestWithUser(t, e, http.MethodPut, caldav.TaskHandler, &testuser15, childVTODO, nil, map[string]string{"project": "36", "task": "uid_child_import"})
		require.NoError(t, err)
		assert.Equal(t, 201, rec.Result().StatusCode)

		const vtodoParentTaskStub = `BEGIN:VTODO
UID:uid_parent_import
DTSTAMP:20230301T073337Z
SUMMARY:Caldav parent task
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=CHILD:uid_child_import
END:VTODO`

		const parentVTODO = vtodoHeader + vtodoParentTaskStub + vtodoFooter
		rec, err = newCaldavTestRequestWithUser(t, e, http.MethodPut, caldav.TaskHandler, &testuser15, parentVTODO, nil, map[string]string{"project": "36", "task": "uid_parent_import"})
		require.NoError(t, err)
		assert.Equal(t, 201, rec.Result().StatusCode)

		rec, err = newCaldavTestRequestWithUser(t, e, http.MethodGet, caldav.ProjectHandler, &testuser15, ``, nil, map[string]string{"project": "36"})
		require.NoError(t, err)
		assert.Equal(t, 200, rec.Result().StatusCode)

		assert.Contains(t, rec.Body.String(), "UID:uid_parent_import")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=CHILD:uid_child_import")
		assert.Contains(t, rec.Body.String(), "UID:uid_child_import")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=PARENT:uid_parent_import")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=CHILD:uid_grand_child_import")
		assert.Contains(t, rec.Body.String(), "UID:uid_grand_child_import")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=PARENT:uid_child_import")
	})

	t.Run("Delete Subtask", func(t *testing.T) {
		e, _ := setupTestEnv()

		rec, err := newCaldavTestRequestWithUser(t, e, http.MethodDelete, caldav.TaskHandler, &testuser15, ``, nil, map[string]string{"project": "36", "task": "uid-caldav-test-child-task"})
		require.NoError(t, err)
		assert.Equal(t, 204, rec.Result().StatusCode)

		rec, err = newCaldavTestRequestWithUser(t, e, http.MethodDelete, caldav.TaskHandler, &testuser15, ``, nil, map[string]string{"project": "36", "task": "uid-caldav-test-child-task-2"})
		require.NoError(t, err)
		assert.Equal(t, 204, rec.Result().StatusCode)

		rec, err = newCaldavTestRequestWithUser(t, e, http.MethodGet, caldav.TaskHandler, &testuser15, ``, nil, map[string]string{"project": "36", "task": "uid-caldav-test-parent-task"})
		require.NoError(t, err)
		assert.Equal(t, 200, rec.Result().StatusCode)

		assert.NotContains(t, rec.Body.String(), "RELATED-TO;RELTYPE=CHILD:uid-caldav-test-child-task")
		assert.NotContains(t, rec.Body.String(), "RELATED-TO;RELTYPE=CHILD:uid-caldav-test-child-task-2")
	})

	t.Run("Delete Parent Task", func(t *testing.T) {
		e, _ := setupTestEnv()

		rec, err := newCaldavTestRequestWithUser(t, e, http.MethodDelete, caldav.TaskHandler, &testuser15, ``, nil, map[string]string{"project": "36", "task": "uid-caldav-test-parent-task"})
		require.NoError(t, err)
		assert.Equal(t, 204, rec.Result().StatusCode)

		rec, err = newCaldavTestRequestWithUser(t, e, http.MethodGet, caldav.TaskHandler, &testuser15, ``, nil, map[string]string{"project": "36", "task": "uid-caldav-test-child-task"})
		require.NoError(t, err)
		assert.Equal(t, 200, rec.Result().StatusCode)

		assert.NotContains(t, rec.Body.String(), "RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-task")
	})

}

func TestCaldavSubtasksDifferentLists(t *testing.T) {
	t.Run("Import Parent Task & Child Task Different Lists", func(t *testing.T) {
		const vtodoParentTask = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 36 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid_parent_import
DTSTAMP:20230301T073337Z
SUMMARY:Caldav parent task
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
END:VTODO
END:VCALENDAR`

		const vtodoChildTask = `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:Project 38 for Caldav tests
PRODID:-//Vikunja Todo App//EN
BEGIN:VTODO
UID:uid_child_import
DTSTAMP:20230301T073337Z
SUMMARY:Caldav child task
CREATED:20230301T073337Z
LAST-MODIFIED:20230301T073337Z
RELATED-TO;RELTYPE=PARENT:uid_parent_import
END:VTODO
END:VCALENDAR`

		e, _ := setupTestEnv()

		rec, err := newCaldavTestRequestWithUser(t, e, http.MethodPut, caldav.TaskHandler, &testuser15, vtodoParentTask, nil, map[string]string{"project": "36", "task": "uid_parent_import"})
		require.NoError(t, err)
		assert.Equal(t, 201, rec.Result().StatusCode)

		rec, err = newCaldavTestRequestWithUser(t, e, http.MethodPut, caldav.TaskHandler, &testuser15, vtodoChildTask, nil, map[string]string{"project": "38", "task": "uid_child_import"})
		require.NoError(t, err)
		assert.Equal(t, 201, rec.Result().StatusCode)

		rec, err = newCaldavTestRequestWithUser(t, e, http.MethodGet, caldav.TaskHandler, &testuser15, ``, nil, map[string]string{"project": "36", "task": "uid_parent_import"})
		require.NoError(t, err)
		assert.Equal(t, 200, rec.Result().StatusCode)
		assert.Contains(t, rec.Body.String(), "UID:uid_parent_import")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=CHILD:uid_child_import")

		rec, err = newCaldavTestRequestWithUser(t, e, http.MethodGet, caldav.TaskHandler, &testuser15, ``, nil, map[string]string{"project": "38", "task": "uid_child_import"})
		require.NoError(t, err)
		assert.Equal(t, 200, rec.Result().StatusCode)
		assert.Contains(t, rec.Body.String(), "UID:uid_child_import")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=PARENT:uid_parent_import")
	})

	t.Run("Check relationships across lists", func(t *testing.T) {
		e, _ := setupTestEnv()

		rec, err := newCaldavTestRequestWithUser(t, e, http.MethodGet, caldav.TaskHandler, &testuser15, ``, nil, map[string]string{"project": "36", "task": "uid-caldav-test-parent-task-another-list"})
		require.NoError(t, err)
		assert.Equal(t, 200, rec.Result().StatusCode)
		assert.Contains(t, rec.Body.String(), "UID:uid-caldav-test-parent-task-another-list")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=CHILD:uid-caldav-test-child-task-another-list")

		rec, err = newCaldavTestRequestWithUser(t, e, http.MethodGet, caldav.TaskHandler, &testuser15, ``, nil, map[string]string{"project": "38", "task": "uid-caldav-test-child-task-another-list"})
		require.NoError(t, err)
		assert.Equal(t, 200, rec.Result().StatusCode)
		assert.Contains(t, rec.Body.String(), "UID:uid-caldav-test-child-task-another-list")
		assert.Contains(t, rec.Body.String(), "RELATED-TO;RELTYPE=PARENT:uid-caldav-test-parent-task-another-list")
	})
}

func TestCaldavProjectReport(t *testing.T) {
	t.Run("REPORT calendar-query returns all tasks", func(t *testing.T) {
		e, _ := setupTestEnv()

		// CalDAV REPORT request for calendar-query
		reportBody := `<?xml version="1.0" encoding="utf-8" ?>
<C:calendar-query xmlns:C="urn:ietf:params:xml:ns:caldav">
    <D:prop xmlns:D="DAV:">
        <D:getetag/>
        <C:calendar-data/>
    </D:prop>
    <C:filter>
        <C:comp-filter name="VCALENDAR">
            <C:comp-filter name="VTODO"/>
        </C:comp-filter>
    </C:filter>
</C:calendar-query>`

		rec, err := newCaldavTestRequestWithUser(t, e, "REPORT", caldav.ProjectHandler, &testuser15, reportBody, nil, map[string]string{"project": "36"})
		require.NoError(t, err)
		assert.Equal(t, 207, rec.Result().StatusCode) // Multi-Status response

		responseBody := rec.Body.String()

		assert.Contains(t, responseBody, "multistatus")
		assert.Contains(t, responseBody, "response")
		assert.Contains(t, responseBody, "href")
		assert.Contains(t, responseBody, "propstat")

		// Parse XML to verify structure
		type Multistatus struct {
			Response []struct {
				Href     string `xml:"href"`
				Propstat struct {
					Prop struct {
						Getetag      string `xml:"getetag"`
						CalendarData string `xml:"calendar-data"`
					} `xml:"prop"`
				} `xml:"propstat"`
			} `xml:"response"`
		}

		var multistatus Multistatus
		err = xml.Unmarshal([]byte(responseBody), &multistatus)
		require.NoError(t, err)

		assert.Len(t, multistatus.Response, 5, "Should have all tasks from the project")

		for i, response := range multistatus.Response {
			assert.NotEmpty(t, response.Href, "Response %d should have an href", i)
			assert.NotEmpty(t, response.Propstat.Prop.CalendarData, "Response %d should have calendar-data", i)
			assert.NotEmpty(t, response.Propstat.Prop.Getetag, "Response %d should have an ETag", i)

			calendarData := response.Propstat.Prop.CalendarData

			cal, err := ics.ParseCalendar(strings.NewReader(calendarData))
			require.NoError(t, err, "Response %d should contain valid iCalendar data", i)

			require.Len(t, cal.Components, 1, "Response %d should contain exactly one VTODO component")

			component := cal.Components[0]
			require.IsType(t, &ics.VTodo{}, component)

			vtodo, _ := component.(*ics.VTodo)
			uid := vtodo.GetProperty(ics.ComponentPropertyUniqueId)
			assert.NotEmpty(t, uid.Value, "Response %d VTODO UID should not be empty", i)

			summary := vtodo.GetProperty(ics.ComponentPropertySummary)
			assert.NotEmpty(t, summary.Value, "Response %d VTODO SUMMARY should not be empty", i)
		}
	})
}
