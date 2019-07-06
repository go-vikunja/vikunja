//   Vikunja is a todo-list application to facilitate your life.
//   Copyright 2019 Vikunja and contributors. All rights reserved.
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <https://www.gnu.org/licenses/>.

package integrations

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes"
	v1 "code.vikunja.io/api/pkg/routes/api/v1"
	"code.vikunja.io/web"
	"code.vikunja.io/web/handler"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// These are the test users, the same way they are in the test database
var (
	testuser1 = models.User{
		ID:       1,
		Username: "user1",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user1@example.com",
		IsActive: true,
	}
	testuser2 = models.User{
		ID:       2,
		Username: "user2",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user2@example.com",
	}
	testuser3 = models.User{
		ID:                 3,
		Username:           "user3",
		Password:           "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:              "user3@example.com",
		PasswordResetToken: "passwordresettesttoken",
	}
	testuser4 = models.User{
		ID:                4,
		Username:          "user4",
		Password:          "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:             "user4@example.com",
		EmailConfirmToken: "tiepiQueed8ahc7zeeFe1eveiy4Ein8osooxegiephauph2Ael",
	}
	testuser5 = models.User{
		ID:                4,
		Username:          "user5",
		Password:          "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:             "user5@example.com",
		EmailConfirmToken: "tiepiQueed8ahc7zeeFe1eveiy4Ein8osooxegiephauph2Ael",
		IsActive:          false,
	}
)

func setupTestEnv() (e *echo.Echo, err error) {
	config.InitConfig()
	models.SetupTests(config.ServiceRootpath.GetString())

	err = models.LoadFixtures()
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

func newTestRequest(t *testing.T, method string, handler func(ctx echo.Context) error, payload string) (rec *httptest.ResponseRecorder, err error) {
	c, rec := bootstrapTestRequest(t, method, payload, nil)
	err = handler(c)
	return
}

func addTokenToContext(t *testing.T, user *models.User, c echo.Context) {
	// Get the token as a string
	token, err := v1.CreateNewJWTTokenForUser(user)
	assert.NoError(t, err)
	// We send the string token through the parsing function to get a valid jwt.Token
	tken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.ServiceJWTSecret.GetString()), nil
	})
	assert.NoError(t, err)
	c.Set("user", tken)
}

func newTestRequestWithUser(t *testing.T, method string, handler echo.HandlerFunc, user *models.User, payload string, queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	c, rec := bootstrapTestRequest(t, method, payload, queryParams)

	var paramNames []string
	var paramValues []string
	for name, value := range urlParams {
		paramNames = append(paramNames, name)
		paramValues = append(paramValues, value)
	}
	c.SetParamNames(paramNames...)
	c.SetParamValues(paramValues...)

	addTokenToContext(t, user, c)
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
	user    *models.User
	strFunc func() handler.CObject
	t       *testing.T
}

func (h *webHandlerTest) getHandler() handler.WebHandler {
	return handler.WebHandler{
		EmptyStruct: func() handler.CObject {
			return h.strFunc()
		},
	}
}

func (h *webHandlerTest) testReadAll(queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	hndl := h.getHandler()
	return newTestRequestWithUser(h.t, http.MethodGet, hndl.ReadAllWeb, h.user, "", queryParams, urlParams)
}

func (h *webHandlerTest) testReadOne(queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	hndl := h.getHandler()
	return newTestRequestWithUser(h.t, http.MethodGet, hndl.ReadOneWeb, h.user, "", queryParams, urlParams)
}

func (h *webHandlerTest) testCreate(queryParams url.Values, urlParams map[string]string, payload string) (rec *httptest.ResponseRecorder, err error) {
	hndl := h.getHandler()
	return newTestRequestWithUser(h.t, http.MethodPut, hndl.CreateWeb, h.user, payload, queryParams, urlParams)
}

func (h *webHandlerTest) testUpdate(queryParams url.Values, urlParams map[string]string, payload string) (rec *httptest.ResponseRecorder, err error) {
	hndl := h.getHandler()
	return newTestRequestWithUser(h.t, http.MethodPost, hndl.UpdateWeb, h.user, payload, queryParams, urlParams)
}

func (h *webHandlerTest) testDelete(queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	hndl := h.getHandler()
	return newTestRequestWithUser(h.t, http.MethodDelete, hndl.DeleteWeb, h.user, "", queryParams, urlParams)
}
