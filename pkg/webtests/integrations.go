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
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	"github.com/labstack/echo/v5"
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
		Issuer:   "local",
	}
	testuser10 = user.User{
		ID:       10,
		Username: "user10",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user10@example.com",
		Issuer:   "local",
	}
	testuser15 = user.User{
		ID:       15,
		Username: "user15",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user15@example.com",
		Issuer:   "local",
	}
	testuser6 = user.User{
		ID:       6,
		Username: "user6",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		Email:    "user6@example.com",
		Issuer:   "local",
	}
)

func setupTestEnv() (e *echo.Echo, err error) {
	config.InitDefaultConfig()
	config.ServicePublicURL.Set("https://localhost")

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

func createRequest(e *echo.Echo, method string, payload string, queryParam url.Values, urlParams map[string]string) (c *echo.Context, rec *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, "/", strings.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.URL.RawQuery = queryParam.Encode()
	rec = httptest.NewRecorder()

	c = e.NewContext(req, rec)
	// In Echo v5, we use SetPathValues to set path parameters
	// Only set path values if there are any, as SetPathValues panics with nil
	if len(urlParams) > 0 {
		pathValues := make(echo.PathValues, 0, len(urlParams))
		for name, value := range urlParams {
			pathValues = append(pathValues, echo.PathValue{Name: name, Value: value})
		}
		c.SetPathValues(pathValues)
	}
	return
}

func bootstrapTestRequest(t *testing.T, method string, payload string, queryParam url.Values, urlParams map[string]string) (c *echo.Context, rec *httptest.ResponseRecorder) {
	// Setup
	e, err := setupTestEnv()
	require.NoError(t, err)

	c, rec = createRequest(e, method, payload, queryParam, urlParams)
	return
}

func newTestRequest(t *testing.T, method string, handler func(ctx *echo.Context) error, payload string, queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	var c *echo.Context
	c, rec = bootstrapTestRequest(t, method, payload, queryParams, urlParams)
	err = handler(c)
	return
}

func addUserTokenToContext(t *testing.T, user *user.User, c *echo.Context) {
	// Get the token as a string
	token, err := auth.NewUserJWTAuthtoken(user, "test-session-id")
	require.NoError(t, err)
	// We send the string token through the parsing function to get a valid jwt.Token
	tken, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return []byte(config.ServiceSecret.GetString()), nil
	})
	require.NoError(t, err)
	c.Set("user", tken)
}

func addLinkShareTokenToContext(t *testing.T, share *models.LinkSharing, c *echo.Context) {
	// Get the token as a string
	token, err := auth.NewLinkShareJWTAuthtoken(share)
	require.NoError(t, err)
	// We send the string token through the parsing function to get a valid jwt.Token
	tken, err := jwt.Parse(token, func(_ *jwt.Token) (interface{}, error) {
		return []byte(config.ServiceSecret.GetString()), nil
	})
	require.NoError(t, err)
	c.Set("user", tken)
}

func newTestRequestWithUser(t *testing.T, method string, handler echo.HandlerFunc, user *user.User, payload string, queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	var c *echo.Context
	c, rec = bootstrapTestRequest(t, method, payload, queryParams, urlParams)
	addUserTokenToContext(t, user, c)
	err = handler(c)
	return
}

func newTestRequestWithLinkShare(t *testing.T, method string, handler echo.HandlerFunc, share *models.LinkSharing, payload string, queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	var c *echo.Context
	c, rec = bootstrapTestRequest(t, method, payload, queryParams, urlParams)
	addLinkShareTokenToContext(t, share, c)
	err = handler(c)
	return
}

func newCaldavTestRequestWithUser(t *testing.T, e *echo.Echo, method string, handler echo.HandlerFunc, user *user.User, payload string, queryParams url.Values, urlParams map[string]string) (rec *httptest.ResponseRecorder, err error) {
	var c *echo.Context
	c, rec = createRequest(e, method, payload, queryParams, urlParams)
	c.Request().Header.Set(echo.HeaderContentType, echo.MIMETextPlain)

	result, _ := caldav.BasicAuth(c, user.Username, "12345678")
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

	// First, try to get error code from HTTPErrorProcessor (domain errors like ValidationHTTPError)
	if httpErr, ok := err.(web.HTTPErrorProcessor); ok {
		assert.Equal(t, expectedErrorCode, httpErr.HTTPError().Code)
		return
	}

	// Try to unwrap to find HTTPErrorProcessor
	unwrapped := errors.Unwrap(err)
	for unwrapped != nil {
		if httpErr, ok := unwrapped.(web.HTTPErrorProcessor); ok {
			assert.Equal(t, expectedErrorCode, httpErr.HTTPError().Code)
			return
		}
		unwrapped = errors.Unwrap(unwrapped)
	}

	// Fall back to echo.HTTPError for middleware/auth errors
	var httperr *echo.HTTPError
	if !errors.As(err, &httperr) {
		t.Errorf("Error is not *echo.HTTPError or web.HTTPErrorProcessor: %T", err)
		t.FailNow()
	}

	// In Echo v5, HTTPError.Message is a string, not interface{}
	// The internal error might contain our web.HTTPError
	if innerErr := httperr.Unwrap(); innerErr != nil {
		if httpErr, ok := innerErr.(web.HTTPErrorProcessor); ok {
			assert.Equal(t, expectedErrorCode, httpErr.HTTPError().Code)
			return
		}
	}

	t.Errorf("Could not extract error code from error: %T - %v", err, err)
	t.FailNow()
}

// httpCodeGetter is an interface for errors that can provide their HTTP status code.
type httpCodeGetter interface {
	GetHTTPCode() int
}

