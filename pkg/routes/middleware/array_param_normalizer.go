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

// NormalizeArrayParams rewrites `foo[]=...` to `foo=...` before routing,
// so handlers use a single `query:"foo"` tag for both shapes. URLSearchParams
// emits the `[]` suffix; echo doesn't unify it with repeated fields. Order
// is preserved for order-sensitive params (sort_by, order_by) when clients
// mix `foo` and `foo[]`.
func NormalizeArrayParams() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := (*c).Request()
			rq := req.URL.RawQuery
			if rq == "" {
				return next(c)
			}
			// Fast path: skip when no `[]` suffix in any encoding (raw or browser-percent-encoded).
			if !strings.Contains(rq, "[]=") &&
				!strings.Contains(rq, "%5B%5D=") &&
				!strings.Contains(rq, "%5b%5d=") &&
				!strings.HasSuffix(rq, "[]") &&
				!strings.HasSuffix(rq, "%5B%5D") &&
				!strings.HasSuffix(rq, "%5b%5d") {
				return next(c)
			}
			req.URL.RawQuery = stripBracketSuffix(rq)
			return next(c)
		}
	}
}

// stripBracketSuffix walks the raw query pair-by-pair (rather than via
// url.ParseQuery's map) to preserve parameter order, trimming `[]` from
// keys. Values are left as-is — already URL-encoded.
func stripBracketSuffix(rq string) string {
	var out strings.Builder
	out.Grow(len(rq))
	first := true
	for _, pair := range strings.Split(rq, "&") {
		if pair == "" {
			continue
		}
		key, val, hasEq := strings.Cut(pair, "=")
		decoded, err := url.QueryUnescape(key)
		if err != nil {
			decoded = key
		}
		decoded = strings.TrimSuffix(decoded, "[]")
		if !first {
			out.WriteByte('&')
		}
		first = false
		out.WriteString(url.QueryEscape(decoded))
		if hasEq {
			out.WriteByte('=')
			out.WriteString(val)
		}
	}
	return out.String()
}
