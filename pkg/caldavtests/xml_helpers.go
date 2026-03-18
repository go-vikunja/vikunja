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
	"net/http/httptest"
	"strings"
	"testing"

	ics "github.com/arran4/golang-ical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Multistatus represents a WebDAV multistatus response (RFC 4918 §13)
type Multistatus struct {
	XMLName   xml.Name   `xml:"DAV: multistatus"`
	Responses []Response `xml:"response"`
}

// Response represents a single response within a multistatus
type Response struct {
	Href     string     `xml:"href"`
	Propstat []Propstat `xml:"propstat"`
}

// Propstat groups a set of properties with a status
type Propstat struct {
	Prop   Prop   `xml:"prop"`
	Status string `xml:"status"`
}

// Prop holds the actual property values returned by PROPFIND/REPORT.
type Prop struct {
	// Standard DAV properties
	DisplayName  string `xml:"displayname,omitempty"`
	ResourceType RawXML `xml:"resourcetype,omitempty"`
	GetETag      string `xml:"getetag,omitempty"`
	GetCTag      string `xml:"http://calendarserver.org/ns/ getctag,omitempty"`

	// CalDAV properties
	CalendarData        string `xml:"urn:ietf:params:xml:ns:caldav calendar-data,omitempty"`
	CalendarHomeSet     RawXML `xml:"urn:ietf:params:xml:ns:caldav calendar-home-set,omitempty"`
	SupportedComponents RawXML `xml:"urn:ietf:params:xml:ns:caldav supported-calendar-component-set,omitempty"`
	CalendarDescription string `xml:"urn:ietf:params:xml:ns:caldav calendar-description,omitempty"`

	// Principal properties
	CurrentUserPrincipal RawXML `xml:"current-user-principal,omitempty"`

	// ACL properties
	CurrentUserPrivilegeSet RawXML `xml:"current-user-privilege-set,omitempty"`

	// Catch-all for unexpected properties
	InnerXML string `xml:",innerxml"`
}

// RawXML captures raw XML content for properties we want to inspect flexibly
type RawXML struct {
	InnerXML string `xml:",innerxml"`
}

// parseMultistatus parses a WebDAV multistatus XML response body.
func parseMultistatus(t *testing.T, rec *httptest.ResponseRecorder) Multistatus {
	t.Helper()
	var ms Multistatus
	err := xml.Unmarshal(rec.Body.Bytes(), &ms)
	require.NoError(t, err, "Failed to parse multistatus XML. Body:\n%s", rec.Body.String())
	return ms
}

// findResponse finds a response in a multistatus by href substring match.
func findResponse(t *testing.T, ms Multistatus, hrefSubstring string) Response {
	t.Helper()
	for _, r := range ms.Responses {
		if strings.Contains(r.Href, hrefSubstring) {
			return r
		}
	}
	t.Fatalf("No response found with href containing %q in multistatus with %d responses", hrefSubstring, len(ms.Responses))
	return Response{} // unreachable
}

// getSuccessfulProp returns the Prop from the first propstat with a 200 status.
func getSuccessfulProp(t *testing.T, r Response) Prop {
	t.Helper()
	for _, ps := range r.Propstat {
		if strings.Contains(ps.Status, "200") {
			return ps.Prop
		}
	}
	t.Fatalf("No successful (200) propstat found in response for href %s", r.Href)
	return Prop{} // unreachable
}

// parseICalFromResponse parses iCalendar data from a response body.
func parseICalFromResponse(t *testing.T, rec *httptest.ResponseRecorder) *ics.Calendar {
	t.Helper()
	cal, err := ics.ParseCalendar(strings.NewReader(rec.Body.String()))
	require.NoError(t, err, "Failed to parse iCalendar. Body:\n%s", rec.Body.String())
	return cal
}

// parseICalFromString parses iCalendar data from a string (e.g., calendar-data property).
func parseICalFromString(t *testing.T, data string) *ics.Calendar {
	t.Helper()
	cal, err := ics.ParseCalendar(strings.NewReader(data))
	require.NoError(t, err, "Failed to parse iCalendar data:\n%s", data)
	return cal
}

// getVTodo extracts the first VTODO component from a calendar.
func getVTodo(t *testing.T, cal *ics.Calendar) *ics.VTodo {
	t.Helper()
	for _, comp := range cal.Components {
		if vtodo, ok := comp.(*ics.VTodo); ok {
			return vtodo
		}
	}
	t.Fatal("No VTODO component found in calendar")
	return nil // unreachable
}

// getVTodoProperty extracts a property value from a VTODO.
func getVTodoProperty(vtodo *ics.VTodo, prop ics.ComponentProperty) string {
	p := vtodo.GetProperty(prop)
	if p == nil {
		return ""
	}
	return p.Value
}

// assertResponseStatus asserts the HTTP status code.
func assertResponseStatus(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	assert.Equal(t, expectedStatus, rec.Code, "Response body:\n%s", rec.Body.String())
}

// assertMultistatusHasResponses asserts that a 207 response contains the expected number of responses.
func assertMultistatusHasResponses(t *testing.T, rec *httptest.ResponseRecorder, expectedCount int) Multistatus {
	t.Helper()
	assertResponseStatus(t, rec, 207)
	ms := parseMultistatus(t, rec)
	assert.Len(t, ms.Responses, expectedCount, "Expected %d responses in multistatus, got %d.\nBody:\n%s", expectedCount, len(ms.Responses), rec.Body.String())
	return ms
}
