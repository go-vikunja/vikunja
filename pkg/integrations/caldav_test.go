// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package integrations

import (
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/routes/caldav"
	"github.com/stretchr/testify/assert"
)

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

func TestCaldav(t *testing.T) {
	t.Run("Delivers VTODO for project", func(t *testing.T) {
		rec, err := newCaldavTestRequestWithUser(t, http.MethodGet, caldav.ProjectHandler, &testuser15, ``, nil, map[string]string{"project": "36"})
		assert.NoError(t, err)
		assert.Contains(t, rec.Body.String(), "BEGIN:VCALENDAR")
		assert.Contains(t, rec.Body.String(), "PRODID:-//Vikunja Todo App//EN")
		assert.Contains(t, rec.Body.String(), "X-WR-CALNAME:Project 36 for Caldav tests")
		assert.Contains(t, rec.Body.String(), "BEGIN:VTODO")
		assert.Contains(t, rec.Body.String(), "END:VTODO")
		assert.Contains(t, rec.Body.String(), "END:VCALENDAR")
	})
	t.Run("Import VTODO", func(t *testing.T) {
		rec, err := newCaldavTestRequestWithUser(t, http.MethodPut, caldav.TaskHandler, &testuser15, vtodo, nil, map[string]string{"project": "36", "task": "uid"})
		assert.NoError(t, err)
		assert.Equal(t, 201, rec.Result().StatusCode)
	})
	t.Run("Export VTODO", func(t *testing.T) {
		rec, err := newCaldavTestRequestWithUser(t, http.MethodGet, caldav.TaskHandler, &testuser15, ``, nil, map[string]string{"project": "36", "task": "uid-caldav-test"})
		assert.NoError(t, err)
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
