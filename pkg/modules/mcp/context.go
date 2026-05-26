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

package mcp

import (
	"context"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
)

// Context propagation between the Echo entry handler and downstream tool
// handlers. The SDK's RequestExtra only carries OAuth TokenInfo + headers —
// it does not expose *http.Request — so we attach the authenticated user
// and the API token to r.Context() at the entry boundary and pull them out
// inside tool handlers via the accessors below.
//
// Typed keys (unexported empty structs) avoid collisions with any other
// package that might write to the same context.

type userCtxKey struct{}
type tokenCtxKey struct{}

// WithUser returns a new context that carries the authenticated user.
func WithUser(ctx context.Context, u *user.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, u)
}

// WithToken returns a new context that carries the API token used for the
// current MCP request.
func WithToken(ctx context.Context, t *models.APIToken) context.Context {
	return context.WithValue(ctx, tokenCtxKey{}, t)
}

// UserFromContext returns the authenticated user attached by the MCP entry
// handler, or nil if no user is present.
func UserFromContext(ctx context.Context) *user.User {
	u, _ := ctx.Value(userCtxKey{}).(*user.User)
	return u
}

// TokenFromContext returns the API token attached by the MCP entry handler,
// or nil if no token is present.
func TokenFromContext(ctx context.Context) *models.APIToken {
	t, _ := ctx.Value(tokenCtxKey{}).(*models.APIToken)
	return t
}
