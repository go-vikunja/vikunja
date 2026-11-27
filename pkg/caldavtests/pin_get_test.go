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
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertLastModified(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	lastMod := rec.Header().Get("Last-Modified")
	parsed, err := http.ParseTime(lastMod)
	require.NoError(t, err, "Last-Modified %q", lastMod)
	assert.Greater(t, parsed.Year(), 2000, "Last-Modified %q", lastMod)
}

func TestPinGETAndHEADTask(t *testing.T) {
	e := setupTestEnv(t)

	get := caldavGET(t, e, pinTaskHref)
	head := caldavRequest(t, e, http.MethodHead, pinTaskHref, "", nil)

	require.Equal(t, http.StatusOK, get.Code)
	require.Equal(t, http.StatusOK, head.Code)

	assert.Equal(t, "text/calendar; component=vcalendar", get.Header().Get("Content-Type"))
	etag := get.Header().Get("ETag")
	assert.Regexp(t, `^"40-\d+"$`, etag)
	assertLastModified(t, get)
	assert.Equal(t, 1, strings.Count(get.Body.String(), "BEGIN:VTODO"))

	for _, h := range []string{
		"Content-Type",
		"ETag",
		"Last-Modified",
	} {
		assert.Equal(t, get.Header().Get(h), head.Header().Get(h), h)
	}
	assert.Empty(t, head.Body.String())
}

func TestPinHEADMissingTask(t *testing.T) {
	e := setupTestEnv(t)

	rec := caldavRequest(t, e, http.MethodHead, "/dav/projects/36/nonexistent-uid.ics", "", nil)

	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Empty(t, rec.Body.String())
}

func TestPinGETCollection(t *testing.T) {
	e := setupTestEnv(t)

	rec := caldavGET(t, e, "/dav/projects/36")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/calendar", rec.Header().Get("Content-Type"))
	assertLastModified(t, rec)
	assert.Equal(t, len(project36TaskTitles), strings.Count(rec.Body.String(), "BEGIN:VTODO"))

	propfind := caldavPROPFIND(t, e, "/dav/projects/36", "0", PropfindDiscoveryProps)
	require.Equal(t, 207, propfind.Code)
	responses := parseDAVResponses(t, propfind)
	require.Len(t, responses, 1)
	assert.Equal(t, responses[0].Propstats[200][davName("getetag")].canonical(), rec.Header().Get("ETag"))
}

func TestPinGETNonCalendarResources(t *testing.T) {
	t.Run("home set", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavGET(t, e, "/dav/projects")

		assert.Equal(t, http.StatusInternalServerError, rec.Code, "Body:\n%s", rec.Body.String())
	})

	// An empty calendar body with an empty ETag and a zero Last-Modified.
	for _, path := range []string{
		"/dav/",
		"/dav/principals/user15/",
	} {
		t.Run(path, func(t *testing.T) {
			e := setupTestEnv(t)

			rec := caldavGET(t, e, path)

			require.Equal(t, http.StatusOK, rec.Code, "Body:\n%s", rec.Body.String())
			assert.Equal(t, "text/calendar; component=vcalendar", rec.Header().Get("Content-Type"))
			assert.Equal(t, []string{""}, rec.Header()["Etag"])
			assert.Equal(t, zeroLastMod, rec.Header().Get("Last-Modified"))
			assert.Empty(t, rec.Body.String())
		})
	}
}
