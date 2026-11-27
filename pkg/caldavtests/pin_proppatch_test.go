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
	"encoding/xml"
	"net/http"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const proppatchDisplayName = `<D:propertyupdate xmlns:D="DAV:"><D:set><D:prop><D:displayname>x</D:displayname></D:prop></D:set></D:propertyupdate>`

const proppatchCalendarColor = `<D:propertyupdate xmlns:D="DAV:"><D:set><D:prop>` +
	`<A:calendar-color xmlns:A="http://apple.com/ns/ical/">#007AFF</A:calendar-color>` +
	`</D:prop></D:set></D:propertyupdate>`

func TestPinPropPatchAccess(t *testing.T) {
	e := setupTestEnv(t)
	asUser1 := map[string]string{"Authorization": basicAuthHeader(testuser1.Username, fixturePassword)}

	t.Run("readable but not writable is refused per property", func(t *testing.T) {
		// user1 has a read share on project 9.
		rec := caldavRequest(t, e, "PROPPATCH", "/dav/projects/9", proppatchCalendarColor, asUser1)
		require.Equal(t, http.StatusMultiStatus, rec.Code, "Body:\n%s", rec.Body.String())
		responses := parseDAVResponses(t, rec)
		require.Len(t, responses, 1)
		assert.Len(t, responses[0].Propstats, 1, "Body:\n%s", rec.Body.String())
		assert.Equal(t, []xml.Name{
			{
				Space: "http://apple.com/ns/ical/",
				Local: "calendar-color",
			},
		}, propNames(responses[0].Propstats[http.StatusForbidden]))

		s := db.NewSession()
		defer s.Close()
		stored, err := models.GetProjectSimpleByID(s, 9)
		require.NoError(t, err)
		assert.Empty(t, stored.HexColor)
	})

	t.Run("unreadable looks missing, before the body is looked at", func(t *testing.T) {
		// user1 cannot read project 36; 999999 does not exist.
		for _, body := range []string{
			proppatchDisplayName,
			"not xml",
			strings.Repeat(" ", 65*1024),
		} {
			for _, path := range []string{
				"/dav/projects/36",
				"/dav/projects/999999",
			} {
				rec := caldavRequest(t, e, "PROPPATCH", path, body, asUser1)
				assert.Equal(t, http.StatusNotFound, rec.Code, "PROPPATCH %s with a %d byte body", path, len(body))
			}
		}
	})
}

func TestPinPropPatchTarget(t *testing.T) {
	e := setupTestEnv(t)

	t.Run("answers under the canonical href", func(t *testing.T) {
		const href = "/dav/projects/36/pp%20a%26b.ics"
		require.Equal(t, http.StatusCreated, caldavPUT(t, e, href, NewVTodo("pp a&b", "Reserved").Build()).Code)

		for path, want := range map[string]string{
			href:                "/dav/projects/36/pp%20a%26b.ics",
			pinTaskHref:         pinTaskHref,
			"/dav/projects/36/": "/dav/projects/36",
		} {
			rec := caldavRequest(t, e, "PROPPATCH", path, proppatchDisplayName, nil)
			require.Equal(t, http.StatusMultiStatus, rec.Code, "PROPPATCH %s:\n%s", path, rec.Body.String())
			ms := parseMultistatus(t, rec)
			require.Len(t, ms.Responses, 1)
			assert.Equal(t, want, ms.Responses[0].Href, "PROPPATCH %s", path)
		}
	})

	t.Run("a task that is not in the collection is missing", func(t *testing.T) {
		// uid-caldav-test lives in project 36; user15 can read 38 too.
		for _, path := range []string{
			"/dav/projects/36/does-not-exist.ics",
			"/dav/projects/38/uid-caldav-test.ics",
		} {
			rec := caldavRequest(t, e, "PROPPATCH", path, proppatchDisplayName, nil)
			assert.Equal(t, http.StatusNotFound, rec.Code, "PROPPATCH %s", path)
		}
	})
}
