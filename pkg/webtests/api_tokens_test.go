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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/routes"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type APITokenTestSuite struct {
	suite.Suite
	th *testHelper
}

func (s *APITokenTestSuite) SetupTest() {
	s.th = NewTestHelper(s.T())
	s.th.Login(s.T(), &testuser1)
}

func TestAPITokenTestSuite(t *testing.T) {
	suite.Run(t, new(APITokenTestSuite))
}

func (s *APITokenTestSuite) TestValidToken() {
	e, err := setupTestEnv()
	s.Require().NoError(err)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/all", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h := routes.SetupTokenMiddleware()(func(c echo.Context) error {
		u, err := auth.GetAuthFromClaims(c)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, u)
	})

	req.Header.Set(echo.HeaderAuthorization, "Bearer tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e") // Token 1
	s.Require().NoError(h(c))
	// check if the request handlers "see" the request as if it came directly from that user
	s.Assert().Contains(res.Body.String(), `"username":"user1"`)
}

func (s *APITokenTestSuite) TestInvalidToken() {
	e, err := setupTestEnv()
	s.Require().NoError(err)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/all", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h := routes.SetupTokenMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	req.Header.Set(echo.HeaderAuthorization, "Bearer tk_loremipsumdolorsitamet")
	s.Require().Error(h(c))
}

func (s *APITokenTestSuite) TestExpiredToken() {
	e, err := setupTestEnv()
	s.Require().NoError(err)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/all", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h := routes.SetupTokenMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	req.Header.Set(echo.HeaderAuthorization, "Bearer tk_a5e6f92ddbad68f49ee2c63e52174db0235008c8") // Token 2
	s.Require().Error(h(c))
}

func (s *APITokenTestSuite) TestValidTokenInvalidScope() {
	e, err := setupTestEnv()
	s.Require().NoError(err)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h := routes.SetupTokenMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	req.Header.Set(echo.HeaderAuthorization, "Bearer tk_2eef46f40ebab3304919ab2e7e39993f75f29d2e")
	s.Require().Error(h(c))
}

func (s *APITokenTestSuite) TestJWT() {
	e, err := setupTestEnv()
	s.Require().NoError(err)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/all", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h := routes.SetupTokenMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	})

	sess := db.NewSession()
	defer sess.Close()
	u, err := user.GetUserByID(sess, 1)
	s.Require().NoError(err)
	jwt, err := auth.NewUserJWTAuthtoken(u, false)
	s.Require().NoError(err)

	req.Header.Set(echo.HeaderAuthorization, "Bearer "+jwt)
	s.Require().NoError(h(c))
}

func (s *APITokenTestSuite) TestNonExistingRoute() {
	e, err := setupTestEnv()
	s.Require().NoError(err)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexisting", nil)
	res := httptest.NewRecorder()
	c := e.NewContext(req, res)
	h := routes.SetupTokenMiddleware()(func(c echo.Context) error {
		return c.String(http.StatusNotFound, "test")
	})

	req.Header.Set(echo.HeaderAuthorization, "Bearer tk_a5e6f92ddbad68f49ee2c63e52174db0235008c8") // Token 2

	err = h(c)
	s.Require().NoError(err)
	s.Assert().Equal(404, c.Response().Status)
}

func (s *APITokenTestSuite) createToken(permissions string) models.APIToken {
	res, err := s.th.Request(s.T(), http.MethodPut, "/api/v1/tokens", strings.NewReader(permissions))
	s.Require().NoError(err)
	s.Assert().Equal(http.StatusCreated, res.Code)

	var createdToken models.APIToken
	err = json.NewDecoder(res.Body).Decode(&createdToken)
	s.Require().NoError(err)
	return createdToken
}

func (s *APITokenTestSuite) TestV1TokenV1Route() {
	expiresAt := time.Now().Add(30 * 24 * time.Hour).UTC().Format("2006-01-02T15:04:05.000Z")
	payload := `{"max_permission":null, "id":0, "title":"test-token", "token":"", "permissions":{"v1_projects":["read_all"]},"expires_at":"` + expiresAt + `","created":"1970-01-01T00:00:00.000Z","updated":null}`
	token := s.createToken(payload)

	s.th.token = token.Token
	res, err := s.th.Request(s.T(), http.MethodGet, "/api/v1/projects", nil)
	s.Require().NoError(err)
	s.Assert().Equal(http.StatusOK, res.Code)
}

func (s *APITokenTestSuite) TestV1TokenV2Route() {
	expiresAt := time.Now().Add(30 * 24 * time.Hour).UTC().Format("2006-01-02T15:04:05.000Z")
	payload := `{"max_permission":null, "id":0, "title":"test-token", "token":"", "permissions":{"v1_projects":["read_all"]},"expires_at":"` + expiresAt + `","created":"1970-01-01T00:00:00.000Z","updated":null}`
	token := s.createToken(payload)
	s.th.token = token.Token
	_, err := s.th.Request(s.T(), http.MethodGet, "/api/v2/projects", nil)
	s.Require().Error(err)

}

func (s *APITokenTestSuite) TestV2TokenV2Route() {
	expiresAt := time.Now().Add(30 * 24 * time.Hour).UTC().Format("2006-01-02T15:04:05.000Z")
	payload := `{"max_permission":null, "id":0, "title":"test-token", "token":"", "permissions":{"v2_projects":["read_all"]},"expires_at":"` + expiresAt + `","created":"1970-01-01T00:00:00.000Z","updated":null}`
	token := s.createToken(payload)
	s.th.token = token.Token
	res, err := s.th.Request(s.T(), http.MethodGet, "/api/v2/projects", nil)
	s.Require().NoError(err)
	s.Assert().Equal(http.StatusOK, res.Code)
}

func (s *APITokenTestSuite) TestV2TokenV1Route() {
	expiresAt := time.Now().Add(30 * 24 * time.Hour).UTC().Format("2006-01-02T15:04:05.000Z")
	payload := `{"max_permission":null, "id":0, "title":"test-token", "token":"", "permissions":{"v2_projects":["read_all"]},"expires_at":"` + expiresAt + `","created":"1970-01-01T00:00:00.000Z","updated":null}`
	token := s.createToken(payload)
	s.th.token = token.Token
	_, err := s.th.Request(s.T(), http.MethodGet, "/api/v1/projects", nil)
	s.Require().Error(err)
}

func (s *APITokenTestSuite) TestV1V2TokenV1V2Routes() {
	expiresAt := time.Now().Add(30 * 24 * time.Hour).UTC().Format("2006-01-02T15:04:05.000Z")
	payload := `{"max_permission":null, "id":0, "title":"test-token", "token":"", "permissions":{"v1_projects":["read_all"],"v2_projects":["read_all"]},"expires_at":"` + expiresAt + `","created":"1970-01-01T00:00:00.000Z","updated":null}`
	token := s.createToken(payload)
	s.th.token = token.Token

	res, err := s.th.Request(s.T(), http.MethodGet, "/api/v1/projects", nil)
	s.Require().NoError(err)
	s.Assert().Equal(http.StatusOK, res.Code)

	res, err = s.th.Request(s.T(), http.MethodGet, "/api/v2/projects", nil)
	s.Require().NoError(err)
	s.Assert().Equal(http.StatusOK, res.Code)
}

func (s *APITokenTestSuite) TestInvalidScope() {
	res, err := s.th.Request(s.T(), http.MethodPut, "/api/v1/tokens", strings.NewReader(`{"title":"test-token", "api_permissions": {"v3_projects": ["read_all"]}}`))
	s.Require().Error(err)
	s.Assert().Equal(http.StatusBadRequest, res.Code)
}
