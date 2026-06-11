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
	"code.vikunja.io/api/pkg/routes/api/shared"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
)

// publicSecurity is the empty security requirement that opts an operation out of
// the globally-applied JWT/API-token auth. The matching Echo path must also be
// listed in unauthenticatedAPIPaths so the token middleware lets it through.
var publicSecurity = []map[string][]string{}

// registerUserBody is the response wrapper for the registration endpoint.
type registerUserBody struct {
	Body *user.User
}

// messageBody carries a human-readable confirmation for endpoints that report
// success without returning a resource (password reset, email confirm).
type messageBody struct {
	Body struct {
		Message string `json:"message" readOnly:"true" doc:"A human-readable confirmation message."`
	}
}

// linkShareTokenBody wraps the issued link-share auth token and its share.
type linkShareTokenBody struct {
	Body *shared.LinkShareToken
}

func init() { AddRouteRegistrar(RegisterPublicAuthRoutes) }

// RegisterPublicAuthRoutes wires the unauthenticated local-account flows
// (registration, password reset, email confirmation) and the link-share auth
// endpoint. The local-account flows mirror v1 by only registering when local
// auth is enabled; the link-share endpoint follows ServiceEnableLinkSharing.
func RegisterPublicAuthRoutes(api huma.API) {
	if config.AuthLocalEnabled.GetBool() {
		registerLocalAuthRoutes(api)
	}

	if config.ServiceEnableLinkSharing.GetBool() {
		Register(api, huma.Operation{
			OperationID:   "auth-link-share",
			Summary:       "Get an auth token for a link share",
			Description:   "Exchanges a link share's public hash (and password, for password-protected shares) for a JWT auth token scoped to the shared project.",
			Method:        http.MethodPost,
			Path:          "/shares/{share}/auth",
			DefaultStatus: http.StatusOK,
			Tags:          []string{"sharing"},
			Security:      publicSecurity,
		}, authLinkShare)
	}
}

func registerLocalAuthRoutes(api huma.API) {
	authTags := []string{"auth"}

	Register(api, huma.Operation{
		OperationID: "auth-register",
		Summary:     "Register",
		Description: "Creates a new local user account. Returns 404 when registration is disabled on this instance.",
		Method:      http.MethodPost,
		Path:        "/register",
		Tags:        authTags,
		Security:    publicSecurity,
	}, authRegister)

	Register(api, huma.Operation{
		OperationID:   "auth-password-token",
		Summary:       "Request a password reset token",
		Description:   "Requests a token to reset the password for the account with the given email. The token is sent to that email; the response is the same whether or not an account exists.",
		Method:        http.MethodPost,
		Path:          "/user/password/token",
		DefaultStatus: http.StatusOK,
		Tags:          authTags,
		Security:      publicSecurity,
	}, authRequestPasswordToken)

	Register(api, huma.Operation{
		OperationID:   "auth-password-reset",
		Summary:       "Reset a password",
		Description:   "Sets a new password using a previously issued reset token. All of the user's existing sessions are invalidated.",
		Method:        http.MethodPost,
		Path:          "/user/password/reset",
		DefaultStatus: http.StatusOK,
		Tags:          authTags,
		Security:      publicSecurity,
	}, authResetPassword)

	Register(api, huma.Operation{
		OperationID:   "auth-confirm-email",
		Summary:       "Confirm an email address",
		Description:   "Confirms the email address of a newly registered user using the token sent to that email.",
		Method:        http.MethodPost,
		Path:          "/user/confirm",
		DefaultStatus: http.StatusOK,
		Tags:          authTags,
		Security:      publicSecurity,
	}, authConfirmEmail)
}

func authRegister(_ context.Context, in *struct{ Body shared.UserRegister }) (*registerUserBody, error) {
	if !config.ServiceEnableRegistration.GetBool() {
		return nil, huma.Error404NotFound("registration is disabled")
	}

	newUser, err := shared.RegisterUser(&in.Body)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &registerUserBody{Body: newUser}, nil
}

func authRequestPasswordToken(_ context.Context, in *struct{ Body user.PasswordTokenRequest }) (*messageBody, error) {
	if err := shared.RequestPasswordResetToken(&in.Body); err != nil {
		return nil, translateDomainError(err)
	}
	out := &messageBody{}
	out.Body.Message = "Token was sent."
	return out, nil
}

func authResetPassword(_ context.Context, in *struct{ Body user.PasswordReset }) (*messageBody, error) {
	if err := shared.ResetPassword(&in.Body); err != nil {
		return nil, translateDomainError(err)
	}
	out := &messageBody{}
	out.Body.Message = "The password was updated successfully."
	return out, nil
}

func authConfirmEmail(_ context.Context, in *struct{ Body user.EmailConfirm }) (*messageBody, error) {
	if err := shared.ConfirmEmail(&in.Body); err != nil {
		return nil, translateDomainError(err)
	}
	out := &messageBody{}
	out.Body.Message = "The email was confirmed successfully."
	return out, nil
}

func authLinkShare(_ context.Context, in *struct {
	Share string `path:"share" doc:"The public hash of the link share."`
	// Pointer so the body is optional: shares without a password are
	// authenticated with no body at all.
	Body *struct {
		Password string `json:"password" doc:"The password for password-protected link shares. Ignored for shares without a password."`
	}
}) (*linkShareTokenBody, error) {
	var password string
	if in.Body != nil {
		password = in.Body.Password
	}

	token, err := shared.AuthenticateLinkShare(in.Share, password)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &linkShareTokenBody{Body: token}, nil
}
