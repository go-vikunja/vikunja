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

package apiv2

import (
	"context"
	"net/http"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/danielgtaylor/huma/v2"
)

// tokenTestBody is the response for the token-check endpoints.
type tokenTestBody struct {
	Body struct {
		Message string `json:"message" readOnly:"true" doc:"A static confirmation message."`
	}
}

// apiRoutesBody is the response for the token-routes endpoint: the available
// API routes grouped by permission, for building API-token scopes.
type apiRoutesBody struct {
	Body map[string]models.APITokenRoute
}

// renewTokenBody wraps a freshly issued link-share JWT. The token field is
// inlined rather than embedding auth.Token because Huma derives schema names
// from the bare Go type name, and a top-level auth.Token body would collide with
// user.Token (the caldav-token schema, also named "Token").
type renewTokenBody struct {
	Body struct {
		Token string `json:"token" readOnly:"true" doc:"The renewed JWT auth token."`
	}
}

func init() { AddRouteRegistrar(RegisterTokenMetaRoutes) }

// RegisterTokenMetaRoutes wires the token introspection helpers and the
// link-share token renewal endpoint.
func RegisterTokenMetaRoutes(api huma.API) {
	tags := []string{"auth"}

	// v1 served GET as a 200 "ok" and POST as a 418 teapot easter egg; v2 makes
	// both a plain 200 so a token check is an ordinary success.
	Register(api, huma.Operation{
		OperationID:   "token-test",
		Summary:       "Test a token",
		Description:   "Returns 200 if the bearer token (JWT or API token) is valid. Used to check authentication.",
		Method:        http.MethodGet,
		Path:          "/token/test",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, tokenTest)

	Register(api, huma.Operation{
		OperationID:   "token-check",
		Summary:       "Check a token",
		Description:   "Returns 200 if the bearer token (JWT or API token) is valid. Used to check authentication.",
		Method:        http.MethodPost,
		Path:          "/token/test",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, tokenCheck)

	Register(api, huma.Operation{
		OperationID: "token-routes",
		Summary:     "List API token routes",
		Description: "Returns every API route available to scope an API token against, grouped by resource and permission. Covers both /api/v1 and /api/v2 routes.",
		Method:      http.MethodGet,
		Path:        "/routes",
		Tags:        []string{"api"},
	}, tokenRoutes)

	Register(api, huma.Operation{
		OperationID:   "token-renew",
		Summary:       "Renew a link-share token",
		Description:   "Issues a fresh JWT for the current link share. Only link-share tokens can be renewed here; user sessions must use the refresh-token flow.",
		Method:        http.MethodPost,
		Path:          "/user/token",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, tokenRenew)
}

func tokenTest(_ context.Context, _ *struct{}) (*tokenTestBody, error) {
	out := &tokenTestBody{}
	out.Body.Message = "ok"
	return out, nil
}

func tokenCheck(_ context.Context, _ *struct{}) (*tokenTestBody, error) {
	out := &tokenTestBody{}
	out.Body.Message = "ok"
	return out, nil
}

func tokenRoutes(_ context.Context, _ *struct{}) (*apiRoutesBody, error) {
	return &apiRoutesBody{Body: models.GetAPITokenRoutes()}, nil
}

func tokenRenew(ctx context.Context, _ *struct{}) (*renewTokenBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	// Only link-share tokens are renewable here; a user JWT lands as *user.User
	// and must use the refresh-token flow instead.
	share, ok := a.(*models.LinkSharing)
	if !ok {
		return nil, huma.Error400BadRequest("User tokens cannot be renewed via this endpoint. Use the refresh-token flow instead.")
	}

	t, err := auth.NewLinkShareJWTAuthtoken(share)
	if err != nil {
		return nil, translateDomainError(err)
	}

	out := &renewTokenBody{}
	out.Body.Token = t
	return out, nil
}
