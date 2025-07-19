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

package webtests

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/routes/caldav"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These are the test users, the same way they are in the test database
var (
	testuser1 = user.User{
		ID:       1,
		Username: "user1",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user1@example.com",
	}
	testuser15 = user.User{
		ID:       15,
		Username: "user15",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user15@example.com",
	}
)

func setupTestEnv() (e *echo.Echo, err error) {
	config.InitDefaultConfig()
	// We need to set the root path even if we're not using the config, otherwise fixtures are not loaded correctly
	config.ServiceRootpath.Set(os.Getenv("VIKUNJA_SERVICE_ROOTPATH"))

	// Initialize logger for tests
	log.InitLogger()

	// Some tests use the file engine, so we'll need to initialize that
	files.InitTests()
	user.InitTests()
	models.SetupTests()
	events.Fake()
	keyvalue.InitStorage()

	err = db.LoadFixtures()
	if err != nil {
		return
	}

	e = routes.NewEcho()
	routes.RegisterRoutes(e)
	return
}

func createRequest(e *echo.Echo, method string, payload string, queryParam url.Values, urlParams map[string]string) (c echo.Context, rec *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, "/", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.URL.RawQuery = queryParam.Encode()
	rec = httptest.NewRecorder()

	c = e.NewContext(req, rec)
	var paramNames []string
	var paramValues []string
	for name, value := range urlParams {
		paramNames = append(paramNames, name)
		paramValues = append(paramValues, value)
	}
	c.SetParamNames(paramNames...)
	c.SetParamValues(paramValues...)
	return
}

func bootstrapTestRequest(t *testing.T, method string, payload string, queryParam url.Values, urlParams map[string]string) (c echo.Context, rec *httptest.ResponseRecorder) {
	// Setup
	e, err := setupTestEnv()
	require.NoError(t, err)

	c, rec = createRequest(e, method, payload, queryParam, urlParams)
	return
}

func newTestRequest(t *testing.T, method string, handler func(ctx echo.Context) error, payload string, queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	var c echo.Context
	c, rec = bootstrapTestRequest(t, method, payload, queryParams, urlParams)
	err = handler(c)
	return
}

func addUserTokenToContext(t *testing.T, user *user.User, c echo.Context) {
	// Get the token as a string
	token, err := auth.NewUserJWTAuthtoken(user, false)
	require.NoError(t, err)
	// We send the string token through the parsing function to get a valid jwt.Token
	tken, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return []byte(config.ServiceJWTSecret.GetString()), nil
	})
	require.NoError(t, err)
	c.Set("user", tken)
}

func addLinkShareTokenToContext(t *testing.T, share *models.LinkSharing, c echo.Context) {
	// Get the token as a string
	token, err := auth.NewLinkShareJWTAuthtoken(share)
	require.NoError(t, err)
	// We send the string token through the parsing function to get a valid jwt.Token
	tken, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return []byte(config.ServiceJWTSecret.GetString()), nil
	})
	require.NoError(t, err)
	c.Set("user", tken)
}

func newTestRequestWithUser(t *testing.T, method string, handler echo.HandlerFunc, user *user.User, payload string, queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	var c echo.Context
	c, rec = bootstrapTestRequest(t, method, payload, queryParams, urlParams)
	addUserTokenToContext(t, user, c)
	err = handler(c)
	return
}

func newTestRequestWithLinkShare(t *testing.T, method string, handler echo.HandlerFunc, share *models.LinkSharing, payload string, queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	var c echo.Context
	c, rec = bootstrapTestRequest(t, method, payload, queryParams, urlParams)
	addLinkShareTokenToContext(t, share, c)
	err = handler(c)
	return
}

func newCaldavTestRequestWithUser(t *testing.T, e *echo.Echo, method string, handler echo.HandlerFunc, user *user.User, payload string, queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	var c echo.Context
	c, rec = createRequest(e, method, payload, queryParams, urlParams)
	c.Request().Header.Set(echo.HeaderContentType, echo.MIMETextPlain)

	result, _ := caldav.BasicAuth(user.Username, "12345678", c)
	if !result {
		t.Error("BasicAuth for caldav failed")
		t.FailNow()
	}
	err = handler(c)
	return
}

