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

import "net/url"

// MetadataPath is where RFC 8414 clients look for the metadata of an issuer without a path.
const MetadataPath = "/.well-known/oauth-authorization-server"

// IssuerPath is the path of the issuer, empty unless Vikunja is served from a subpath.
func IssuerPath() string {
	issuer, ok := Issuer()
	if !ok {
		return ""
	}
	u, err := url.Parse(issuer)
	if err != nil {
		return ""
	}
	return u.Path
}

// AuthorizationServerMetadata is the RFC 8414 metadata document.
type AuthorizationServerMetadata struct {
	Issuer                                     string   `json:"issuer"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	GrantTypesSupported                        []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported              []string `json:"code_challenge_methods_supported"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	ScopesSupported                            []Scope  `json:"scopes_supported"`
	ClientIDMetadataDocumentSupported          bool     `json:"client_id_metadata_document_supported"`
	AuthorizationResponseIssParameterSupported bool     `json:"authorization_response_iss_parameter_supported"`
}

// Metadata describes this authorization server; there is none without an issuer.
func Metadata() (*AuthorizationServerMetadata, bool) {
	issuer, ok := Issuer()
	if !ok {
		return nil, false
	}
	return &AuthorizationServerMetadata{
		Issuer: issuer,
		// The frontend route: it asks the user for consent and then calls the authorize API.
		AuthorizationEndpoint:  issuer + "/oauth/authorize",
		TokenEndpoint:          issuer + "/api/v2/oauth/token",
		ResponseTypesSupported: []string{"code"},
		// refresh_token only renews first-party sessions; scoped grants are never refreshable.
		GrantTypesSupported:                        []string{"authorization_code"},
		CodeChallengeMethodsSupported:              []string{"S256"},
		TokenEndpointAuthMethodsSupported:          []string{"none"},
		ScopesSupported:                            SupportedScopes(),
		ClientIDMetadataDocumentSupported:          true,
		AuthorizationResponseIssParameterSupported: true,
	}, true
}
