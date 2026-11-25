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
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	appleCalendarColor = xml.Name{Space: "http://apple.com/ns/ical/", Local: "calendar-color"}
	appleCalendarOrder = xml.Name{Space: "http://apple.com/ns/ical/", Local: "calendar-order"}
)

// What iOS Reminders sends for every list on launch.
const iosSetColor = `<A:propertyupdate xmlns:A="DAV:"><A:set><A:prop>` +
	`<D:calendar-color xmlns:D="http://apple.com/ns/ical/" symbolic-color="custom">#007AFF</D:calendar-color>` +
	`</A:prop></A:set></A:propertyupdate>`

const removeColor = `<A:propertyupdate xmlns:A="DAV:"><A:remove><A:prop>` +
	`<D:calendar-color xmlns:D="http://apple.com/ns/ical/"/>` +
	`</A:prop></A:remove></A:propertyupdate>`

func setColorBody(value string) string {
	return `<A:propertyupdate xmlns:A="DAV:"><A:set><A:prop>` +
		`<D:calendar-color xmlns:D="http://apple.com/ns/ical/">` + value + `</D:calendar-color>` +
		`</A:prop></A:set></A:propertyupdate>`
}

type testProp struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

type testPropstat struct {
	Prop struct {
		Props []testProp `xml:",any"`
	} `xml:"DAV: prop"`
	Status string `xml:"DAV: status"`
}

type testResponse struct {
	Href      string         `xml:"DAV: href"`
	Propstats []testPropstat `xml:"DAV: propstat"`
}

type testMultistatus struct {
	Responses []testResponse `xml:"DAV: response"`
}

func parseMultistatus(t *testing.T, rec *httptest.ResponseRecorder) testMultistatus {
	t.Helper()

	var ms testMultistatus
	require.NoError(t, xml.Unmarshal(rec.Body.Bytes(), &ms), rec.Body.String())
	return ms
}

func (ms testMultistatus) props(t *testing.T, href string) (statuses map[xml.Name]string, values map[xml.Name]string) {
	t.Helper()

	statuses = map[xml.Name]string{}
	values = map[xml.Name]string{}
	found := false
	for _, r := range ms.Responses {
		if r.Href != href {
			continue
		}
		found = true
		for _, ps := range r.Propstats {
			for _, p := range ps.Prop.Props {
				_, twice := statuses[p.XMLName]
				require.False(t, twice, "%v is answered twice", p.XMLName)
				statuses[p.XMLName] = ps.Status
				values[p.XMLName] = p.Value
			}
		}
	}
	require.True(t, found, "no response for %s", href)
	return statuses, values
}

func proppatch(t *testing.T, u *user.User, projectID int64, body string) map[xml.Name]string {
	t.Helper()

	rec, err := callCaldavHandler(t, ProjectHandler, u, "PROPPATCH", projectPath(projectID), body, projectParams(projectID))
	requireHTTPStatus(t, http.StatusMultiStatus, rec, err)
	statuses, _ := parseMultistatus(t, rec).props(t, projectPath(projectID))
	return statuses
}

func storedProject(t *testing.T, projectID int64) *models.Project {
	t.Helper()

	s := db.NewSession()
	defer s.Close()
	p, err := models.GetProjectSimpleByID(s, projectID)
	require.NoError(t, err)
	return p
}

