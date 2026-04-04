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
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscovery(t *testing.T) {
	// RFC 6764 §5 (rfc6764.txt line 205):
	// "A CalDAV server SHOULD provide a well-known URI that redirects
	//  to the context path of the CalDAV service."

	t.Run("well-known/caldav responds", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavRequest(t, e, "PROPFIND", "/.well-known/caldav", PropfindCurrentUserPrincipal, map[string]string{
			"Depth": "0",
		})

		// Should get either a redirect (301/302) or a 207 with principal info
		// Both are acceptable per RFC 6764 §5
		assert.True(t,
			rec.Code == http.StatusMovedPermanently ||
				rec.Code == http.StatusFound ||
				rec.Code == http.StatusMultiStatus,
			"Expected 301, 302, or 207 from /.well-known/caldav, got %d. Body:\n%s", rec.Code, rec.Body.String())
	})

	t.Run("well-known/caldav/ with trailing slash responds", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavRequest(t, e, "PROPFIND", "/.well-known/caldav/", PropfindCurrentUserPrincipal, map[string]string{
			"Depth": "0",
		})

		assert.True(t,
			rec.Code == http.StatusMovedPermanently ||
				rec.Code == http.StatusFound ||
				rec.Code == http.StatusMultiStatus,
			"Expected 301, 302, or 207 from /.well-known/caldav/, got %d. Body:\n%s", rec.Code, rec.Body.String())
	})

	t.Run("well-known/caldav without auth returns 401", func(t *testing.T) {
		e := setupTestEnv(t)

		req := httptest.NewRequestWithContext(context.Background(), "PROPFIND", "/.well-known/caldav", strings.NewReader(PropfindCurrentUserPrincipal))
		req.Header.Set("Content-Type", "application/xml; charset=utf-8")
		req.Header.Set("Depth", "0")

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code,
			"CalDAV well-known endpoint should require authentication")
	})
}

func TestDiscoveryPrincipal(t *testing.T) {
	// RFC 5397 §3 (rfc5397.txt line 126):
	// "This property contains a URL that identifies the principal resource
	//  corresponding to the currently authenticated user."

	t.Run("PROPFIND on /dav/ returns current-user-principal", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/", "0", PropfindCurrentUserPrincipal)

		// Should get 207 Multi-Status
		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		assert.NotEmpty(t, ms.Responses, "Multistatus should contain at least one response")

		// The current-user-principal should point to a principal resource
		// containing the username
		body := rec.Body.String()
		assert.Contains(t, body, "current-user-principal",
			"Response should contain current-user-principal property")
		// Should contain the username in the principal URL
		assert.Contains(t, body, "user15",
			"Principal URL should contain the authenticated username")
	})

	t.Run("PROPFIND on /dav/principals/user15/ returns principal info", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/principals/user15/", "0", PropfindCalendarHomeSet)

		assertResponseStatus(t, rec, 207)

		body := rec.Body.String()
		// Per RFC 4791 §6.2.1, the principal should advertise calendar-home-set
		assert.Contains(t, body, "calendar-home-set",
			"Principal resource should include calendar-home-set property")
		// The home set should point to /dav/projects
		assert.Contains(t, body, "/dav/projects",
			"calendar-home-set should point to /dav/projects")
	})
}

func TestDiscoveryCalendarHome(t *testing.T) {
	// RFC 4791 §6.2.1 (rfc4791.txt line 1651):
	// "The calendar-home-set property identifies the URL of any
	//  WebDAV collections that contain calendar collections owned
	//  by the associated principal resource."

	t.Run("PROPFIND Depth:1 on /dav/projects lists calendars", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/projects", "1", PropfindCalendarCollectionProperties)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		// testuser15 owns projects 36 and 38 (from fixtures)
		// The response should include at least these projects
		assert.GreaterOrEqual(t, len(ms.Responses), 2,
			"Should list at least the 2 projects owned by testuser15")

		// Each response should have an href and a displayname
		for _, r := range ms.Responses {
			assert.NotEmpty(t, r.Href, "Each calendar response should have an href")
		}

		body := rec.Body.String()
		// Check that the projects we know about are listed
		assert.Contains(t, body, "Project 36 for Caldav tests",
			"Should list Project 36")
		assert.Contains(t, body, "Project 38 for Caldav tests",
			"Should list Project 38")
	})

	t.Run("PROPFIND Depth:0 on /dav/projects returns just the home collection", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/projects", "0", PropfindCalendarCollectionProperties)

		assertResponseStatus(t, rec, 207)
		ms := parseMultistatus(t, rec)

		// Depth 0 should return just the collection itself, not children
		assert.Len(t, ms.Responses, 1,
			"Depth 0 PROPFIND should return only the collection itself")
	})

	t.Run("Each listed calendar has resourcetype with calendar", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/projects", "1", PropfindCalendarCollectionProperties)

		assertResponseStatus(t, rec, 207)
		body := rec.Body.String()

		// Per RFC 4791 §5.2, calendar collections MUST report
		// DAV:collection and CALDAV:calendar in resourcetype
		assert.Contains(t, body, "calendar",
			"Calendar collections should have calendar in resourcetype")
	})

	t.Run("Each listed calendar has supported-calendar-component-set", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavPROPFIND(t, e, "/dav/projects", "1", PropfindCalendarCollectionProperties)

		assertResponseStatus(t, rec, 207)
		body := rec.Body.String()

		// Per RFC 4791 §5.2.3 (rfc4791.txt line 768), calendar collections
		// SHOULD report supported-calendar-component-set
		assert.Contains(t, body, "VTODO",
			"supported-calendar-component-set should include VTODO")
	})
}

