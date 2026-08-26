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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const propfindPrincipalAndData = `<?xml version="1.0" encoding="utf-8" ?>
<D:propfind xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
  <D:prop>
    <D:current-user-principal/>
    <C:calendar-data/>
  </D:prop>
</D:propfind>`

// Returns an error instead of failing the test so it can be called from a goroutine.
func propfind(handler func(*echo.Context) error, u *user.User, path string, params echo.PathValues) (string, error) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PROPFIND", path, strings.NewReader(propfindPrincipalAndData))
	req.Header.Set("Depth", "1")

	c := echo.New().NewContext(req, rec)
	c.Set("userBasicAuth", u)
	if params != nil {
		c.SetPathValues(params)
	}

	if err := handler(c); err != nil {
		return "", err
	}
	if rec.Code != http.StatusMultiStatus {
		return "", fmt.Errorf("PROPFIND %s answered %d: %s", path, rec.Code, rec.Body.String())
	}
	return rec.Body.String(), nil
}

func projectPropfind(u *user.User, projectID int64) (string, error) {
	return propfind(ProjectHandler, u, projectPath(projectID), echo.PathValues{
		{Name: "project", Value: strconv.FormatInt(projectID, 10)},
	})
}

func taskPropfind(u *user.User, projectID int64, taskUID string) (string, error) {
	return propfind(TaskHandler, u, projectPath(projectID)+taskUID+".ics", echo.PathValues{
		{Name: "project", Value: strconv.FormatInt(projectID, 10)},
		{Name: "task", Value: taskUID + ".ics"},
	})
}

func principalPropfind(u *user.User) (string, error) {
	return propfind(PrincipalHandler, u, principalPathForUser(u.Username), nil)
}

func principalHref(username string) string {
	return "<D:href>" + principalPathForUser(username) + "</D:href>"
}

// The caldav-go setup used to live in package-level globals, so a request could be answered
// from another concurrent request's storage. Run with -race.
func TestConcurrentRequestsDoNotShareState(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	callers := []struct {
		user      *user.User
		projectID int64
		mine      string
		theirs    []string
	}{
		{caldavFilterUser, 36, "Title Caldav Test", []string{"task #1", ProjectBasePath + "/1/"}},
		{caldavOtherUser, 1, "task #1", []string{"Title Caldav Test", ProjectBasePath + "/36/"}},
	}

	// A single pair of requests only rarely interleaves inside the setup-then-handle window,
	// so hammer it from several goroutines per user.
	const workers = 8
	const rounds = 25

	type result struct {
		caller int
		body   string
		err    error
	}
	results := make(chan result, workers*rounds*len(callers))

	start := make(chan struct{})
	var wg sync.WaitGroup
	for i, caller := range callers {
		for range workers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				for range rounds {
					body, err := projectPropfind(caller.user, caller.projectID)
					results <- result{caller: i, body: body, err: err}
				}
			}()
		}
	}
	close(start)
	wg.Wait()
	close(results)

	for r := range results {
		caller := callers[r.caller]
		require.NoError(t, r.err)
		assert.Contains(t, r.body, caller.mine,
			"%s must see its own tasks", caller.user.Username)
		assert.Contains(t, r.body, "<D:href>"+ProjectHomeSetPath+"</D:href>")
		for _, theirs := range caller.theirs {
			assert.NotContains(t, r.body, theirs,
				"%s must not see %q of the other user", caller.user.Username, theirs)
		}
	}
}

// TaskHandler set no user at all, so it answered current-user-principal from whatever
// the last request had left in the globals.
func TestTaskHandlerCurrentUserPrincipal(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	principal, err := principalPropfind(caldavOtherUser)
	require.NoError(t, err)
	require.Contains(t, principal, principalHref(caldavOtherUser.Username))

	body, err := taskPropfind(caldavFilterUser, 36, "uid-caldav-test")
	require.NoError(t, err)
	assert.Contains(t, body, principalHref(caldavFilterUser.Username))
	assert.NotContains(t, body, principalHref(caldavOtherUser.Username))
}
