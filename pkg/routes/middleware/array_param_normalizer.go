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

package middleware

import (
	"net/url"
	"strings"

	"github.com/labstack/echo/v5"
)

// NormalizeArrayParams rewrites `foo[]=...` query parameters to `foo=...`
// before the router sees them. The frontend's URLSearchParams emits the
// PHP-style `[]` suffix for arrays; echo does not unify those with plain
// repeated fields. Normalising here lets handlers declare a single
// `query:"foo"` tag and receive both shapes.
func NormalizeArrayParams() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := (*c).Request()
			rq := req.URL.RawQuery
			if rq == "" {
				return next(c)
			}
			// Fast path: skip parsing when the query carries no `[]` suffix in
			// any of its possible encodings (raw or percent-encoded by browsers
			// via URLSearchParams).
			if !strings.Contains(rq, "[]=") &&
				!strings.Contains(rq, "%5B%5D=") &&
				!strings.Contains(rq, "%5b%5d=") &&
				!strings.HasSuffix(rq, "[]") &&
				!strings.HasSuffix(rq, "%5B%5D") &&
				!strings.HasSuffix(rq, "%5b%5d") {
				return next(c)
			}
			values, err := url.ParseQuery(rq)
			if err != nil {
				return next(c) // malformed; let downstream handle it
			}
			rewritten := url.Values{}
			for k, vs := range values {
				key := strings.TrimSuffix(k, "[]")
				rewritten[key] = append(rewritten[key], vs...)
			}
			req.URL.RawQuery = rewritten.Encode()
			return next(c)
		}
	}
}
