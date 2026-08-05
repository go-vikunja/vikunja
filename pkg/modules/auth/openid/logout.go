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

package openid

import (
	"net/url"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
)

// EndSessionEndpoint returns the provider's RP-Initiated Logout endpoint
// (discovery's end_session_endpoint, cached at init), falling back to the static
// logouturl. Never triggers discovery so logout stays responsive when the OP is
// unreachable.
func (p *Provider) EndSessionEndpoint() string {
	if p.EndSessionURL != "" {
		return p.EndSessionURL
	}
	return p.LogoutURL
}

// discoveredEndSessionEndpoint reads end_session_endpoint from the discovery
// document already cached on the *oidc.Provider, so Claims unmarshals in memory
// without a request.
func (p *Provider) discoveredEndSessionEndpoint() string {
	if p.openIDProvider == nil {
		return ""
	}

	var meta struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}
	if err := p.openIDProvider.Claims(&meta); err != nil {
		log.Debugf("Could not read end_session_endpoint for provider %s: %v", p.Key, err)
		return ""
	}
	return meta.EndSessionEndpoint
}

// BuildEndSessionURL builds an OpenID Connect RP-Initiated Logout 1.0 request URL
// (id_token_hint + post_logout_redirect_uri + client_id; see RP-Initiated Logout
// 1.0 §2). post_logout_redirect_uri defaults to service.publicurl, and the OP
// only honors it when id_token_hint is present. Returns "" when neither an
// end_session_endpoint nor a static logouturl is configured.
func BuildEndSessionURL(providerKey string, oidc *models.SessionOIDCData) (string, error) {
	// GetProvider would trigger OIDC discovery (a live HTTP GET that blocks when
	// the OP is down); the cached static fields are all logout needs.
	provider, err := getCachedProvider(providerKey)
	if err != nil {
		return "", err
	}
	if provider == nil {
		return "", nil
	}

	idToken := ""
	if oidc != nil {
		idToken = oidc.IDToken
	}

	return buildEndSessionURL(
		provider.EndSessionEndpoint(),
		provider.ClientID,
		idToken,
		config.ServicePublicURL.GetString(),
	)
}

// buildEndSessionURL appends the logout query params onto endpoint, omitting
// empty ones, and returns "" for an empty endpoint.
func buildEndSessionURL(endpoint, clientID, idToken, postLogoutRedirectURI string) (string, error) {
	if endpoint == "" {
		return "", nil
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	q := u.Query()
	if clientID != "" {
		q.Set("client_id", clientID)
	}
	if idToken != "" {
		q.Set("id_token_hint", idToken)
	}
	if postLogoutRedirectURI != "" {
		q.Set("post_logout_redirect_uri", postLogoutRedirectURI)
	}
	u.RawQuery = q.Encode()

	return u.String(), nil
}
