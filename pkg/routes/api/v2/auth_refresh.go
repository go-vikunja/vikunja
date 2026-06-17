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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/auth"
	user2 "code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
)

func init() { AddRouteRegistrar(RegisterRefreshTokenRoutes) }

// RegisterRefreshTokenRoutes wires the refresh-token endpoint. It authenticates
// via the HttpOnly refresh cookie rather than a JWT, so it is a public operation.
func RegisterRefreshTokenRoutes(api huma.API) {
	Register(api, huma.Operation{
		OperationID:   "auth-refresh-token",
		Summary:       "Refresh user token",
		Description:   "Exchanges the refresh-token cookie for a new short-lived JWT. The refresh token is rotated on every call, so the previous one stops working. A new HttpOnly refresh cookie is set on the response.",
		Method:        http.MethodPost,
		Path:          "/user/token/refresh",
		DefaultStatus: http.StatusOK,
		Tags:          []string{"auth"},
		Security:      publicSecurity,
	}, authRefreshToken)
}

func authRefreshToken(ctx context.Context, _ *struct{}) (*authTokenBody, error) {
	ec := echoContextFromCtx(ctx)
	if ec == nil {
		return nil, huma.Error401Unauthorized("No refresh token provided.")
	}

	cookie, err := ec.Cookie(auth.RefreshTokenCookieName)
	if err != nil || cookie.Value == "" {
		return nil, huma.Error401Unauthorized("No refresh token provided.")
	}

	result, err := auth.RefreshSession(cookie.Value)
	if err != nil {
		if user2.IsErrUserStatusError(err) {
			auth.ClearRefreshTokenCookie(ec)
		}
		return nil, translateDomainError(err)
	}

	cookieMaxAge := int(config.ServiceJWTTTL.GetInt64())
	if result.IsLongSession {
		cookieMaxAge = int(config.ServiceJWTTTLLong.GetInt64())
	}
	auth.SetRefreshTokenCookie(ec, result.NewRefreshToken, cookieMaxAge)

	out := &authTokenBody{CacheControl: "no-store"}
	out.Body.Token = result.AccessToken
	return out, nil
}
