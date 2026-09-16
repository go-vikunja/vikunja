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
// forge a path or inject XML markup - and the server must still resolve every
// href it hands out, whichever way the router decoded the request.

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/samedi/caldav-go/ixml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Project 36 in the fixtures belongs to caldavFilterUser and holds the caldav tasks.
const caldavHrefProjectID = 36

// A UID that tries both path forgery and XML injection at once.
const hostileUID = `evil/../../../projects/5/pwned<x>&y`

const propfindGetetag = `<?xml version="1.0"?><D:propfind xmlns:D="DAV:"><D:prop><D:getetag/></D:prop></D:propfind>`

func insertTaskWithUID(t *testing.T, uid string, index int64) {
	t.Helper()

	s := db.NewSession()
	defer s.Close()

	task := &models.Task{
		Title:       "Task with uid " + uid,
		UID:         uid,
		ProjectID:   caldavHrefProjectID,
		Index:       index,
		CreatedByID: caldavFilterUser.ID,
	}
	_, err := s.Insert(task)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
}

// Production builds its echo with UnescapePathParamValues (pkg/routes/routes.go), so a
// test on echo's default config exercises a decode path production never takes.
func newTaskRouter(unescapePathParamValues bool) *echo.Echo {
	e := echo.NewWithConfig(echo.Config{
		Router: echo.NewRouter(echo.RouterConfig{UnescapePathParamValues: unescapePathParamValues}),
	})
	e.Any("/dav/projects/:project/:task", func(c *echo.Context) error {
		c.Set("userBasicAuth", caldavFilterUser)
		return TaskHandler(c)
	})
	return e
}

func serve(e *echo.Echo, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

var uidRoundTripCases = []struct {
	name string
	uid  string
	href string
	// A second task whose uid a double-decode of uid resolves to.
	decoyUID string
}{
	{name: "uuid", uid: "550e8400-e29b-41d4-a716-446655440000", href: "/dav/projects/36/550e8400-e29b-41d4-a716-446655440000.ics"},
	// RFC 5545 recommends the addr-spec form, so encoding it would resync every client.
	{name: "addr-spec", uid: "task-1@host.example", href: "/dav/projects/36/task-1@host.example.ics"},
	{name: "literal percent", uid: "50%off-sale", href: "/dav/projects/36/50%25off-sale.ics"},
	{name: "percent escape lookalike", uid: "a%41b", href: "/dav/projects/36/a%2541b.ics", decoyUID: "aAb"},
	{name: "space", uid: "sale 2026 spring", href: "/dav/projects/36/sale%202026%20spring.ics"},
	{name: "utf-8", uid: "täsk-Ω", href: "/dav/projects/36/t%C3%A4sk-%CE%A9.ics"},
	{name: "path forgery and xml injection", uid: hostileUID, href: `/dav/projects/36/evil%2F..%2F..%2F..%2Fprojects%2F5%2Fpwned%3Cx%3E%26y.ics`},
}

func TestTaskUIDHrefRoundTrip(t *testing.T) {
	for _, tc := range uidRoundTripCases {
		t.Run(tc.name, func(t *testing.T) {
			href := taskURL(caldavHrefProjectID, &models.Task{UID: tc.uid})
			require.Equal(t, tc.href, href)

			// ixml.HrefTag does no escaping, so an unencoded href produces broken XML.
			var parsed struct {
				Href string `xml:",chardata"`
			}
			require.NoError(t, xml.Unmarshal([]byte(ixml.HrefTag(href)), &parsed))
			assert.Equal(t, href, parsed.Href)

			setup := func(t *testing.T) {
				db.LoadAndAssertFixtures(t)
				insertTaskWithUID(t, tc.uid, 98)
				if tc.decoyUID != "" {
					insertTaskWithUID(t, tc.decoyUID, 99)
				}
			}

			t.Run("multiget resolves the href", func(t *testing.T) {
				setup(t)

				resources, err := storageFor(caldavFilterUser, caldavHrefProjectID).GetResourcesByList([]string{href})
				require.NoError(t, err)
				require.Len(t, resources, 1, "the server must resolve the href it emitted")
				assert.Equal(t, href, resources[0].Path)

				content, found := resources[0].GetContentData()
				require.True(t, found)
				assert.Contains(t, content, "UID:"+tc.uid)
			})

			for _, unescapePathParamValues := range []bool{false, true} {
				t.Run("router unescape_path_param_values="+strconv.FormatBool(unescapePathParamValues), func(t *testing.T) {
					setup(t)
					e := newTaskRouter(unescapePathParamValues)

					rec := serve(e, httptest.NewRequest(http.MethodGet, href, nil))
					require.Equal(t, http.StatusOK, rec.Code)
					assert.Contains(t, rec.Body.String(), "UID:"+tc.uid)
					if tc.decoyUID != "" {
						assert.NotContains(t, rec.Body.String(), "UID:"+tc.decoyUID, "the uid must not be decoded twice")
					}

					req := httptest.NewRequest("PROPFIND", href, strings.NewReader(propfindGetetag))
					req.Header.Set("Depth", "0")
					rec = serve(e, req)
					require.Equal(t, http.StatusMultiStatus, rec.Code)

					// caldav-go builds the href from the decoded request path, so an
					// echoed one would be forged or unescaped.
					var multistatus struct {
						Responses []struct {
							Href string `xml:"href"`
						} `xml:"response"`
					}
					require.NoError(t, xml.Unmarshal(rec.Body.Bytes(), &multistatus))
					require.Len(t, multistatus.Responses, 1)
					assert.Equal(t, href, multistatus.Responses[0].Href)
				})
			}
		})
	}
}

func TestGetResourcesByList_MalformedEscape(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	insertTaskWithUID(t, "uid-caldav-%zz", 98)

	resources, err := storageFor(caldavFilterUser, caldavHrefProjectID).
		GetResourcesByList([]string{"/dav/projects/36/uid-caldav-%zz.ics"})
	require.NoError(t, err)
	assert.Empty(t, resources, "an href that is not valid percent-encoding is not an href we emitted")
}