func setStoredColor(t *testing.T, color string) {
	t.Helper()

	s := db.NewSession()
	defer s.Close()
	_, err := s.ID(36).Cols("hex_color").NoAutoTime().Update(&models.Project{HexColor: color})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

func TestNormalizeCalendarColor(t *testing.T) {
	for _, tc := range []struct{ value, want string }{
		{"#007AFFCC", "007AFF"},
		{"  #007AFF\n ", "007AFF"},
	} {
		got, ok := normalizeCalendarColor(tc.value)
		assert.True(t, ok, tc.value)
		assert.Equal(t, tc.want, got, tc.value)
	}

	for _, value := range []string{"#FFF", "#007AFFzz"} {
		_, ok := normalizeCalendarColor(value)
		assert.False(t, ok, value)
	}
}

func TestPropPatchCalendarColor(t *testing.T) {
	const ok = "HTTP/1.1 200 OK"
	const forbidden = "HTTP/1.1 403 Forbidden"
	const failedDependency = "HTTP/1.1 424 Failed Dependency"

	t.Run("the owner sets a color", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		statuses := proppatch(t, caldavFilterUser, 36, iosSetColor)
		assert.Equal(t, map[xml.Name]string{appleCalendarColor: ok}, statuses)
		assert.Equal(t, "007AFF", storedProject(t, 36).HexColor)
	})

	t.Run("remove clears the color", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setStoredColor(t, "007AFF")

		statuses := proppatch(t, caldavFilterUser, 36, removeColor)
		assert.Equal(t, map[xml.Name]string{appleCalendarColor: ok}, statuses)
		assert.Empty(t, storedProject(t, 36).HexColor)
	})

	t.Run("an unchanged color is not written again", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setStoredColor(t, "007AFF")
		before := storedProject(t, 36).Updated

		statuses := proppatch(t, caldavFilterUser, 36, iosSetColor)
		assert.Equal(t, map[xml.Name]string{appleCalendarColor: ok}, statuses)
		after := storedProject(t, 36).Updated
		assert.True(t, before.Equal(after), "updated changed from %v to %v", before, after)
	})

	t.Run("a color that differs only in case is written", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setStoredColor(t, "007aff")

		statuses := proppatch(t, caldavFilterUser, 36, iosSetColor)
		assert.Equal(t, map[xml.Name]string{appleCalendarColor: ok}, statuses)
		assert.Equal(t, "007AFF", storedProject(t, 36).HexColor)
	})

	t.Run("the last instruction for a property wins", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setStoredColor(t, "112233")

		body := `<A:propertyupdate xmlns:A="DAV:" xmlns:D="http://apple.com/ns/ical/">` +
			`<A:remove><A:prop><D:calendar-color/></A:prop></A:remove>` +
			`<A:set><A:prop><D:calendar-color>#007AFF</D:calendar-color></A:prop></A:set>` +
			`</A:propertyupdate>`
		statuses := proppatch(t, caldavFilterUser, 36, body)
		assert.Equal(t, map[xml.Name]string{appleCalendarColor: ok}, statuses)
		assert.Equal(t, "007AFF", storedProject(t, 36).HexColor)
	})

	t.Run("an invalid color is refused", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setStoredColor(t, "112233")

		statuses := proppatch(t, caldavFilterUser, 36, setColorBody("not a color"))
		assert.Equal(t, map[xml.Name]string{appleCalendarColor: forbidden}, statuses)
		assert.Equal(t, "112233", storedProject(t, 36).HexColor)
	})

	t.Run("nothing but the color changes", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		s := db.NewSession()
		fav := &models.Project{ID: 36}
		require.NoError(t, fav.ReadOne(s, caldavFilterUser))
		before := storedProject(t, 36)
		fav.Title, fav.Description, fav.Identifier, fav.Position = before.Title, before.Description, before.Identifier, before.Position
		fav.IsFavorite = true
		require.NoError(t, fav.Update(s, caldavFilterUser))
		require.NoError(t, s.Commit())
		s.Close()

		proppatch(t, caldavFilterUser, 36, iosSetColor)

		after := storedProject(t, 36)
		assert.Equal(t, "007AFF", after.HexColor)
		assert.Equal(t, before.Title, after.Title)
		assert.Equal(t, before.Description, after.Description)
		assert.Equal(t, before.Identifier, after.Identifier)
		assert.Equal(t, before.ParentProjectID, after.ParentProjectID)
		assert.InDelta(t, before.Position, after.Position, 0)
		assert.False(t, after.IsArchived)

		s = db.NewSession()
		defer s.Close()
		read := &models.Project{ID: 36}
		require.NoError(t, read.ReadOne(s, caldavFilterUser))
		assert.True(t, read.IsFavorite, "the project is no longer a favorite")
	})

	t.Run("a read-only share cannot set a color", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		// Project 3 is owned by user 3 and shared read-only with user 1.
		statuses := proppatch(t, caldavOtherUser, 3, iosSetColor)
		assert.Equal(t, map[xml.Name]string{appleCalendarColor: forbidden}, statuses)
		assert.Empty(t, storedProject(t, 3).HexColor)
	})

	t.Run("an archived project keeps its color", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setStoredColor(t, "112233")
		archiveProject(t, 36)

		statuses := proppatch(t, caldavFilterUser, 36, iosSetColor)
		assert.Equal(t, map[xml.Name]string{appleCalendarColor: forbidden}, statuses)
		stored := storedProject(t, 36)
		assert.Equal(t, "112233", stored.HexColor)
		assert.True(t, stored.IsArchived)
	})

	t.Run("a refused property keeps the color from being applied", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		body := `<A:propertyupdate xmlns:A="DAV:" xmlns:D="http://apple.com/ns/ical/"><A:set><A:prop>` +
			`<D:calendar-color>#007AFF</D:calendar-color>` +
			`<D:calendar-order>3</D:calendar-order>` +
			`</A:prop></A:set></A:propertyupdate>`
		statuses := proppatch(t, caldavFilterUser, 36, body)
		assert.Equal(t, map[xml.Name]string{
			appleCalendarColor: failedDependency,
			appleCalendarOrder: forbidden,
		}, statuses)
		assert.Empty(t, storedProject(t, 36).HexColor)
	})

	t.Run("a saved filter is not writable", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)
		statuses := proppatch(t, caldavFilterUser, pseudoID, iosSetColor)
		assert.Equal(t, map[xml.Name]string{appleCalendarColor: forbidden}, statuses)

		s := db.NewSession()
		defer s.Close()
		sf, err := models.GetSavedFilterSimpleByID(s, models.GetSavedFilterIDFromProjectID(pseudoID))
		require.NoError(t, err)
		assert.Equal(t, "caldav test filter", sf.Title)
	})

	t.Run("a task resource has no color to set", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		path := projectPath(36) + "uid-caldav-test.ics"
		rec, err := callCaldavHandler(t, TaskHandler, caldavFilterUser, "PROPPATCH", path, iosSetColor, taskParams(36, "uid-caldav-test"))
		requireHTTPStatus(t, http.StatusMultiStatus, rec, err)
		statuses, _ := parseMultistatus(t, rec).props(t, path)
		assert.Equal(t, map[xml.Name]string{appleCalendarColor: forbidden}, statuses)
		assert.Empty(t, storedProject(t, 36).HexColor)
	})
}
