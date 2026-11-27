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

const (
	davComplianceHeader = "1, 3, calendar-access"
	allowReadWrite      = "GET, HEAD, PUT, DELETE, OPTIONS, PROPFIND, REPORT"
	allowReadOnly       = "GET, HEAD, OPTIONS, PROPFIND, REPORT"
	allowHomeSet        = "OPTIONS, PROPFIND, REPORT"
)

func TestPinOPTIONS(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		status int
		allow  string
	}{
		{
			name:   "entry",
			path:   "/dav/",
			status: http.StatusOK,
			allow:  allowReadOnly,
		},
		{
			name:   "well-known",
			path:   "/.well-known/caldav",
			status: http.StatusOK,
			allow:  allowReadOnly,
		},
		{
			name:   "principal",
			path:   "/dav/principals/user15/",
			status: http.StatusOK,
			allow:  allowReadOnly,
		},
		{
			name:   "home set",
			path:   "/dav/projects",
			status: http.StatusOK,
			allow:  allowHomeSet,
		},
		{
			name:   "own project",
			path:   "/dav/projects/36",
			status: http.StatusOK,
			allow:  allowReadWrite,
		},
		{
			name:   "task in own project",
			path:   pinTaskHref,
			status: http.StatusOK,
			allow:  allowReadWrite,
		},
		{
			name:   "read-only shared project",
			path:   "/dav/projects/26",
			status: http.StatusOK,
			allow:  allowReadOnly,
		},
		{
			name:   "project the user cannot see",
			path:   "/dav/projects/1",
			status: http.StatusNotFound,
		},
		{
			name:   "missing project",
			path:   "/dav/projects/99999",
			status: http.StatusNotFound,
		},
		{
			name:   "task in read-only shared project",
			path:   "/dav/projects/26/any.ics",
			status: http.StatusOK,
			allow:  allowReadOnly,
		},
		{
			name:   "task in project the user cannot see",
			path:   "/dav/projects/1/any.ics",
			status: http.StatusNotFound,
		},
		{
			name:   "task in favorites",
			path:   "/dav/projects/-1/any.ics",
			status: http.StatusOK,
			allow:  allowReadWrite,
		},
	}

	e := setupTestEnv(t)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := caldavOPTIONS(t, e, tc.path)

			assert.Equal(t, tc.status, rec.Code, "Body:\n%s", rec.Body.String())
			if tc.status != http.StatusOK {
				assert.Empty(t, rec.Header().Values("DAV"))
				assert.Empty(t, rec.Header().Values("Allow"))
				return
			}
			assert.Equal(t, []string{davComplianceHeader}, rec.Header().Values("DAV"))
			assert.Equal(t, []string{tc.allow}, rec.Header().Values("Allow"))
			assert.Empty(t, rec.Body.String())
		})
	}
}
