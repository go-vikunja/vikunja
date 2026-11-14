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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests pin what the server answers today, including the parts that look wrong.

const (
	pinTaskHref     = "/dav/projects/36/uid-caldav-test.ics"
	user15Principal = "/dav/principals/user15/"
	homeSetHref     = "/dav/projects/"
	zeroLastMod     = "Mon, 01 Jan 0001 00:00:00 GMT"

	resourceTypeHome      = "{DAV:}collection"
	resourceTypeEntry     = "{DAV:}collection"
	resourceTypePrincipal = "{DAV:}collection {DAV:}principal"
	resourceTypeCalendar  = "{DAV:}collection {urn:ietf:params:xml:ns:caldav}calendar"
	supportedComps        = "{urn:ietf:params:xml:ns:caldav}comp[name=VCALENDAR] {urn:ietf:params:xml:ns:caldav}comp[name=VTODO]"
)

var project36TaskTitles = map[string]string{
	"/dav/projects/36/uid-caldav-test.ics":                          "Title Caldav Test",
	"/dav/projects/36/uid-caldav-test-parent-task.ics":              "Parent task for Caldav Test",
	"/dav/projects/36/uid-caldav-test-parent-task-2.ics":            "Parent task for Caldav Test 2",
	"/dav/projects/36/uid-caldav-test-child-task.ics":               "Child task for Caldav Test",
	"/dav/projects/36/uid-caldav-test-parent-task-another-list.ics": "Parent task for Caldav Test",
}

const user6Principal = "/dav/principals/user6/"

// user15's home set: own projects 36 and 38 plus read shares on user6's 26 and 35.
var user15Projects = map[string]struct{ title, owner string }{
	"/dav/projects/36": {"Project 36 for Caldav tests", user15Principal},
	"/dav/projects/38": {"Project 38 for Caldav tests", user15Principal},
	"/dav/projects/35": {"Test35 with background", user6Principal},
	"/dav/projects/26": {"Test26", user6Principal},
}

func expectEntry() expectedDAVResponse {
	return expectedDAVResponse{
		href: "/dav/",
		ok: map[xml.Name]string{
			davName("resourcetype"):                     resourceTypeEntry,
			davName("displayname"):                      "dav",
			davName("current-user-principal"):           user15Principal,
			calName("calendar-home-set"):                homeSetHref,
			davName("principal-URL"):                    user15Principal,
			calName("supported-calendar-component-set"): supportedComps,
			davName("getcontenttype"):                   "text/calendar",
		},
		notFound: []xml.Name{
			davName("owner"), davName("getetag"), csGetCTag, davName("getlastmodified"),
			davName("getcontentlength"), davName("sync-token"), calName("calendar-description"),
		},
	}
}

// href is the principal path, or the well-known path the principal was asked as.
func expectPrincipal(href, displayName string) expectedDAVResponse {
	return expectedDAVResponse{
		href: href,
		ok: map[xml.Name]string{
			davName("resourcetype"):                     resourceTypePrincipal,
			davName("displayname"):                      displayName,
			davName("current-user-principal"):           user15Principal,
			calName("calendar-home-set"):                homeSetHref,
			davName("principal-URL"):                    user15Principal,
			calName("supported-calendar-component-set"): supportedComps,
			davName("getcontenttype"):                   "text/calendar",
		},
		notFound: []xml.Name{
			davName("owner"), davName("getetag"), csGetCTag, davName("getlastmodified"),
			davName("getcontentlength"), davName("sync-token"), calName("calendar-description"),
		},
	}
}

func expectHomeSet() expectedDAVResponse {
	return expectedDAVResponse{
		href: "/dav/projects",
		ok: map[xml.Name]string{
			davName("resourcetype"):                     resourceTypeHome,
			davName("displayname"):                      "projects",
			davName("current-user-principal"):           user15Principal,
			calName("supported-calendar-component-set"): supportedComps,
			davName("getcontenttype"):                   "text/calendar",
		},
		notFound: []xml.Name{
			davName("owner"), davName("getetag"), csGetCTag, davName("getlastmodified"),
			davName("getcontentlength"), davName("sync-token"), calName("calendar-description"),
			calName("calendar-home-set"), davName("principal-URL"),
		},
	}
}

