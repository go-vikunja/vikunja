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

// Package humabridge mounts Huma's echo/v5 adapter (upstream humaecho) onto
// an Echo group and adds the Vikunja-specific glue upstream doesn't provide:
//
//   - every request through the group stashes its *echo.Context on the
//     request context under EchoContextKey, so handlers can reach the echo
//     context via auth.GetAuthFromContext without per-handler wiring, and
//   - internal Huma dispatches (autopatch's GET+PUT round trip) resolve
//     paths relative to the group, so the returned API's Adapter() rewrites
//     them onto absolute URLs before re-entering the root router.
package humabridge

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v5"
)

type echoContextKey struct{}

// EchoContextKey retrieves the underlying *echo.Context from a Huma
// handler's context.Context.
var EchoContextKey = echoContextKey{}

func stashEchoContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		req := c.Request()
		c.SetRequest(req.WithContext(context.WithValue(req.Context(), EchoContextKey, c)))
		return next(c)
	}
}

// groupPrefixAdapter prepends groupPrefix to internal Huma dispatches whose
// path doesn't already start with it. External requests never pass through
// here — they hit the echo router directly.
type groupPrefixAdapter struct {
	huma.Adapter
	groupPrefix string
}

func (a *groupPrefixAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a.groupPrefix != "" && !strings.HasPrefix(r.URL.Path, a.groupPrefix) {
		r = r.Clone(r.Context())
		r.URL.Path = a.groupPrefix + r.URL.Path
		if r.URL.RawPath != "" {
			r.URL.RawPath = a.groupPrefix + r.URL.RawPath
		}
	}
	a.Adapter.ServeHTTP(w, r)
}

type groupPrefixAPI struct {
	huma.API
	adapter huma.Adapter
}

// Adapter overrides the embedded API's adapter so everything huma resolves
// through api.Adapter() — operation registration and autopatch's internal
// ServeHTTP round trips — goes through the prefix rewrite.
func (a *groupPrefixAPI) Adapter() huma.Adapter { return a.adapter }

// NewWithGroup mounts a Huma API on a group so handlers inherit its
// middleware. groupPrefix must equal the prefix g was constructed with.
func NewWithGroup(e *echo.Echo, g *echo.Group, groupPrefix string, config huma.Config) huma.API {
	// Must run before humaecho registers any route: echo snapshots a group's
	// middleware chain at route-registration time.
	g.Use(stashEchoContext)
	api := humaecho.NewWithGroup(e, g, config)
	return &groupPrefixAPI{
		API:     api,
		adapter: &groupPrefixAdapter{Adapter: api.Adapter(), groupPrefix: groupPrefix},
	}
}