func assertHandlerErrorCode(t *testing.T, err error, expectedErrorCode int) {
	if err == nil {
		t.Error("Error is nil")
		t.FailNow()
	}
	var httperr *echo.HTTPError
	if !errors.As(err, &httperr) {
		t.Error("Error is not *echo.HTTPError")
		t.FailNow()
	}
	webhttperr, ok := httperr.Message.(web.HTTPError)
	if !ok {
		t.Error("Error is not *web.HTTPError")
		t.FailNow()
	}
	assert.Equal(t, expectedErrorCode, webhttperr.Code)
}

type webHandlerTest struct {
	user      *user.User
	linkShare *models.LinkSharing
	strFunc   func() handler.CObject
	t         *testing.T
}

func (h *webHandlerTest) getHandler() handler.WebHandler {
	return handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return h.strFunc()
		},
	}
}

func (h *webHandlerTest) testReadAllWithUser(queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	hndl := h.getHandler()
	return newTestRequestWithUser(h.t, http.MethodGet, hndl.ReadAllWeb, h.user, "", queryParams, urlParams)
}

func (h *webHandlerTest) testReadOneWithUser(queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	hndl := h.getHandler()
	return newTestRequestWithUser(h.t, http.MethodGet, hndl.ReadOneWeb, h.user, "", queryParams, urlParams)
}

func (h *webHandlerTest) testCreateWithUser(queryParams url.Values, urlParams map[string]string, payload string) (rec *httptest.ResponseRecorder, err error) {
	hndl := h.getHandler()
	return newTestRequestWithUser(h.t, http.MethodPut, hndl.CreateWeb, h.user, payload, queryParams, urlParams)
}

func (h *webHandlerTest) testUpdateWithUser(queryParams url.Values, urlParams map[string]string, payload string) (rec *httptest.ResponseRecorder, err error) {
	hndl := h.getHandler()
	return newTestRequestWithUser(h.t, http.MethodPost, hndl.UpdateWeb, h.user, payload, queryParams, urlParams)
}

func (h *webHandlerTest) testDeleteWithUser(queryParams url.Values, urlParams map[string]string, payload ...string) (rec *httptest.ResponseRecorder, err error) {
	pl := ""
	if len(payload) > 0 {
		pl = payload[0]
	}
	hndl := h.getHandler()
	return newTestRequestWithUser(h.t, http.MethodDelete, hndl.DeleteWeb, h.user, pl, queryParams, urlParams)
}

func (h *webHandlerTest) testReadAllWithLinkShare(queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	hndl := h.getHandler()
	return newTestRequestWithLinkShare(h.t, http.MethodGet, hndl.ReadAllWeb, h.linkShare, "", queryParams, urlParams)
}

func (h *webHandlerTest) testReadOneWithLinkShare(queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	hndl := h.getHandler()
	return newTestRequestWithLinkShare(h.t, http.MethodGet, hndl.ReadOneWeb, h.linkShare, "", queryParams, urlParams)
}

func (h *webHandlerTest) testCreateWithLinkShare(queryParams url.Values, urlParams map[string]string, payload string) (rec *httptest.ResponseRecorder, err error) {
	hndl := h.getHandler()
	return newTestRequestWithLinkShare(h.t, http.MethodPut, hndl.CreateWeb, h.linkShare, payload, queryParams, urlParams)
}

func (h *webHandlerTest) testUpdateWithLinkShare(queryParams url.Values, urlParams map[string]string, payload string) (rec *httptest.ResponseRecorder, err error) {
	hndl := h.getHandler()
	return newTestRequestWithLinkShare(h.t, http.MethodPost, hndl.UpdateWeb, h.linkShare, payload, queryParams, urlParams)
}

func (h *webHandlerTest) testDeleteWithLinkShare(queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	hndl := h.getHandler()
	return newTestRequestWithLinkShare(h.t, http.MethodDelete, hndl.DeleteWeb, h.linkShare, "", queryParams, urlParams)
}
