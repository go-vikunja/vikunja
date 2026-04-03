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
	"net/url"
	"strings"

	"code.vikunja.io/api/pkg/models"
)

// ValidateRedirectURI checks that the redirect_uri uses a scheme starting with
// "vikunja-". This allowlists only Vikunja native app schemes (e.g.
// vikunja-flutter://callback) and rejects dangerous schemes like javascript:,
// data:, http:, https:, etc. The stored redirect URIs are URL-encoded to prevent
// injection attacks, so they are decoded before comparison.
func ValidateRedirectURI(req authorizeRequest, client *models.OAuthClient) bool {

	decodedStored := urlDecodeRedirectURIs(client.RedirectURIs)
	if contains(decodedStored, req.RedirectURI) {
		return true
	}

	u, err := url.Parse(req.RedirectURI)
	if err != nil || u.Scheme == "" {
		return false
	}

	return strings.HasPrefix(u.Scheme, "vikunja-")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func urlDecodeRedirectURIs(encoded string) []string {
	if encoded == "" {
		return nil
	}
	parts := strings.Split(encoded, ",")
	decoded := make([]string, len(parts))
	for i, part := range parts {
		decoded[i], _ = url.QueryUnescape(part)
	}
	return decoded
}
