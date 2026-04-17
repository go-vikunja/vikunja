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
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// User re-exposes fields hidden by the default User JSON view.
type User struct {
	*user.User
	IsAdmin      bool        `json:"is_admin"`
	Status       user.Status `json:"status"`
	Issuer       string      `json:"issuer"`
	Subject      string      `json:"subject,omitempty"`
	AuthProvider string      `json:"auth_provider,omitempty"`
}

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

// UserList is the CRUDable wrapper backing the admin list-users route via
// handler.ReadAllWeb. Only ReadAll is used; everything else is gated by
// the RequireInstanceAdmin middleware.
type UserList struct{}

// ReadAll returns paginated users, optionally filtered by username/email.
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
func (*UserList) ReadAll(s *xorm.Session, _ web.Auth, search string, page, perPage int) (interface{}, int, int64, error) {
	finder := s.Limit(perPage, (page-1)*perPage).OrderBy("id ASC")
	counter := s
	if search != "" {
		q := "%" + search + "%"
		finder = finder.Where("username LIKE ? OR email LIKE ?", q, q)
		counter = s.Where("username LIKE ? OR email LIKE ?", q, q)
	}

	var users []*user.User
	if err := finder.Find(&users); err != nil {
		return nil, 0, 0, err
	}

	totalCount, err := counter.Count(&user.User{})
	if err != nil {
		return nil, 0, 0, err
	}

	providers, err := openid.GetAllProviders()
	if err != nil {
		return nil, 0, 0, err
	}

	out := make([]*User, 0, len(users))
	for _, u := range users {
		out = append(out, newAdminUser(u, providers))
	}
	return out, len(out), totalCount, nil
}

func (*UserList) ReadOne(*xorm.Session, web.Auth) error              { return nil }
func (*UserList) Create(*xorm.Session, web.Auth) error               { return nil }
func (*UserList) Update(*xorm.Session, web.Auth) error               { return nil }
func (*UserList) Delete(*xorm.Session, web.Auth) error               { return nil }
func (*UserList) CanCreate(*xorm.Session, web.Auth) (bool, error)    { return false, nil }
func (*UserList) CanDelete(*xorm.Session, web.Auth) (bool, error)    { return false, nil }
func (*UserList) CanUpdate(*xorm.Session, web.Auth) (bool, error)    { return false, nil }
func (*UserList) CanRead(*xorm.Session, web.Auth) (bool, int, error) { return true, 0, nil }
