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
	"encoding/xml"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePropfind(t *testing.T) {
	unknown := xml.Name{Space: "urn:example", Local: "x"}

	tests := []struct {
		name string
		body string
		want propfindRequest
	}{
		{name: "empty body", body: "", want: allPropRequest{}},
		{name: "whitespace body", body: " \n\t", want: allPropRequest{}},
		{name: "allprop", body: `<D:propfind xmlns:D="DAV:"><D:allprop/></D:propfind>`, want: allPropRequest{}},
		{
			name: "allprop with include",
			body: `<D:propfind xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav"><D:allprop/><D:include><C:calendar-data/></D:include></D:propfind>`,
			want: allPropRequest{include: []xml.Name{propCalendarData}},
		},
		{name: "propname", body: `<D:propfind xmlns:D="DAV:"><D:propname/></D:propfind>`, want: propNameRequest{}},
		{name: "root without children", body: `<D:propfind xmlns:D="DAV:"/>`, want: allPropRequest{}},
		{
			name: "prop in request order",
			body: `<D:propfind xmlns:D="DAV:" xmlns:X="urn:example"><D:prop><D:getetag/><X:x/><D:displayname/></D:prop></D:propfind>`,
			want: propRequest{names: []xml.Name{propGetETag, unknown, propDisplayName}},
		},
		{
			name: "prop listed twice is asked once",
			body: `<D:propfind xmlns:D="DAV:"><D:prop><D:getetag/><D:getetag/></D:prop></D:propfind>`,
			want: propRequest{names: []xml.Name{propGetETag}},
		},
		{
			name: "children of a property are not properties",
			body: `<D:propfind xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav"><D:prop><C:calendar-data><C:comp name="VCALENDAR"/></C:calendar-data></D:prop></D:propfind>`,
			want: propRequest{names: []xml.Name{propCalendarData}},
		},
		{
			name: "prop outside the DAV namespace is not the prop element",
			body: `<D:propfind xmlns:D="DAV:" xmlns:X="urn:example"><X:prop><D:getetag/></X:prop></D:propfind>`,
			want: allPropRequest{},
		},
		{name: "empty prop", body: `<D:propfind xmlns:D="DAV:"><D:prop/></D:propfind>`, want: propRequest{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parsePropfind(tc.body)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("at the property limit", func(t *testing.T) {
		props := make([]string, 0, maxPropfindProperties)
		for i := range maxPropfindProperties {
			props = append(props, `<X:p`+strconv.Itoa(i)+` xmlns:X="urn:example"/>`)
		}
		got, err := parsePropfind(propfindWithProps(props...))
		require.NoError(t, err)
		assert.Len(t, got.(propRequest).names, maxPropfindProperties)
	})

	for name, body := range map[string]string{
		"too many properties": tooManyProps(),
		"not xml":             "this is not xml",
		"unclosed root":       `<D:propfind xmlns:D="DAV:"><D:prop><D:getetag/>`,
		"mismatched tags":     `<D:propfind xmlns:D="DAV:"><D:prop></D:propfind>`,
	} {
		t.Run(name+" is rejected", func(t *testing.T) {
			_, err := parsePropfind(body)
			assert.Equal(t, http.StatusBadRequest, errorStatus(err), "got %v", err)
		})
	}
}

func TestPropfindMalformedBody(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	rec, err := callCaldavHandler(t, ProjectHandler, caldavFilterUser, "PROPFIND", projectPath(36), "this is not xml", projectParams(36))
	requireHTTPStatus(t, http.StatusBadRequest, rec, err)
}

func TestPropfindBodyTooLarge(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	body := propfindWithProps(`<D:getetag/>`) + strings.Repeat(" ", maxPropRequestBodySize)
	rec, err := callCaldavHandler(t, ProjectHandler, caldavFilterUser, "PROPFIND", projectPath(36), body, projectParams(36))
	requireHTTPStatus(t, http.StatusRequestEntityTooLarge, rec, err)
}

func TestPropfindOwner(t *testing.T) {
	owner := propfindWithProps(`<D:owner/>`)
	ownerOf := func(t *testing.T, handler func(*echo.Context) error, path string, params echo.PathValues) string {
		t.Helper()
		rec, err := callCaldavHandler(t, handler, caldavFilterUser, "PROPFIND", path, owner, params)
		requireHTTPStatus(t, http.StatusMultiStatus, rec, err)
		return rec.Body.String()
	}

	t.Run("a saved filter belongs to the user looking at it", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		pseudoID := createSavedFilter(t, caldavFilterUser, caldavFilterMatchingTasks)

		body := ownerOf(t, ProjectHandler, projectPath(pseudoID), projectParams(pseudoID))
		assert.Equal(t, 1+caldavFilterTaskCount, strings.Count(body, "<D:owner>"+principalHref(caldavFilterUser.Username)+"</D:owner>"), body)
	})

	t.Run("Favorites belong to the user looking at them", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		body := ownerOf(t, ProjectHandler, projectPath(models.FavoritesPseudoProjectID), projectParams(models.FavoritesPseudoProjectID))
		assert.Contains(t, body, "<D:owner>"+principalHref(caldavFilterUser.Username)+"</D:owner>")
	})

	t.Run("a task URL answers its project's owner", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		body := ownerOf(t, TaskHandler, projectPath(36)+"uid-caldav-test.ics", taskParams(36, "uid-caldav-test"))
		assert.Contains(t, body, "<D:owner>"+principalHref(caldavFilterUser.Username)+"</D:owner>")
	})
}
