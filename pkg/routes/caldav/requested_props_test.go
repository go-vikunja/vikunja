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

package caldav

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceContentIsRenderedOnce(t *testing.T) {
	task := &models.Task{UID: "uid-rendered-once", Title: "first"}
	rr := &VikunjaProjectResourceAdapter{task: task}

	content := rr.GetContent()
	require.Contains(t, content, "SUMMARY:first")

	// A second render would pick up the new title.
	task.Title = "second"
	assert.Equal(t, content, rr.GetContent())
	assert.Equal(t, int64(len(content)), rr.GetContentSize())
}

func propfindWithProps(props ...string) string {
	return `<?xml version="1.0" encoding="utf-8" ?>
<D:propfind xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
  <D:prop>` + strings.Join(props, "") + `</D:prop>
</D:propfind>`
}

func multigetWithProps(props ...string) string {
	return multiget(props, "/dav/projects/36/uid-caldav-test.ics")
}

func multiget(props []string, hrefs ...string) string {
	return `<?xml version="1.0" encoding="utf-8" ?>
<C:calendar-multiget xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
  <D:prop>` + strings.Join(props, "") + `</D:prop>
  <D:href>` + strings.Join(hrefs, "</D:href><D:href>") + `</D:href>
</C:calendar-multiget>`
}

func TestRequestedProperties(t *testing.T) {
	sameAnswer := func(t *testing.T, handler func(*echo.Context) error, method, path, repeated, single string, params echo.PathValues) {
		db.LoadAndAssertFixtures(t)
		want, err := callCaldavHandler(t, handler, caldavFilterUser, method, path, single, params)
		requireHTTPStatus(t, http.StatusMultiStatus, want, err)
		got, err := callCaldavHandler(t, handler, caldavFilterUser, method, path, repeated, params)
		requireHTTPStatus(t, http.StatusMultiStatus, got, err)
		assert.Equal(t, want.Body.String(), got.Body.String())
	}

	t.Run("PROPFIND answers a property listed twice once", func(t *testing.T) {
		sameAnswer(t, ProjectHandler, "PROPFIND", projectPath(36),
			propfindWithProps(`<D:getetag/>`, `<C:calendar-data/>`, `<C:calendar-data/>`),
			propfindWithProps(`<D:getetag/>`, `<C:calendar-data/>`),
			projectParams(36))
	})

	t.Run("task PROPFIND answers a property listed twice once", func(t *testing.T) {
		sameAnswer(t, TaskHandler, "PROPFIND", projectPath(36)+"uid-caldav-test.ics",
			propfindWithProps(`<C:calendar-data/>`, `<C:calendar-data/>`),
			propfindWithProps(`<C:calendar-data/>`),
			taskParams(36, "uid-caldav-test"))
	})

	t.Run("REPORT answers a property listed twice once", func(t *testing.T) {
		sameAnswer(t, ProjectHandler, "REPORT", projectPath(36),
			multigetWithProps(`<C:calendar-data/>`, `<C:calendar-data/>`),
			multigetWithProps(`<C:calendar-data/>`),
			projectParams(36))
	})

	t.Run("REPORT answers an href listed twice once", func(t *testing.T) {
		href := "/dav/projects/36/uid-caldav-test.ics"
		props := []string{`<D:getetag/>`, `<C:calendar-data/>`}
		sameAnswer(t, ProjectHandler, "REPORT", projectPath(36),
			multiget(props, href, href),
			multiget(props, href),
			projectParams(36))
	})

	t.Run("REPORT with distinct hrefs still works", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		body := multiget([]string{`<D:getetag/>`, `<C:calendar-data/>`}, "/dav/projects/36/uid-caldav-test.ics", "/dav/projects/36/does-not-exist.ics")
		rec, err := callCaldavHandler(t, ProjectHandler, caldavFilterUser, "REPORT", projectPath(36), body, projectParams(36))
		requireHTTPStatus(t, http.StatusMultiStatus, rec, err)
		assert.Equal(t, 1, strings.Count(rec.Body.String(), "BEGIN:VTODO"))
	})

	t.Run("PROPFIND with too many properties is rejected", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		// caldav-go caps requests at 200 properties.
		props := make([]string, 0, 201)
		for i := range 201 {
			props = append(props, `<X:p`+strconv.Itoa(i)+` xmlns:X="urn:example"/>`)
		}
		rec, err := callCaldavHandler(t, PrincipalHandler, caldavFilterUser, "PROPFIND", principalPathForUser(caldavFilterUser.Username), propfindWithProps(props...), nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("regular client PROPFIND still works", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		body := propfindWithProps(`<D:resourcetype/>`, `<D:displayname/>`, `<D:getetag/>`, `<C:calendar-data/>`, `<D:getcontentlength/>`)
		rec, err := callCaldavHandler(t, ProjectHandler, caldavFilterUser, "PROPFIND", projectPath(36), body, projectParams(36))
		requireHTTPStatus(t, http.StatusMultiStatus, rec, err)
		assert.Contains(t, rec.Body.String(), "BEGIN:VTODO")
	})
}
