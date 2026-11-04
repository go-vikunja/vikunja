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

package middleware_test

import (
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/routes/middleware"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureUserAgent(t *testing.T, userAgent string) string {
	t.Helper()
	e := echo.New()
	e.Use(middleware.RequestMeta())
	var got string
	e.GET("/", func(c *echo.Context) error {
		meta := events.RequestMetaFromContext((*c).Request().Context())
		require.NotNil(t, meta)
		got = meta.UserAgent
		return (*c).NoContent(200)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("User-Agent", userAgent)
	e.ServeHTTP(httptest.NewRecorder(), req)
	return got
}

func TestRequestMetaUserAgent(t *testing.T) {
	t.Run("short user agent is kept", func(t *testing.T) {
		assert.Equal(t, "Mozilla/5.0", captureUserAgent(t, "Mozilla/5.0"))
	})
	t.Run("long user agent is truncated", func(t *testing.T) {
		assert.Equal(t, strings.Repeat("a", 512), captureUserAgent(t, strings.Repeat("a", 2000)))
	})
	t.Run("truncation does not split a multi-byte character", func(t *testing.T) {
		got := captureUserAgent(t, strings.Repeat("a", 511)+strings.Repeat("ü", 10))
		assert.True(t, utf8.ValidString(got))
		assert.Equal(t, strings.Repeat("a", 511), got)
	})
}