func expectCalendar(href, title, owner string) expectedDAVResponse {
	return expectedDAVResponse{
		href: href,
		ok: map[xml.Name]string{
			davName("resourcetype"):                     resourceTypeCalendar,
			davName("displayname"):                      title,
			davName("current-user-principal"):           user15Principal,
			calName("supported-calendar-component-set"): supportedComps,
			davName("getcontenttype"):                   "text/calendar",
			davName("owner"):                            owner,
			davName("getetag"):                          anyValue,
			csGetCTag:                                   anyValue,
			davName("getlastmodified"):                  anyValue,
			davName("sync-token"):                       anyValue,
		},
		notFound: []xml.Name{
			davName("getcontentlength"), calName("calendar-description"),
			calName("calendar-home-set"), davName("principal-URL"),
		},
	}
}

func expectTask(href, title string) expectedDAVResponse {
	return expectedDAVResponse{
		href: href,
		ok: map[xml.Name]string{
			davName("resourcetype"):           "",
			davName("displayname"):            title,
			davName("current-user-principal"): user15Principal,
			davName("getcontenttype"):         "text/calendar; component=vcalendar",
			davName("owner"):                  user15Principal,
			davName("getetag"):                anyValue,
			csGetCTag:                         anyValue,
			davName("getlastmodified"):        anyValue,
			davName("getcontentlength"):       anyValue,
		},
		notFound: []xml.Name{
			calName("supported-calendar-component-set"), davName("sync-token"), calName("calendar-description"),
			calName("calendar-home-set"), davName("principal-URL"),
		},
	}
}

func project36Tasks() []expectedDAVResponse {
	out := make([]expectedDAVResponse, 0, len(project36TaskTitles))
	for href, title := range project36TaskTitles {
		out = append(out, expectTask(href, title))
	}
	return out
}

func minimal(responses []expectedDAVResponse) []expectedDAVResponse {
	out := make([]expectedDAVResponse, len(responses))
	for i, r := range responses {
		r.notFound = nil
		out[i] = r
	}
	return out
}

func TestPinPropfindMatrix(t *testing.T) {
	type resourceKind struct {
		name    string
		path    string
		shallow []expectedDAVResponse
		deep    []expectedDAVResponse
	}

	homeDeep := []expectedDAVResponse{expectHomeSet()}
	for href, p := range user15Projects {
		homeDeep = append(homeDeep, expectCalendar(href, p.title, p.owner))
	}

	task := expectTask(pinTaskHref, "Title Caldav Test")

	kinds := []resourceKind{
		{
			name:    "entry",
			path:    "/dav/",
			shallow: []expectedDAVResponse{expectEntry()},
			deep:    []expectedDAVResponse{expectEntry()},
		},
		{
			name:    "principal",
			path:    "/dav/principals/user15/",
			shallow: []expectedDAVResponse{expectPrincipal(user15Principal, "user15")},
			deep:    []expectedDAVResponse{expectPrincipal(user15Principal, "user15")},
		},
		{
			name:    "well-known",
			path:    "/.well-known/caldav",
			shallow: []expectedDAVResponse{expectPrincipal("/.well-known/caldav", "caldav")},
			deep:    []expectedDAVResponse{expectPrincipal("/.well-known/caldav", "caldav")},
		},
		{
			name:    "home set",
			path:    "/dav/projects",
			shallow: []expectedDAVResponse{expectHomeSet()},
			deep:    homeDeep,
		},
		{
			name:    "project collection",
			path:    "/dav/projects/36",
			shallow: []expectedDAVResponse{expectCalendar("/dav/projects/36", "Project 36 for Caldav tests", user15Principal)},
			deep: append(
				[]expectedDAVResponse{expectCalendar("/dav/projects/36", "Project 36 for Caldav tests", user15Principal)},
				project36Tasks()...,
			),
		},
		{
			name:    "task resource",
			path:    pinTaskHref,
			shallow: []expectedDAVResponse{task},
			// A task has no members.
			deep: []expectedDAVResponse{task},
		},
	}

	// PROPFIND does not write, so all subtests share one fixture load.
	e := setupTestEnv(t)
	for _, kind := range kinds {
		for _, depth := range []string{"0", "1", "infinity"} {
			expected := kind.shallow
			if depth == "1" {
				expected = kind.deep
			}
			for _, isMinimal := range []bool{false, true} {
				name := kind.name + " Depth " + depth
				headers := map[string]string{"Depth": depth}
				want := expected
				if isMinimal {
					name += " minimal"
					headers["Prefer"] = "return=minimal"
					want = minimal(expected)
				}

				t.Run(name, func(t *testing.T) {
					rec := caldavRequest(t, e, "PROPFIND", kind.path, PropfindDiscoveryProps, headers)

					require.Equal(t, 207, rec.Code, "Body:\n%s", rec.Body.String())
					assertDAVResponses(t, rec, want)

					applied, has := rec.Header()["Preference-Applied"]
					if isMinimal {
						assert.Equal(t, []string{"return=minimal"}, applied)
					} else {
						assert.False(t, has, "Preference-Applied must be absent without Prefer, got %v", applied)
					}
				})
			}
		}
	}
}

