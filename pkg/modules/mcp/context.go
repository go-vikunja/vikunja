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
	"net/http"

	"code.vikunja.io/api/pkg/models"
)

type tokenCtxKey struct{}
type callerCtxKey struct{}

func WithToken(ctx context.Context, t *models.APIToken) context.Context {
	return context.WithValue(ctx, tokenCtxKey{}, t)
}
func TokenFromContext(ctx context.Context) *models.APIToken {
	t, _ := ctx.Value(tokenCtxKey{}).(*models.APIToken)
	return t
}

// The SDK does not pass the inbound request to tool handlers.
func WithCaller(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, callerCtxKey{}, r)
}
func CallerFromContext(ctx context.Context) *http.Request {
	r, _ := ctx.Value(callerCtxKey{}).(*http.Request)
	return r
}