// getHTTPErrorCode extracts the HTTP status code from various error types
func getHTTPErrorCode(err error) int {
	// First, try domain errors that implement HTTPErrorProcessor
	if httpErr, ok := err.(web.HTTPErrorProcessor); ok {
		return httpErr.HTTPError().HTTPCode
	}

	// Try errors that implement httpCodeGetter (like ValidationHTTPError)
	if codeGetter, ok := err.(httpCodeGetter); ok {
		return codeGetter.GetHTTPCode()
	}

	// Fall back to echo.HTTPError
	var httperr *echo.HTTPError
	if errors.As(err, &httperr) {
		return httperr.Code
	}

	return 0
}

// getHTTPErrorMessage extracts the message from various error types
func getHTTPErrorMessage(err error) interface{} {
	// First, try domain errors that implement HTTPErrorProcessor
	if httpErr, ok := err.(web.HTTPErrorProcessor); ok {
		return httpErr.HTTPError().Message
	}

	// Then try echo.HTTPError (for Forbidden etc.)
	var httperr *echo.HTTPError
	if errors.As(err, &httperr) {
		return httperr.Message
	}

	// Fall back to error string
	return err.Error()
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

// webHandlerTestV2 mirrors webHandlerTest's signatures but dispatches
// through the full Echo+Huma stack, so v2 tests read side-by-side with v1.
// urlParams keys match v1 so the same map can be reused.
type webHandlerTestV2 struct {
	user     *user.User
	basePath string
	idParam  string // matches v1 urlParams keys, e.g. "label"
	t        *testing.T
	e        *echo.Echo
}

// v2HTTPError implements web.HTTPErrorProcessor so existing
// getHTTPErrorCode / assertHandlerErrorCode helpers work against v2.
type v2HTTPError struct {
	httpCode int
	code     int
	message  string
}

func (e *v2HTTPError) Error() string {
	return e.message
}

func (e *v2HTTPError) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: e.httpCode,
		Code:     e.code,
		Message:  e.message,
	}
}

// v2ProblemJSON is the subset of the RFC 9457 body the harness reads.
type v2ProblemJSON struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
	// Domain errors with web.HTTPErrorProcessor carry a numeric code; 0 otherwise.
	Code int `json:"code"`
}

// newV2Error wraps a >=400 recorder so v1-style assertions keep working.
// Non-JSON / non-problem bodies fall back to the raw body string.
func newV2Error(rec *httptest.ResponseRecorder) error {
	msg := strings.TrimSpace(rec.Body.String())
	var body v2ProblemJSON
	if jsonErr := json.Unmarshal(rec.Body.Bytes(), &body); jsonErr == nil {
		if body.Detail != "" {
			msg = body.Detail
		} else if body.Title != "" {
			msg = body.Title
		}
	}
	return &v2HTTPError{
		httpCode: rec.Code,
		code:     body.Code,
		message:  msg,
	}
}

func (h *webHandlerTestV2) ensureEnv() error {
	if h.e != nil {
		return nil
	}
	e, err := setupTestEnv()
	if err != nil {
		return err
	}
	h.e = e
	return nil
}

// buildURL assembles basePath[/{id}]?query using the idParam lookup.
func (h *webHandlerTestV2) buildURL(queryParams url.Values, urlParams map[string]string, withID bool) string {
	u := h.basePath
	if withID {
		id := ""
		if h.idParam != "" {
			id = urlParams[h.idParam]
		}
		if id == "" {
			// Fallback for tests that pass a differently-named key or omit idParam.
			for _, v := range urlParams {
				id = v
				break
			}
		}
		u += "/" + id
	}
	if q := queryParams.Encode(); q != "" {
		u += "?" + q
	}
	return u
}

func (h *webHandlerTestV2) serve(method, path, payload string) (*httptest.ResponseRecorder, error) {
	require.NoError(h.t, h.ensureEnv())
	token, err := auth.NewUserJWTAuthtoken(h.user, "test-session-id")
	require.NoError(h.t, err)
	var reader *strings.Reader
	if payload != "" {
		reader = strings.NewReader(payload)
	} else {
		reader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	h.e.ServeHTTP(rec, req)
	if rec.Code >= 400 {
		return rec, newV2Error(rec)
	}
	return rec, nil
}

func (h *webHandlerTestV2) testReadAllWithUser(queryParams url.Values, urlParams map[string]string) (*httptest.ResponseRecorder, error) {
	return h.serve(http.MethodGet, h.buildURL(queryParams, urlParams, false), "")
}

func (h *webHandlerTestV2) testReadOneWithUser(queryParams url.Values, urlParams map[string]string) (*httptest.ResponseRecorder, error) {
	return h.serve(http.MethodGet, h.buildURL(queryParams, urlParams, true), "")
}

// v2 uses POST for create; otherwise identical to v1's testCreateWithUser.
func (h *webHandlerTestV2) testCreateWithUser(queryParams url.Values, urlParams map[string]string, payload string) (*httptest.ResponseRecorder, error) {
	return h.serve(http.MethodPost, h.buildURL(queryParams, urlParams, false), payload)
}

func (h *webHandlerTestV2) testUpdateWithUser(queryParams url.Values, urlParams map[string]string, payload string) (*httptest.ResponseRecorder, error) {
	return h.serve(http.MethodPut, h.buildURL(queryParams, urlParams, true), payload)
}

func (h *webHandlerTestV2) testDeleteWithUser(queryParams url.Values, urlParams map[string]string, payload ...string) (*httptest.ResponseRecorder, error) {
	pl := ""
	if len(payload) > 0 {
		pl = payload[0]
	}
	return h.serve(http.MethodDelete, h.buildURL(queryParams, urlParams, true), pl)
}
