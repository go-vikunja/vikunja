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
	"encoding/xml"
	"net/http"
	"testing"

	"code.vikunja.io/api/pkg/db"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

const colorProp = `<I:calendar-color xmlns:I="http://apple.com/ns/ical/"/>`

func TestPropfindCalendarColor(t *testing.T) {
	const ok = "HTTP/1.1 200 OK"
	const notFound = "HTTP/1.1 404 Not Found"

	propfind := func(t *testing.T, path string, params echo.PathValues) (map[xml.Name]string, map[xml.Name]string) {
		t.Helper()

		rec, err := callCaldavHandler(t, ProjectHandler, caldavFilterUser, "PROPFIND", path, propfindWithProps(colorProp), params)
		requireHTTPStatus(t, http.StatusMultiStatus, rec, err)
		return parseMultistatus(t, rec).props(t, collectionHref(36))
	}

	t.Run("a project with a color answers it", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setStoredColor(t, "007AFF")

		statuses, values := propfind(t, projectPath(36), projectParams(36))
		assert.Equal(t, map[xml.Name]string{appleCalendarColor: ok}, statuses)
		assert.Equal(t, "#007AFFFF", values[appleCalendarColor])
	})

	t.Run("a project without a color answers 404 for it", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		statuses, _ := propfind(t, projectPath(36), projectParams(36))
		assert.Equal(t, map[xml.Name]string{appleCalendarColor: notFound}, statuses)
	})

	t.Run("iOS reads back the color it set", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		proppatch(t, caldavFilterUser, 36, iosSetColor)
		_, values := propfind(t, ProjectHomeSetPath, nil)
		assert.Equal(t, "#007AFFFF", values[appleCalendarColor])
	})
}
