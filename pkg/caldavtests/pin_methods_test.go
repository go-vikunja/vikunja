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
	methods := []string{
		"MKCALENDAR",
		"MKCOL",
		"COPY",
		"MOVE",
		"LOCK",
		"UNLOCK",
		http.MethodPost,
		http.MethodPatch,
	}
	paths := []string{
		"/dav/",
		"/dav/principals/user15/",
		"/dav/projects",
		"/dav/projects/36",
		pinTaskHref,
	}

	e := setupTestEnv(t)
	for _, path := range paths {
		for _, method := range methods {
			t.Run(method+" "+path, func(t *testing.T) {
				rec := caldavRequest(t, e, method, path, "", nil)

				assert.Equal(t, http.StatusNotImplemented, rec.Code, "Body:\n%s", rec.Body.String())
				assert.Empty(t, rec.Body.String())
			})
		}
	}
}
