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

package oauth2server

import (
	"net/http"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"github.com/labstack/echo/v5"
)

type OIDCWellKnownResponse struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	EndSessionEndpoint                string   `json:"end_session_endpoint,omitempty"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	RegistrationEndpoint              string   `json:"registration_endpoint"`
}

func OidHandler(c *echo.Context) error {
	publicURL := strings.TrimSuffix(config.ServicePublicURL.GetString(), "/")

	response := OIDCWellKnownResponse{
		Issuer:                            publicURL,
		AuthorizationEndpoint:             publicURL + "/oauth/authorize",
		TokenEndpoint:                     publicURL + "/api/v1/oauth/token",
		UserInfoEndpoint:                  publicURL + "/api/v1/user",
		JwksURI:                           publicURL + "/api/v1/.well-known/jwks.json",
		ResponseTypesSupported:            []string{"code"},
		SubjectTypesSupported:             []string{"public"},
		IDTokenSigningAlgValuesSupported:  []string{"RS256"},
		ScopesSupported:                   []string{"openid", "profile", "email"},
		TokenEndpointAuthMethodsSupported: []string{"client_secret_post", "client_secret_basic"},
		ClaimsSupported: []string{
			"sub", "name", "email", "email_verified",
		},
		CodeChallengeMethodsSupported: []string{"S256"},
		GrantTypesSupported:           []string{"authorization_code", "refresh_token"},
		EndSessionEndpoint:            publicURL + "/api/v1/auth/openid/logout",
		RegistrationEndpoint:          publicURL + "/api/v1/auth/openid/register",
	}

	return c.JSON(http.StatusOK, response)
}
