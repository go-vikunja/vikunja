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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const proppatchDisplayName = `<D:propertyupdate xmlns:D="DAV:"><D:set><D:prop><D:displayname>x</D:displayname></D:prop></D:set></D:propertyupdate>`

func TestPinPropPatchAccess(t *testing.T) {
	e := setupTestEnv(t)
	asUser1 := map[string]string{"Authorization": basicAuthHeader(testuser1.Username, fixturePassword)}

	t.Run("readable but not writable is refused per property", func(t *testing.T) {
		// user1 has a read share on project 9.
		rec := caldavRequest(t, e, "PROPPATCH", "/dav/projects/9", proppatchDisplayName, asUser1)
		require.Equal(t, http.StatusMultiStatus, rec.Code, "Body:\n%s", rec.Body.String())
		assert.Contains(t, rec.Body.String(), "403 Forbidden")
	})

	t.Run("unreadable looks missing, before the body is looked at", func(t *testing.T) {
		// user1 cannot read project 36; 999999 does not exist.
		for _, body := range []string{proppatchDisplayName, "not xml", strings.Repeat(" ", 65*1024)} {
			for _, path := range []string{"/dav/projects/36", "/dav/projects/999999"} {
				rec := caldavRequest(t, e, "PROPPATCH", path, body, asUser1)
				assert.Equal(t, http.StatusNotFound, rec.Code, "PROPPATCH %s with a %d byte body", path, len(body))
			}
		}
	})
}
