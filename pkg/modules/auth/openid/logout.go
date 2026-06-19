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

// EndSessionEndpoint returns the provider's RP-Initiated Logout endpoint as
// published in its OpenID Connect discovery document (the REQUIRED
// `end_session_endpoint` metadata, RP-Initiated Logout 1.0 §2.1). When the
// provider does not publish one, it falls back to the statically configured
// `logouturl` so existing setups keep working.
func (p *Provider) EndSessionEndpoint() string {
	if p.openIDProvider == nil {
		if err := p.setOicdProvider(); err != nil {
			return p.LogoutURL
		}
	}

	var meta struct {
		EndSessionEndpoint string `json:"end_session_endpoint"`
	}
	if err := p.openIDProvider.Claims(&meta); err != nil {
		log.Debugf("Could not read end_session_endpoint for provider %s: %v", p.Key, err)
		return p.LogoutURL
	}
	if meta.EndSessionEndpoint == "" {
		return p.LogoutURL
	}
	return meta.EndSessionEndpoint
}

// BuildEndSessionURL constructs an OpenID Connect RP-Initiated Logout 1.0 request
// URL for the given provider key and stored session OIDC data.
//
// Per RP-Initiated Logout 1.0 §2 it appends:
//   - id_token_hint: the ID token previously issued to this session. RECOMMENDED;
//     it lets the OP skip the logout-confirmation prompt and is what makes the OP
//     honor post_logout_redirect_uri (the OP MAY require it, §3).
//   - post_logout_redirect_uri: where the OP redirects the user agent after
//     logout. MUST be pre-registered with the OP. Defaults to service.publicurl
//     (the Vikunja frontend) so the user lands back on Vikunja's login page.
//   - client_id: the RP's client identifier (§2). Always sent; the OP verifies it
//     matches the one in id_token_hint.
//
// It returns "" (and the caller skips the redirect) when neither an
// end_session_endpoint nor a static logouturl is configured.
func BuildEndSessionURL(providerKey string, oidc *models.SessionOIDCData) (string, error) {
	provider, err := GetProvider(providerKey)
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

// buildEndSessionURL assembles the RP-Initiated Logout query string onto the
// given end-session endpoint. Empty optional params are omitted. Returns "" when
// no endpoint is configured.
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
