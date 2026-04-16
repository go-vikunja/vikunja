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

package admin

import (
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v5"
)

// User re-exposes fields that the default User JSON view hides.
type User struct {
	*user.User
	IsAdmin bool        `json:"is_admin"`
	Status  user.Status `json:"status"`
	Issuer  string      `json:"issuer"`
	// Subject is the external identifier for federated accounts (OIDC `sub` claim or LDAP DN). Empty for local accounts.
	Subject string `json:"subject,omitempty"`
	// AuthProvider is a display-ready label for the account's auth source. Empty for local accounts (caller is expected to render "Local"), "LDAP" for LDAP, the configured friendly name for OIDC accounts (e.g. "Keycloak"), or the raw issuer URL for OIDC accounts whose issuer no longer matches any configured provider.
	AuthProvider string `json:"auth_provider,omitempty"`
}

// newAdminUser wraps a user.User with the extra admin-only fields, resolving
// the auth provider label when applicable.
func newAdminUser(u *user.User, providers []*openid.Provider) *User {
	return &User{
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

// ListUsers returns paginated users for the admin panel with optional search.
// @Summary List users (admin)
// @Description Paginated list of all users on the instance. Supports search by username/email. Exposes fields hidden from the normal user API (is_admin, status).
// @tags admin
// @Produce json
// @Security JWTKeyAuth
// @Param s query string false "Search string matched against username and email."
// @Param page query int false "Page number, defaults to 1."
// @Param per_page query int false "Items per page, defaults to the service setting."
// @Success 200 {array} admin.User
// @Failure 404 {object} web.HTTPError
// @Router /admin/users [get]
func ListUsers(c *echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = config.ServiceMaxItemsPerPage.GetInt()
	}

	query := c.QueryParam("s")

	var users []*user.User
	sess := s.Limit(perPage, (page-1)*perPage).OrderBy("id ASC")
	if query != "" {
		q := "%" + query + "%"
		sess = sess.Where("username LIKE ? OR email LIKE ?", q, q)
	}
	if err := sess.Find(&users); err != nil {
		return err
	}

	providers, err := openid.GetAllProviders()
	if err != nil {
		return err
	}

	out := make([]*User, 0, len(users))
	for _, u := range users {
		out = append(out, newAdminUser(u, providers))
	}

	return c.JSON(http.StatusOK, out)
}