func TestDiscoveryOPTIONS(t *testing.T) {
	// RFC 4791 §5.1 (rfc4791.txt line 602):
	// "A CalDAV server MUST include 'calendar-access' as a field in the
	//  DAV response header from an OPTIONS request on any resource that
	//  supports the CalDAV extensions."

	t.Run("OPTIONS on /dav/ returns DAV header with calendar-access", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavOPTIONS(t, e, "/dav/")

		assert.Equal(t, http.StatusOK, rec.Code)

		davHeader := rec.Header().Get("DAV")
		assert.NotEmpty(t, davHeader, "OPTIONS response should include DAV header")
		assert.Contains(t, davHeader, "calendar-access",
			"DAV header should include 'calendar-access' per RFC 4791 §5.1")
	})

	t.Run("OPTIONS on /dav/projects/36 returns DAV header", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavOPTIONS(t, e, "/dav/projects/36")

		assert.Equal(t, http.StatusOK, rec.Code)

		davHeader := rec.Header().Get("DAV")
		assert.NotEmpty(t, davHeader, "OPTIONS response should include DAV header")
	})

	t.Run("OPTIONS on /dav/ returns Allow header with supported methods", func(t *testing.T) {
		e := setupTestEnv(t)

		rec := caldavOPTIONS(t, e, "/dav/")

		allowHeader := rec.Header().Get("Allow")
		// A CalDAV server should advertise at least these methods
		for _, method := range []string{"OPTIONS", "GET", "PUT", "DELETE", "PROPFIND", "REPORT"} {
			assert.Contains(t, allowHeader, method,
				"Allow header should include %s", method)
		}
	})
}

func TestDiscoveryFullChain(t *testing.T) {
	// RFC 6764 §6 (rfc6764.txt line 254) describes the full bootstrapping flow:
	// 1. Client does PROPFIND on /.well-known/caldav (or follows redirect)
	// 2. Client extracts current-user-principal from response
	// 3. Client does PROPFIND on principal URL for calendar-home-set
	// 4. Client does PROPFIND Depth:1 on calendar-home-set to list calendars

	t.Run("Full discovery chain: well-known -> principal -> home -> calendars", func(t *testing.T) {
		e := setupTestEnv(t)

		// Step 1: Hit well-known endpoint
		rec1 := caldavRequest(t, e, "PROPFIND", "/.well-known/caldav", PropfindCurrentUserPrincipal, map[string]string{
			"Depth": "0",
		})
		// Accept either redirect or direct response
		assert.True(t, rec1.Code == 207 || rec1.Code == 301 || rec1.Code == 302,
			"Step 1: /.well-known/caldav should respond with 207, 301, or 302, got %d", rec1.Code)

		// Step 2: PROPFIND the entry point for principal info
		rec2 := caldavPROPFIND(t, e, "/dav/", "0", PropfindCurrentUserPrincipal)
		assertResponseStatus(t, rec2, 207)

		// Step 3: PROPFIND the principal URL for calendar-home-set
		// The principal URL for testuser15 should be /dav/principals/user15/
		rec3 := caldavPROPFIND(t, e, "/dav/principals/user15/", "0", PropfindCalendarHomeSet)
		assertResponseStatus(t, rec3, 207)

		body3 := rec3.Body.String()
		assert.Contains(t, body3, "calendar-home-set",
			"Step 3: Principal should advertise calendar-home-set")
		assert.Contains(t, body3, "/dav/projects",
			"Step 3: calendar-home-set should point to /dav/projects")

		// Step 4: PROPFIND Depth:1 on calendar home to list calendars
		rec4 := caldavPROPFIND(t, e, "/dav/projects", "1", PropfindCalendarCollectionProperties)
		assertResponseStatus(t, rec4, 207)

		ms4 := parseMultistatus(t, rec4)
		assert.GreaterOrEqual(t, len(ms4.Responses), 2,
			"Step 4: Should list at least 2 calendars for testuser15")

		body4 := rec4.Body.String()
		assert.Contains(t, body4, "Project 36 for Caldav tests",
			"Step 4: Should list Project 36")
	})
}
