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

// Package shared holds helpers used by both the v1 and v2 route packages. It
// sits above the auth/user modules in the import graph, so it can combine them
// without creating a cycle.
package shared

import (
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/user"
)

// GetAuthProviderName resolves the human-readable name of the source a user
// authenticated with: "local"/"ldap" for those issuers, otherwise the
// configured OpenID provider whose issuer URL matches the user's. Returns ""
// when no provider matches.
func GetAuthProviderName(u *user.User) (string, error) {
	switch u.Issuer {
	case user.IssuerLocal:
		return "local", nil
	case user.IssuerLDAP:
		return "ldap", nil
	}

	providers, err := openid.GetAllProviders()
	if err != nil {
		return "", err
	}
	for _, provider := range providers {
		issuerURL, err := provider.Issuer()
		if err != nil {
			return "", err
		}
		if issuerURL == u.Issuer {
			return provider.Name, nil
		}
	}

	return "", nil
}
