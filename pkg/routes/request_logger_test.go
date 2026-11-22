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

package routes

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/mcp"
	"code.vikunja.io/api/pkg/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

// GHSA-ff52-3296-2fvr
func TestHTTPRequestLoggerOmitsSecrets(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	e := echo.New()
	e.Use(httpRequestLogger(logger))
	e.POST("/api/v1/shares/:share/auth", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	t.Run("matched route logs the template", func(t *testing.T) {
		buf.Reset()
		req := httptest.NewRequest(http.MethodPost, "/api/v1/shares/sharehashsecret/auth?token=querysecret", nil)
		e.ServeHTTP(httptest.NewRecorder(), req)

		out := buf.String()
		assert.Contains(t, out, "/api/v1/shares/:share/auth")
		assert.NotContains(t, out, "sharehashsecret")
		assert.NotContains(t, out, "querysecret")
	})

	t.Run("unmatched route logs no path", func(t *testing.T) {
		buf.Reset()
		req := httptest.NewRequest(http.MethodGet, "/share/sharehashsecret/auth?userPasswordReset=querysecret", nil)
		e.ServeHTTP(httptest.NewRecorder(), req)

		out := buf.String()
		assert.Contains(t, out, "status=404")
		assert.NotContains(t, out, "sharehashsecret")
		assert.NotContains(t, out, "querysecret")
	})

	t.Run("frontend asset logs its path", func(t *testing.T) {
		buf.Reset()
		fe := echo.New()
		fe.Use(httpRequestLogger(logger))
		fe.Use(static())
		req := httptest.NewRequest(http.MethodGet, "/index.html?token=querysecret", nil)
		fe.ServeHTTP(httptest.NewRecorder(), req)

		out := buf.String()
		assert.Contains(t, out, "route=/index.html")
		assert.NotContains(t, out, "querysecret")
	})

	t.Run("frontend fallback logs no path", func(t *testing.T) {
		buf.Reset()
		fe := echo.New()
		fe.Use(httpRequestLogger(logger))
		fe.Use(static())
		req := httptest.NewRequest(http.MethodGet, "/share/sharehashsecret/auth", nil)
		fe.ServeHTTP(httptest.NewRecorder(), req)

		out := buf.String()
		assert.Contains(t, out, "status=200")
		assert.NotContains(t, out, "sharehashsecret")
	})
}

func TestHTTPRequestLoggerAuthKind(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	e := echo.New()
	e.Use(httpRequestLogger(logger))
	route := func(path string, set func(c *echo.Context)) {
		e.GET(path, func(c *echo.Context) error {
			set(c)
			return c.NoContent(http.StatusOK)
		})
	}
	route("/session", func(c *echo.Context) {
		c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"type": float64(user.AuthTypeUser)}})
	})
	route("/link-share", func(c *echo.Context) {
		c.Set("user", &jwt.Token{Claims: jwt.MapClaims{"type": float64(2)}})
	})
	route("/token", func(c *echo.Context) { c.Set("api_token", &models.APIToken{}) })
	route("/caldav", func(c *echo.Context) { c.Set("userBasicAuth", &user.User{}) })
	route(mcp.RoutePrefix, func(c *echo.Context) { c.Set("api_token", &models.APIToken{}) })
	route("/none", func(*echo.Context) {})

	for path, want := range map[string]string{
		"/session":      "session",
		"/link-share":   "link_share",
		"/token":        "api_token",
		"/caldav":       "caldav",
		mcp.RoutePrefix: "mcp",
		"/none":         "none",
	} {
		t.Run(path, func(t *testing.T) {
			buf.Reset()
			e.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
			assert.Contains(t, buf.String(), "auth="+want+"\n")
		})
	}
}
