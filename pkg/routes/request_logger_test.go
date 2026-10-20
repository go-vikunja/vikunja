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
}
