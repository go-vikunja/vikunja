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
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/log"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func newStaticTestEcho() *echo.Echo {
	log.InitLogger()
	e := echo.New()
	e.Use(static())
	e.GET("/api/v1/ping", func(c *echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})
	return e
}

// Paths still looking url-encoded after the first decode used to be decoded twice
// and 500 with url.EscapeError.
// See https://github.com/go-vikunja/vikunja/issues/3434
func TestStaticEncodedPath(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "decodes to an invalid escape sequence",
			url:  "/%25u002f.env", // decodes to /%u002f.env
		},
		{
			name: "decodes to an encoded slash",
			url:  "/%252e%252e%252fetc/passwd", // decodes to /%2e%2e%2fetc/passwd
		},
		{
			name: "decodes to a trailing percent sign",
			url:  "/foo%25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newStaticTestEcho()
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.url, nil))

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

func TestStaticServesExistingFile(t *testing.T) {
	e := newStaticTestEcho()
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/index.html", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestStaticDoesNotHandleAPIRoutes(t *testing.T) {
	e := newStaticTestEcho()
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "pong", rec.Body.String())
}
