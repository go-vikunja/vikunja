// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package integrations

import (
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
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
	"code.vikunja.io/web/handler"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// These are the test users, the same way they are in the test database
var (
	testuser1 = user.User{
		ID:       1,
		Username: "user1",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user1@example.com",
	}
	testuser2 = user.User{
		ID:       2,
		Username: "user2",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user2@example.com",
	}
	testuser3 = user.User{
		ID:       3,
		Username: "user3",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user3@example.com",
	}
	testuser4 = user.User{
		ID:       4,
		Username: "user4",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user4@example.com",
	}
	testuser5 = user.User{
		ID:       4,
		Username: "user5",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user5@example.com",
		Status:   user.StatusDisabled,
	}
)

func setupTestEnv() (e *echo.Echo, err error) {
	config.InitDefaultConfig()
	// We need to set the root path even if we're not using the config, otherwise fixtures are not loaded correctly
	config.ServiceRootpath.Set(os.Getenv("VIKUNJA_SERVICE_ROOTPATH"))
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

func bootstrapTestRequest(t *testing.T, method string, payload string, queryParam url.Values) (c echo.Context, rec *httptest.ResponseRecorder) {
	// Setup
	e, err := setupTestEnv()
	assert.NoError(t, err)

	// Do the actual request
	req := httptest.NewRequest(method, "/", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.URL.RawQuery = queryParam.Encode()
	rec = httptest.NewRecorder()

	c = e.NewContext(req, rec)
	return
}

func newTestRequest(t *testing.T, method string, handler func(ctx echo.Context) error, payload string, queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	rec, c := testRequestSetup(t, method, payload, queryParams, urlParams)
	err = handler(c)
	return
}

func addUserTokenToContext(t *testing.T, user *user.User, c echo.Context) {
	// Get the token as a string
	token, err := auth.NewUserJWTAuthtoken(user, false)
	assert.NoError(t, err)
	// We send the string token through the parsing function to get a valid jwt.Token
	tken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.ServiceJWTSecret.GetString()), nil
	})
	assert.NoError(t, err)
	c.Set("user", tken)
}

func addLinkShareTokenToContext(t *testing.T, share *models.LinkSharing, c echo.Context) {
	// Get the token as a string
	token, err := auth.NewLinkShareJWTAuthtoken(share)
	assert.NoError(t, err)
	// We send the string token through the parsing function to get a valid jwt.Token
	tken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.ServiceJWTSecret.GetString()), nil
	})
	assert.NoError(t, err)
	c.Set("user", tken)
}

func testRequestSetup(t *testing.T, method string, payload string, queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, c echo.Context) {
	c, rec = bootstrapTestRequest(t, method, payload, queryParams)

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

func newTestRequestWithUser(t *testing.T, method string, handler echo.HandlerFunc, user *user.User, payload string, queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	rec, c := testRequestSetup(t, method, payload, queryParams, urlParams)
	addUserTokenToContext(t, user, c)
	err = handler(c)
	return
}

func newTestRequestWithLinkShare(t *testing.T, method string, handler echo.HandlerFunc, share *models.LinkSharing, payload string, queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	rec, c := testRequestSetup(t, method, payload, queryParams, urlParams)
	addLinkShareTokenToContext(t, share, c)
	err = handler(c)
	return
}

func assertHandlerErrorCode(t *testing.T, err error, expectedErrorCode int) {
	if err == nil {
		t.Error("Error is nil")
		t.FailNow()
	}
	httperr, ok := err.(*echo.HTTPError)
	if !ok {
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
