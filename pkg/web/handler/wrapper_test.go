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

package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"xorm.io/xorm"
)

func TestWithDBAndUser(t *testing.T) {
	t.Run("should call handler with session and user", func(t *testing.T) {
		// Create a test handler that verifies it receives the session and user
		var receivedSession *xorm.Session
		var receivedUser *user.User
		var receivedContext echo.Context

		testHandler := func(s *xorm.Session, u *user.User, c echo.Context) error {
			receivedSession = s
			receivedUser = u
			receivedContext = c
			return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
		}

		// Create Echo instance and request
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Mock user in context (simulating authentication middleware)
		testUser := &user.User{ID: 1, Username: "testuser"}
		c.Set("api_user", testUser)

		// Execute the wrapped handler
		wrappedHandler := WithDBAndUser(testHandler, false)
		err := wrappedHandler(c)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, receivedSession)
		assert.NotNil(t, receivedUser)
		assert.NotNil(t, receivedContext)
		assert.Equal(t, testUser.ID, receivedUser.ID)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "ok")
	})

	t.Run("should handle transaction commit for write operations", func(t *testing.T) {
		testHandler := func(s *xorm.Session, u *user.User, c echo.Context) error {
			// Simulate a write operation
			return c.JSON(http.StatusCreated, map[string]string{"status": "created"})
		}

		// Create Echo instance and request
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Mock user in context
		testUser := &user.User{ID: 1, Username: "testuser"}
		c.Set("api_user", testUser)

		// Execute the wrapped handler with transaction enabled
		wrappedHandler := WithDBAndUser(testHandler, true)
		err := wrappedHandler(c)

		// Assertions
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), "created")
	})

	t.Run("should handle errors from business logic", func(t *testing.T) {
		testHandler := func(s *xorm.Session, u *user.User, c echo.Context) error {
			return echo.NewHTTPError(http.StatusBadRequest, "test error")
		}

		// Create Echo instance and request
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Mock user in context
		testUser := &user.User{ID: 1, Username: "testuser"}
		c.Set("api_user", testUser)

		// Execute the wrapped handler
		wrappedHandler := WithDBAndUser(testHandler, false)
		err := wrappedHandler(c)

		// Assertions
		assert.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	})
}

func TestWithDB(t *testing.T) {
	t.Run("should call handler with session only", func(t *testing.T) {
		var receivedSession *xorm.Session
		var receivedContext echo.Context

		testHandler := func(s *xorm.Session, c echo.Context) error {
			receivedSession = s
			receivedContext = c
			return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
		}

		// Create Echo instance and request
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Execute the wrapped handler
		wrappedHandler := WithDB(testHandler, false)
		err := wrappedHandler(c)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, receivedSession)
		assert.NotNil(t, receivedContext)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "ok")
	})
}
