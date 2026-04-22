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
//
// The rewrite preserves the original left-to-right appearance order of
// parameters, which matters for order-sensitive multi-value fields like
// sort_by and order_by when a client sends a mix of `foo` and `foo[]`.
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
			req.URL.RawQuery = stripBracketSuffix(rq)
			return next(c)
		}
	}
}

// stripBracketSuffix walks a raw query string pair by pair, trimming the
// `[]` suffix from any keys that have one (whether literal or percent-
// encoded). Values are left untouched — they are already URL-encoded and
// don't need re-escaping. Walking the raw query keeps the original
// parameter order intact, unlike url.ParseQuery which returns a map.
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
