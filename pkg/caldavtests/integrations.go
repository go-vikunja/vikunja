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
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

// These are the test users, the same way they are in the test database
var (
	testuser1 = user.User{
		ID:       1,
		Username: "user1",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user1@example.com",
		Issuer:   "local",
	}
	testuser15 = user.User{
		ID:       15,
		Username: "user15",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user15@example.com",
		Issuer:   "local",
	}
)

// fixturePassword is the plaintext password for all test fixture users
const fixturePassword = "12345678"

func setupTestEnv(t *testing.T) *echo.Echo {
	t.Helper()

	config.InitDefaultConfig()
	config.ServicePublicURL.Set("https://localhost")

	log.InitLogger()
	files.InitTests()
	user.InitTests()
	models.SetupTests()
	events.Fake()
	keyvalue.InitStorage()

	err := db.LoadFixtures()
	require.NoError(t, err)

	e := routes.NewEcho()
	routes.RegisterRoutes(e)
	return e
}

// basicAuthHeader returns the Authorization header value for HTTP Basic Auth.
func basicAuthHeader(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

// caldavRequest sends an HTTP request through the full Echo router and returns the response.
func caldavRequest(t *testing.T, e *echo.Echo, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")

	// Default to testuser15 basic auth (the caldav test user) unless overridden
	if _, hasAuth := headers["Authorization"]; !hasAuth {
		req.Header.Set("Authorization", basicAuthHeader(testuser15.Username, fixturePassword))
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

// caldavPROPFIND sends a PROPFIND request.
func caldavPROPFIND(t *testing.T, e *echo.Echo, path, depth, body string) *httptest.ResponseRecorder {
	t.Helper()
	return caldavRequest(t, e, "PROPFIND", path, body, map[string]string{
		"Depth": depth,
	})
}

// caldavREPORT sends a REPORT request.
func caldavREPORT(t *testing.T, e *echo.Echo, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	return caldavRequest(t, e, "REPORT", path, body, nil)
}

// caldavGET sends a GET request.
func caldavGET(t *testing.T, e *echo.Echo, path string) *httptest.ResponseRecorder {
	t.Helper()
	return caldavRequest(t, e, http.MethodGet, path, "", nil)
}

// caldavPUT sends a PUT request with iCalendar content.
func caldavPUT(t *testing.T, e *echo.Echo, path, vcalendar string) *httptest.ResponseRecorder {
	t.Helper()
	return caldavRequest(t, e, http.MethodPut, path, vcalendar, map[string]string{
		"Content-Type": "text/calendar; charset=utf-8",
	})
}

// caldavDELETE sends a DELETE request.
func caldavDELETE(t *testing.T, e *echo.Echo, path string) *httptest.ResponseRecorder {
	t.Helper()
	return caldavRequest(t, e, http.MethodDelete, path, "", nil)
}

// caldavOPTIONS sends an OPTIONS request.
func caldavOPTIONS(t *testing.T, e *echo.Echo, path string) *httptest.ResponseRecorder {
	t.Helper()
	return caldavRequest(t, e, http.MethodOptions, path, "", nil)
}

// caldavRequestAsUser sends a request authenticated as a specific user.
func caldavRequestAsUser(t *testing.T, e *echo.Echo, method, path, body string, u *user.User, password string) *httptest.ResponseRecorder {
	t.Helper()
	return caldavRequest(t, e, method, path, body, map[string]string{
		"Authorization": basicAuthHeader(u.Username, password),
	})
}
