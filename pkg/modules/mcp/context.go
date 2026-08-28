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

// The SDK's RequestExtra carries only OAuth TokenInfo and headers, never the
// *http.Request, so the entry handler stashes the user and token on r.Context()
// for tool handlers to read back out.

type userCtxKey struct{}
type tokenCtxKey struct{}

func WithUser(ctx context.Context, u *user.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, u)
}

func WithToken(ctx context.Context, t *models.APIToken) context.Context {
	return context.WithValue(ctx, tokenCtxKey{}, t)
}

func UserFromContext(ctx context.Context) *user.User {
	u, _ := ctx.Value(userCtxKey{}).(*user.User)
	return u
}

func TokenFromContext(ctx context.Context) *models.APIToken {
	t, _ := ctx.Value(tokenCtxKey{}).(*models.APIToken)
	return t
}