func TestPinPropfindTaskURLReportsTaskETag(t *testing.T) {
	e := setupTestEnv(t)

	etagOf := func(path string) string {
		rec := caldavPROPFIND(t, e, path, "0", PropfindDiscoveryProps)
		require.Equal(t, 207, rec.Code)
		responses := parseDAVResponses(t, rec)
		require.Len(t, responses, 1)
		etag := responses[0].Propstats[200][davName("getetag")].canonical()
		require.NotEmpty(t, etag)
		return etag
	}

	taskETag := etagOf(pinTaskHref)
	assert.NotEqual(t, etagOf("/dav/projects/36"), taskETag)

	get := caldavGET(t, e, pinTaskHref)
	require.Equal(t, 200, get.Code)
	assert.Equal(t, taskETag, get.Header().Get("ETag"))
}

func TestPinPropfindMissingTaskURL(t *testing.T) {
	e := setupTestEnv(t)

	for _, depth := range []string{"0", "1"} {
		rec := caldavPROPFIND(t, e, "/dav/projects/36/does-not-exist.ics", depth, PropfindDiscoveryProps)
		assert.Equal(t, 404, rec.Code, "Depth %s. Body:\n%s", depth, rec.Body.String())
	}
}

// RFC 4918 treats an empty PROPFIND body as allprop; today it yields bare hrefs.
func TestPinPropfindEmptyBody(t *testing.T) {
	e := setupTestEnv(t)
	for path, href := range map[string]string{
		"/dav/":                   "/dav/",
		"/dav/principals/user15/": user15Principal,
		"/dav/projects":           "/dav/projects",
		"/dav/projects/36":        "/dav/projects/36",
		pinTaskHref:               pinTaskHref,
	} {
		t.Run(path, func(t *testing.T) {
			rec := caldavPROPFIND(t, e, path, "0", "")

			require.Equal(t, 207, rec.Code, "Body:\n%s", rec.Body.String())
			assertDAVResponses(t, rec, []expectedDAVResponse{{href: href}})
		})
	}
}

// The unknown property is echoed in the 404 propstat without its namespace.
func TestPinPropfindUnknownNamespace(t *testing.T) {
	e := setupTestEnv(t)
	body := `<?xml version="1.0" encoding="utf-8" ?>
<D:propfind xmlns:D="DAV:" xmlns:X="urn:x-vikunja-unknown">
  <D:prop>
    <D:displayname/>
    <X:bogus/>
  </D:prop>
</D:propfind>`

	rec := caldavPROPFIND(t, e, "/dav/projects/36", "0", body)

	require.Equal(t, 207, rec.Code, "Body:\n%s", rec.Body.String())
	assertDAVResponses(t, rec, []expectedDAVResponse{{
		href:     "/dav/projects/36",
		ok:       map[xml.Name]string{davName("displayname"): "Project 36 for Caldav tests"},
		notFound: []xml.Name{{Space: "", Local: "bogus"}},
	}})
}
