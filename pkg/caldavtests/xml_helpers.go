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
	"fmt"
	"net/http/httptest"
	"sort"
	"strconv"
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

const (
	nsDAV    = "DAV:"
	nsCalDAV = "urn:ietf:params:xml:ns:caldav"
	nsCS     = "http://calendarserver.org/ns/"
)

func davName(local string) xml.Name { return xml.Name{Space: nsDAV, Local: local} }
func calName(local string) xml.Name { return xml.Name{Space: nsCalDAV, Local: local} }

var csGetCTag = xml.Name{Space: nsCS, Local: "getctag"}

// xmlNode is a namespace-resolved XML element, independent of the prefixes the server chose.
type xmlNode struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Text     string     `xml:",chardata"`
	Children []xmlNode  `xml:",any"`
}

// canonical renders a property value so it can be compared regardless of prefixes,
// whitespace and child order: leaf text, DAV:href text, or "{ns}local[name=x]" per child.
func (n xmlNode) canonical() string {
	if len(n.Children) == 0 {
		return strings.TrimSpace(n.Text)
	}
	parts := make([]string, 0, len(n.Children))
	for _, c := range n.Children {
		if c.XMLName == davName("href") {
			parts = append(parts, strings.TrimSpace(c.Text))
			continue
		}
		part := "{" + c.XMLName.Space + "}" + c.XMLName.Local
		for _, a := range c.Attrs {
			if a.Name.Local == "name" {
				part += "[name=" + a.Value + "]"
			}
		}
		parts = append(parts, part)
	}
	sort.Strings(parts)
	return strings.Join(parts, " ")
}

// davResponse is one multistatus <response>, with properties grouped by propstat status.
type davResponse struct {
	Href string
	// Status is the response-level status, set when the response has no propstat (e.g. a missing multiget href).
	Status    int
	Propstats map[int]map[xml.Name]xmlNode
}

func parseDAVStatus(t *testing.T, line string) int {
	t.Helper()
	fields := strings.Fields(line)
	require.GreaterOrEqual(t, len(fields), 2, "malformed status line %q", line)
	code, err := strconv.Atoi(fields[1])
	require.NoError(t, err, "malformed status line %q", line)
	return code
}

// parseDAVResponses parses a multistatus body structurally, preserving response order and duplicates.
func parseDAVResponses(t *testing.T, rec *httptest.ResponseRecorder) []davResponse {
	t.Helper()
	var raw struct {
		XMLName   xml.Name `xml:"DAV: multistatus"`
		Responses []struct {
			Hrefs     []string `xml:"DAV: href"`
			Status    string   `xml:"DAV: status"`
			Propstats []struct {
				Prop struct {
					Props []xmlNode `xml:",any"`
				} `xml:"DAV: prop"`
				Status string `xml:"DAV: status"`
			} `xml:"DAV: propstat"`
		} `xml:"DAV: response"`
	}
	require.NoError(t, xml.Unmarshal(rec.Body.Bytes(), &raw), "Failed to parse multistatus XML. Body:\n%s", rec.Body.String())

	out := make([]davResponse, 0, len(raw.Responses))
	for _, r := range raw.Responses {
		require.Len(t, r.Hrefs, 1, "expected exactly one href per response. Body:\n%s", rec.Body.String())
		resp := davResponse{Href: r.Hrefs[0], Propstats: map[int]map[xml.Name]xmlNode{}}
		if r.Status != "" {
			resp.Status = parseDAVStatus(t, r.Status)
		}
		for _, ps := range r.Propstats {
			code := parseDAVStatus(t, ps.Status)
			require.NotContains(t, resp.Propstats, code, "duplicate %d propstat for %s", code, resp.Href)
			props := map[xml.Name]xmlNode{}
			for _, p := range ps.Prop.Props {
				require.NotContains(t, props, p.XMLName, "duplicate property %v for %s", p.XMLName, resp.Href)
				props[p.XMLName] = p
			}
			resp.Propstats[code] = props
		}
		out = append(out, resp)
	}
	return out
}

