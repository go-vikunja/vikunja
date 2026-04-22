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

	"code.vikunja.io/api/pkg/routes/middleware"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeArrayParams(t *testing.T) {
	cases := []struct {
		name     string
		rawQuery string
	}{
		{"literal brackets", "foo[]=a&foo[]=b&bar=x"},
		{"percent-encoded uppercase", "foo%5B%5D=a&foo%5B%5D=b&bar=x"},
		{"percent-encoded lowercase", "foo%5b%5d=a&foo%5b%5d=b&bar=x"},
		{"mixed plain and bracketed", "foo=a&foo%5B%5D=b&bar=x"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			e.Use(middleware.NormalizeArrayParams())
			e.GET("/", func(c *echo.Context) error {
				got := (*c).Request().URL.RawQuery
				return (*c).String(200, got)
			})

			req := httptest.NewRequest("GET", "/?"+tc.rawQuery, nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, 200, rec.Code)
			body := rec.Body.String()
			assert.Contains(t, body, "foo=a")
			assert.Contains(t, body, "foo=b")
			assert.Contains(t, body, "bar=x")
			assert.NotContains(t, body, "foo[]")
			assert.NotContains(t, body, "%5B%5D")
			assert.NotContains(t, body, "%5b%5d")
		})
	}
}

// TestNormalizeArrayParamsPreservesOrder guards against map-iteration
// reordering when mixed bracketed and non-bracketed forms of the same
// key are present. Order matters for sort_by / order_by.
func TestNormalizeArrayParamsPreservesOrder(t *testing.T) {
	e := echo.New()
	e.Use(middleware.NormalizeArrayParams())
	e.GET("/", func(c *echo.Context) error {
		return (*c).String(200, (*c).Request().URL.RawQuery)
	})

	// Run many times — a map-iteration bug would surface probabilistically.
	for range 50 {
		req := httptest.NewRequest("GET", "/?sort_by=a&sort_by%5B%5D=b&sort_by=c&sort_by%5B%5D=d", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, 200, rec.Code)
		body := rec.Body.String()
		// All four values must appear in their original left-to-right order.
		idxA := strings.Index(body, "sort_by=a")
		idxB := strings.Index(body, "sort_by=b")
		idxC := strings.Index(body, "sort_by=c")
		idxD := strings.Index(body, "sort_by=d")
		assert.True(t, idxA >= 0 && idxB > idxA && idxC > idxB && idxD > idxC,
			"expected a<b<c<d, got a=%d b=%d c=%d d=%d (body=%s)", idxA, idxB, idxC, idxD, body)
	}
}
