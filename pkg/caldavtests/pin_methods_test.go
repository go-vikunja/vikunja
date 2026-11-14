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

func TestPinUnsupportedMethods(t *testing.T) {
	methods := []string{"MKCALENDAR", "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK", http.MethodPost, http.MethodPatch}
	allow := map[string]string{
		"/dav":                    allowEntry,
		"/dav/":                   allowEntry,
		"/.well-known/caldav":     allowEntry,
		"/.well-known/caldav/":    allowEntry,
		"/dav/principals/user15":  allowEntry,
		"/dav/principals/user15/": allowEntry,
		"/dav/projects":           allowHomeSet,
		"/dav/projects/":          allowHomeSet,
		"/dav/projects/36":        allowReadWrite,
		"/dav/projects/36/":       allowReadWrite,
		pinTaskHref:               allowReadWrite,
	}

	e := setupTestEnv(t)
	for path, want := range allow {
		for _, method := range methods {
			t.Run(method+" "+path, func(t *testing.T) {
				rec := caldavRequest(t, e, method, path, "", nil)

				assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, "Body:\n%s", rec.Body.String())
				assert.Equal(t, want, rec.Header().Get("Allow"))
				assert.Empty(t, rec.Body.String())
			})
		}
	}

	// PROPPATCH is answered on collections and tasks only.
	for _, path := range []string{"/dav/", user15Principal, "/dav/projects"} {
		rec := caldavRequest(t, e, "PROPPATCH", path, "", nil)
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, "PROPPATCH %s", path)
		assert.Equal(t, allow[path], rec.Header().Get("Allow"), "PROPPATCH %s", path)
	}
}

func TestPinUnsupportedMethodsAccess(t *testing.T) {
	e := setupTestEnv(t)

	t.Run("unauthenticated", func(t *testing.T) {
		for _, path := range []string{"/dav/", "/.well-known/caldav", user15Principal, "/dav/projects", "/dav/projects/36", pinTaskHref} {
			rec := caldavRequest(t, e, "MKCALENDAR", path, "", map[string]string{"Authorization": basicAuthHeader(testuser15.Username, "wrong")})
			assert.Equal(t, http.StatusUnauthorized, rec.Code, "MKCALENDAR %s", path)
			assert.Empty(t, rec.Header().Get("Allow"), "MKCALENDAR %s", path)
		}
	})

	t.Run("unreadable or missing resource", func(t *testing.T) {
		// user1 cannot read project 36; 999999 does not exist.
		as := map[string]string{"Authorization": basicAuthHeader(testuser1.Username, fixturePassword)}
		for _, path := range []string{"/dav/projects/36", "/dav/projects/36/", pinTaskHref, "/dav/projects/999999", "/dav/projects/999999/x.ics", user15Principal} {
			rec := caldavRequest(t, e, "MKCALENDAR", path, "", as)
			assert.Equal(t, http.StatusNotFound, rec.Code, "MKCALENDAR %s", path)
			assert.Empty(t, rec.Header().Get("Allow"), "MKCALENDAR %s", path)
		}
	})

	t.Run("read-only collection", func(t *testing.T) {
		// user1 has a read share on user6's project 9.
		as := map[string]string{"Authorization": basicAuthHeader(testuser1.Username, fixturePassword)}
		rec := caldavRequest(t, e, "MKCALENDAR", "/dav/projects/9", "", as)
		assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
		assert.Equal(t, allowReadOnly, rec.Header().Get("Allow"))
	})
}

func TestPinWritesOnEntryAndPrincipal(t *testing.T) {
	e := setupTestEnv(t)

	for _, path := range []string{"/dav/", "/.well-known/caldav", user15Principal} {
		for _, headers := range []map[string]string{nil, {"If-Match": `"stale"`}} {
			for _, method := range []string{http.MethodPut, http.MethodDelete} {
				t.Run(method+" "+path, func(t *testing.T) {
					rec := caldavRequest(t, e, method, path, NewVTodo("pin-entry-write", "Entry").Build(), headers)

					assert.Equal(t, http.StatusMethodNotAllowed, rec.Code, "Body:\n%s", rec.Body.String())
					assert.Equal(t, allowEntry, rec.Header().Get("Allow"))
				})
			}
		}
	}
}