// anyValue marks an expected property whose value is not pinned, only its presence.
const anyValue = "\x00any"

// expectedDAVResponse pins one <response>: the 200 properties with their canonical values
// (or anyValue) and the names that must land in the 404 propstat.
type expectedDAVResponse struct {
	href     string
	status   int
	ok       map[xml.Name]string
	notFound []xml.Name
}

func diffDAVResponse(exp expectedDAVResponse, act davResponse) []string {
	var diffs []string
	if exp.status != act.Status {
		diffs = append(diffs, fmt.Sprintf("response status: want %d, got %d", exp.status, act.Status))
	}

	wantStatuses := map[int]bool{}
	if len(exp.ok) > 0 {
		wantStatuses[200] = true
	}
	if len(exp.notFound) > 0 {
		wantStatuses[404] = true
	}
	for code := range act.Propstats {
		if !wantStatuses[code] {
			diffs = append(diffs, fmt.Sprintf("unexpected %d propstat with %v", code, propNames(act.Propstats[code])))
		}
	}

	got200 := act.Propstats[200]
	for name, want := range exp.ok {
		got, has := got200[name]
		switch {
		case !has:
			diffs = append(diffs, fmt.Sprintf("missing from 200: %v", name))
		case want != anyValue && got.canonical() != want:
			diffs = append(diffs, fmt.Sprintf("%v: want %q, got %q", name, want, got.canonical()))
		}
	}
	for name := range got200 {
		if _, want := exp.ok[name]; !want {
			diffs = append(diffs, fmt.Sprintf("unexpected in 200: %v", name))
		}
	}

	wantNotFound := map[xml.Name]bool{}
	for _, n := range exp.notFound {
		wantNotFound[n] = true
		if _, has := act.Propstats[404][n]; !has {
			diffs = append(diffs, fmt.Sprintf("missing from 404: %v", n))
		}
	}
	for name := range act.Propstats[404] {
		if !wantNotFound[name] {
			diffs = append(diffs, fmt.Sprintf("unexpected in 404: %v", name))
		}
	}
	sort.Strings(diffs)
	return diffs
}

func propNames(props map[xml.Name]xmlNode) []xml.Name {
	names := make([]xml.Name, 0, len(props))
	for n := range props {
		names = append(names, n)
	}
	return names
}

// assertDAVResponses matches expected against actual responses as a multiset: order is
// not pinned, but every response (including duplicate hrefs) must be accounted for.
func assertDAVResponses(t *testing.T, rec *httptest.ResponseRecorder, expected []expectedDAVResponse) {
	t.Helper()
	actual := parseDAVResponses(t, rec)
	used := make([]bool, len(actual))

	for _, exp := range expected {
		var candidateDiffs []string
		matched := false
		for i, act := range actual {
			if used[i] || act.Href != exp.href {
				continue
			}
			diffs := diffDAVResponse(exp, act)
			if len(diffs) == 0 {
				used[i] = true
				matched = true
				break
			}
			candidateDiffs = append(candidateDiffs, strings.Join(diffs, "\n  "))
		}
		if !matched {
			assert.Failf(t, "no matching response", "href %s:\n  %s\nBody:\n%s",
				exp.href, strings.Join(candidateDiffs, "\n  ---\n  "), rec.Body.String())
		}
	}
	for i, act := range actual {
		if !used[i] {
			assert.Failf(t, "unexpected response", "href %s. Body:\n%s", act.Href, rec.Body.String())
		}
	}
}

// assertResponseStatus asserts the HTTP status code.
func assertResponseStatus(t *testing.T, rec *httptest.ResponseRecorder, expectedStatus int) { //nolint:unparam
	t.Helper()
	assert.Equal(t, expectedStatus, rec.Code, "Response body:\n%s", rec.Body.String())
}
