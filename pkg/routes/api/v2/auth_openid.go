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
	"errors"
	"net/http"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/auth/openid"

	"github.com/danielgtaylor/huma/v2"
)

func init() { AddRouteRegistrar(RegisterOpenIDRoutes) }

// RegisterOpenIDRoutes wires the OpenID Connect callback endpoint. It is only
// registered when OpenID is enabled; individual providers are still resolved per
// request, so an unknown provider key 404s even when others are configured.
func RegisterOpenIDRoutes(api huma.API) {
	if !config.AuthOpenIDEnabled.GetBool() {
		return
	}

	Register(api, huma.Operation{
		OperationID:   "auth-openid-callback",
		Summary:       "Authenticate with OpenID Connect",
		Description:   "Exchanges the authorization code returned by an OpenID Connect provider for a Vikunja JWT, creating or updating the matching user. A long-lived refresh token is set as an HttpOnly cookie. When the resolved user has 2FA enabled, the call returns 412 and must be retried with totp_passcode set.",
		Method:        http.MethodPost,
		Path:          "/auth/openid/{provider}/callback",
		DefaultStatus: http.StatusOK,
		Tags:          []string{"auth"},
		Security:      publicSecurity,
	}, authOpenIDCallback)
}

func authOpenIDCallback(ctx context.Context, in *struct {
	Provider string          `path:"provider" doc:"The OpenID Connect provider key as returned by the /info endpoint."`
	Body     openid.Callback `doc:"The provider callback, carrying the authorization code."`
}) (*authTokenBody, error) {
	u, err := openid.AuthenticateCallback(&in.Body, in.Provider) //nolint:contextcheck // resolves providers from a cached, context-less map and runs OIDC discovery on its own background context, like the v1 callback.
	if err != nil {
		return nil, translateOpenIDError(err)
	}

	deviceInfo, ipAddress := requestClientInfo(ctx)
	// OIDC logins are not "remember me" sessions; v1 always issues a short one.
	token, err := auth.IssueUserToken(u, deviceInfo, ipAddress, false)
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

// translateOpenIDError maps OIDC callback errors to RFC 9457 responses.
// ErrOpenIDBadRequestWithDetails carries no HTTP semantics of its own (v1 renders
// it with a bespoke {message, details} body), so v2 maps it to a 400 with the
// provider detail attached as a structured error detail rather than porting the
// bespoke shape. Everything else flows through translateDomainError.
func translateOpenIDError(err error) error {
	var detailedErr *models.ErrOpenIDBadRequestWithDetails
	if errors.As(err, &detailedErr) {
		return huma.Error400BadRequest(detailedErr.Message, &huma.ErrorDetail{
			Message: "The identity provider rejected the request.",
			Value:   detailedErr.Details,
		})
	}
	return translateDomainError(err)
}
