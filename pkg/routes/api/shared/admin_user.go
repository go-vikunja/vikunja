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

// Package shared holds route helpers used by both /api/v1 and /api/v2 so the two
// versions render identical responses without one importing the other.
package shared

import (
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/user"
)

// AdminUser re-exposes fields hidden by the default user.User JSON view.
type AdminUser struct {
	*user.User
	IsAdmin      bool        `json:"is_admin" readOnly:"true" doc:"Whether the user is an instance admin."`
	Status       user.Status `json:"status" readOnly:"true" doc:"Account status (0=active, 1=email-confirmation required, 2=disabled, 3=locked)."`
	Issuer       string      `json:"issuer" readOnly:"true" doc:"Authentication issuer; empty or 'local' for local accounts."`
	Subject      string      `json:"subject,omitempty" readOnly:"true" doc:"External subject identifier, for non-local accounts."`
	AuthProvider string      `json:"auth_provider,omitempty" readOnly:"true" doc:"Resolved auth provider name (e.g. 'LDAP' or an OIDC provider), empty for local accounts."`
}

// NewAdminUser builds the admin-facing user view, resolving the auth-provider
// display name from the configured OIDC providers.
func NewAdminUser(u *user.User, providers []*openid.Provider) *AdminUser {
	return &AdminUser{
		User:         u,
		IsAdmin:      u.IsAdmin,
		Status:       u.Status,
		Issuer:       u.Issuer,
		Subject:      u.Subject,
		AuthProvider: resolveAuthProvider(u, providers),
	}
}

func resolveAuthProvider(u *user.User, providers []*openid.Provider) string {
	switch u.Issuer {
	case "", user.IssuerLocal:
		return ""
	case user.IssuerLDAP:
		return "LDAP"
	}
	for _, provider := range providers {
		issuerURL, err := provider.Issuer()
		if err != nil {
			continue
		}
		if issuerURL == u.Issuer {
			return provider.Name
		}
	}
	return u.Issuer
}
