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
	"testing"

	"code.vikunja.io/api/pkg/routes/middleware"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeArrayParams(t *testing.T) {
	e := echo.New()
	e.Use(middleware.NormalizeArrayParams())
	e.GET("/", func(c *echo.Context) error {
		got := (*c).Request().URL.RawQuery
		return (*c).String(200, got)
	})

	req := httptest.NewRequest("GET", "/?foo[]=a&foo[]=b&bar=x", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
	// foo[] should have been rewritten to foo; relative order preserved
	assert.Contains(t, rec.Body.String(), "foo=a")
	assert.Contains(t, rec.Body.String(), "foo=b")
	assert.Contains(t, rec.Body.String(), "bar=x")
	assert.NotContains(t, rec.Body.String(), "foo[]")
}
