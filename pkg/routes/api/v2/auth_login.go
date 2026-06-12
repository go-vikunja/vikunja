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

	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/humaecho5"
	"code.vikunja.io/api/pkg/routes/api/shared"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
)

// authTokenBody wraps the issued user JWT. The token is inlined rather than
// embedding auth.Token because Huma derives schema names from the bare Go type
// name and a top-level auth.Token body would collide with user.Token (the
// caldav-token schema, also named "Token"). The refresh token is delivered out
// of band as an HttpOnly cookie, so it is intentionally absent from the schema.
type authTokenBody struct {
	// Cache-Control: no-store keeps the access token out of any shared cache.
	CacheControl string `header:"Cache-Control"`
	Body         struct {
		Token string `json:"token" readOnly:"true" doc:"The short-lived JWT auth token. Send it as a bearer token on subsequent requests."`
	}
}

// logoutBody confirms a successful logout.
type logoutBody struct {
	Body struct {
		Message string `json:"message" readOnly:"true" doc:"A human-readable confirmation message."`
	}
}

func init() { AddRouteRegistrar(RegisterLoginRoutes) }

// RegisterLoginRoutes wires the local/LDAP login and logout endpoints. Login is
// always registered (LDAP-only deployments still log in here); logout inherits
// the global JWT auth.
func RegisterLoginRoutes(api huma.API) {
	tags := []string{"auth"}

	Register(api, huma.Operation{
		OperationID:   "auth-login",
		Summary:       "Login",
		Description:   "Logs a user in with username and password (and a TOTP passcode when 2FA is enabled), returning a short-lived JWT. A long-lived refresh token is set as an HttpOnly cookie scoped to the refresh endpoint.",
		Method:        http.MethodPost,
		Path:          "/login",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
		Security:      publicSecurity,
	}, authLogin)

	Register(api, huma.Operation{
		OperationID:   "auth-logout",
		Summary:       "Logout",
		Description:   "Destroys the current session server-side and clears the refresh-token cookie. A no-op for API tokens and link shares, which carry no session.",
		Method:        http.MethodPost,
		Path:          "/logout",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, authLogout)
}

func authLogin(ctx context.Context, in *struct{ Body user.Login }) (*authTokenBody, error) {
	u, err := shared.AuthenticateUserCredentials(&in.Body)
	if err != nil {
		return nil, translateDomainError(err)
	}

	deviceInfo, ipAddress := requestClientInfo(ctx)
	token, err := auth.IssueUserToken(u, deviceInfo, ipAddress, in.Body.LongToken)
	if err != nil {
		return nil, translateDomainError(err)
	}

	if ec := echoContextFromCtx(ctx); ec != nil {
		auth.WriteUserAuthCookies(ec, token)
	}

	out := &authTokenBody{CacheControl: "no-store"}
	out.Body.Token = token.AccessToken
	return out, nil
}

func authLogout(ctx context.Context, _ *struct{}) (*logoutBody, error) {
	var sid string
	if ec := echoContextFromCtx(ctx); ec != nil {
		auth.ClearRefreshTokenCookie(ec)
		sid = auth.SessionIDFromContext(ec)
	}

	if err := shared.DeleteSession(sid); err != nil {
		return nil, translateDomainError(err)
	}

	out := &logoutBody{}
	out.Body.Message = "Successfully logged out."
	return out, nil
}

// echoContextFromCtx pulls the underlying *echo.Context off a Huma request
// context so a handler can set cookies and headers the OpenAPI schema does not
// model (the refresh-token cookie). Returns nil when the context carries no echo
// context (it always does under the humaecho5 adapter).
func echoContextFromCtx(ctx context.Context) *echo.Context {
	ec, ok := ctx.Value(humaecho5.EchoContextKey).(*echo.Context)
	if !ok || ec == nil {
		return nil
	}
	return ec
}
