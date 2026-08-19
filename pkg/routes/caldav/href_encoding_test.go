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

// Client-chosen task UIDs end up in CalDAV hrefs, so they must not be able to
// forge a path or inject XML markup.

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/samedi/caldav-go/ixml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

// A UID that tries both path forgery and XML injection at once.
const hostileUID = `evil/../../../projects/5/pwned<x>&y`

func caldavTestUser() *user.User {
	return &user.User{ID: 15, Username: "user15", Email: "user15@example.com"}
}

// insertHostileTask puts a task with a client-chosen UID into project 36.
func insertHostileTask(t *testing.T, uid string) *models.Task {
	t.Helper()

	s := db.NewSession()
	defer s.Close()

	task := &models.Task{
		Title:       "Hostile UID task",
		UID:         uid,
		ProjectID:   36,
		Index:       98,
		CreatedByID: 15,
	}
	_, err := s.Insert(task)
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	return task
}

func countTasksWithUID(t *testing.T, uid string) int64 {
	t.Helper()

	s := db.NewSession()
	defer s.Close()

	count, err := s.Where(builder.Eq{"uid": uid}).Count(&models.Task{})
	require.NoError(t, err)
	return count
}

func TestTaskURL_UIDEncoding(t *testing.T) {
	t.Run("uuid uids are encoded byte for byte identically", func(t *testing.T) {
		const uid = "550e8400-e29b-41d4-a716-446655440000"

		href := taskURL(36, &models.Task{UID: uid})

		assert.Equal(t, "/dav/projects/36/"+uid+".ics", href,
			"encoding must be a no-op for existing uuid uids or every client resyncs")
	})

	t.Run("a uid with slashes cannot escape its collection", func(t *testing.T) {
		href := taskURL(36, &models.Task{UID: `x/../../../projects/5/y`})

		assert.Equal(t, `/dav/projects/36/x%2F..%2F..%2F..%2Fprojects%2F5%2Fy.ics`, href)
		assert.True(t, strings.HasPrefix(href, ProjectBasePath+"/36/"), "href must stay in its own collection")
		assert.Len(t, strings.Split(strings.TrimPrefix(href, "/"), "/"), 4, "href must keep its path depth")
		assert.Equal(t, href, path.Clean(href), "href must not contain resolvable traversal segments")
	})

	t.Run("a uid with xml metacharacters yields a well formed href tag", func(t *testing.T) {
		href := taskURL(36, &models.Task{UID: `a<b>&c"d'e`})

		assert.Equal(t, `/dav/projects/36/a%3Cb%3E%26c%22d%27e.ics`, href)
		assert.NotContains(t, href, "<")
		assert.NotContains(t, href, ">")
		assert.NotContains(t, href, "&")

		// ixml.HrefTag does no escaping, so an unencoded href produces broken XML.
		var parsed struct {
			Href string `xml:",chardata"`
		}
		require.NoError(t, xml.Unmarshal([]byte(ixml.HrefTag(href)), &parsed))
		assert.Equal(t, href, parsed.Href)
	})
}

// The server must be able to resolve the hrefs it emits, or calendar-multiget
// silently drops the task.
func TestGetResourcesByList_HostileUIDRoundTrip(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	task := insertHostileTask(t, hostileUID)
	require.Equal(t, int64(1), countTasksWithUID(t, hostileUID))

	href := taskURL(task.ProjectID, task)
	storage := &VikunjaCaldavProjectStorage{user: caldavTestUser()}

	resources, err := storage.GetResourcesByList([]string{href})
	require.NoError(t, err)
	require.Len(t, resources, 1, "the server must resolve the href it emitted")
	assert.Equal(t, href, resources[0].Path)

	content, found := resources[0].GetContentData()
	require.True(t, found)
	assert.Contains(t, content, "UID:"+hostileUID)
}

func TestGetResourcesByList_MalformedEscape(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	storage := &VikunjaCaldavProjectStorage{user: caldavTestUser()}

	resources, err := storage.GetResourcesByList([]string{"/dav/projects/36/uid-caldav-%zz.ics"})
	require.NoError(t, err)
	assert.Empty(t, resources)
}

// Drives a real request through an echo router so the test pins how echo hands
// back the :task parameter.
func TestTaskHandler_HostileUIDRoundTrip(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	task := insertHostileTask(t, hostileUID)
	href := taskURL(task.ProjectID, task)

	e := echo.New()
	e.Any("/dav/projects/:project/:task", func(c *echo.Context) error {
		c.Set("userBasicAuth", caldavTestUser())
		return TaskHandler(c)
	})

	req := httptest.NewRequest(http.MethodGet, href, nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "UID:"+hostileUID)
}

// caldav-go echoes the decoded request path back as the href, so a client
// following one of our hrefs must not get a forged or unescaped one back.
func TestTaskHandler_PropfindHrefIsRebuilt(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	task := insertHostileTask(t, hostileUID)
	href := taskURL(task.ProjectID, task)

	e := echo.New()
	e.Any("/dav/projects/:project/:task", func(c *echo.Context) error {
		c.Set("userBasicAuth", caldavTestUser())
		return TaskHandler(c)
	})

	req := httptest.NewRequest("PROPFIND", href,
		strings.NewReader(`<?xml version="1.0"?><D:propfind xmlns:D="DAV:"><D:prop><D:getetag/></D:prop></D:propfind>`))
	req.Header.Set("Depth", "0")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	require.Equal(t, http.StatusMultiStatus, rec.Code)
	assert.Contains(t, rec.Body.String(), "<D:href>"+href+"</D:href>")

	var multistatus struct {
		Responses []struct {
			Href string `xml:"href"`
		} `xml:"response"`
	}
	require.NoError(t, xml.Unmarshal(rec.Body.Bytes(), &multistatus))
	require.Len(t, multistatus.Responses, 1)
	assert.Equal(t, href, multistatus.Responses[0].Href)
}
