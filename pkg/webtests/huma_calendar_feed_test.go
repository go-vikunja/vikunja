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
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixtures: user 13 owns project 20 and cannot read project 1 (user 1's).
func TestHumaCalendarFeed(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	s := db.NewSession()
	defer s.Close()
	task := &models.Task{
		Title:     "calendar-feed-task",
		ProjectID: 20,
		DueDate:   time.Date(2030, 1, 2, 9, 30, 0, 0, time.UTC),
	}
	require.NoError(t, task.Create(s, &testuser13))
	doneTask := &models.Task{
		Title:     "calendar-feed-done-task",
		ProjectID: 20,
		Done:      true,
		DueDate:   time.Date(2030, 1, 3, 9, 30, 0, 0, time.UTC),
	}
	require.NoError(t, doneTask.Create(s, &testuser13))
	require.NoError(t, s.Commit())

	feedURL := func(project string, query string) string {
		return "/api/v2/projects/" + project + "/calendar.ics?" + query
	}

	t.Run("project feed with a feeds token", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, feedURL("20", "token="+feedsTokenUser13), "", "", "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.True(t, strings.HasPrefix(rec.Header().Get(echo.HeaderContentType), "text/calendar"),
			"unexpected content type %q", rec.Header().Get(echo.HeaderContentType))

		body := rec.Body.String()
		assert.Contains(t, body, "BEGIN:VCALENDAR\r\n")
		assert.Contains(t, body, "X-WR-CALNAME:Test20\r\n")
		assert.Contains(t, body, "UID:vikunja-task-"+strconv.FormatInt(task.ID, 10)+"@localhost\r\n")
		assert.Contains(t, body, "SUMMARY:calendar-feed-task\r\n")
		assert.Contains(t, body, "DTSTART:20300102T093000Z\r\n")
		assert.Contains(t, body, "URL:https://localhost/tasks/"+strconv.FormatInt(task.ID, 10)+"\r\n")
		assert.NotContains(t, body, "calendar-feed-done-task")
	})

	t.Run("include_done adds done tasks", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, feedURL("20", "include_done=true&token="+feedsTokenUser13), "", "", "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "SUMMARY:calendar-feed-done-task\r\n")
	})

	t.Run("user feed", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/user/calendar.ics?token="+feedsTokenUser13, "", "", "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "SUMMARY:calendar-feed-task\r\n")
	})

	t.Run("forbidden for a project the token owner cannot read", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, feedURL("1", "token="+feedsTokenUser13), "", "", "")
		assert.Equal(t, http.StatusForbidden, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("nonexistent project", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, feedURL("9999999", "token="+feedsTokenUser13), "", "", "")
		assert.Equal(t, http.StatusNotFound, rec.Code, "body: %s", rec.Body.String())
	})

	for _, tc := range []struct {
		name  string
		query string
		jwt   string
	}{
		{
			name:  "missing token",
			query: "",
		},
		{
			name:  "unknown token",
			query: "token=tk_nonexistent_token_value_aaaaaaaaaaaaaaaa",
		},
		{
			name:  "token without feeds scope",
			query: "token=tk_readonly_tasks_user1_00000000abcd1234",
		},
		{
			name:  "session jwt instead of a query token",
			query: "",
			jwt:   humaTokenFor(t, &testuser13),
		},
	} {
		t.Run("unauthorized with "+tc.name, func(t *testing.T) {
			rec := humaRequest(t, e, http.MethodGet, feedURL("20", tc.query), "", tc.jwt, "")
			assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
		})
	}

	t.Run("the query token is not accepted by other routes", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/projects/20?token="+feedsTokenUser13, "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})

	t.Run("documented with query token security", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/openapi.json", "", "", "")
		require.Equal(t, http.StatusOK, rec.Code)

		var spec struct {
			Paths map[string]map[string]struct {
				Security  []map[string][]string `json:"security"`
				Responses map[string]struct {
					Content map[string]any `json:"content"`
				} `json:"responses"`
			} `json:"paths"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &spec))

		for _, path := range []string{"/projects/{project}/calendar.ics", "/user/calendar.ics"} {
			get, ok := spec.Paths[path]["get"]
			require.True(t, ok, "%s must document a GET operation", path)
			assert.Equal(t, []map[string][]string{{"FeedTokenQuery": {}}}, get.Security, path)
			assert.Contains(t, get.Responses["200"].Content, "text/calendar", path)
		}
	})
}
