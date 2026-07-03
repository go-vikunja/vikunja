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
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/routes/api/shared"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// UserList backs the admin list-users route via handler.ReadAllWeb; only ReadAll is used.
type UserList struct {
	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// ReadAll returns paginated users, optionally filtered by username/email.
// @Summary List users (admin)
// @Description Paginated list of all users on the instance. Supports search by username/email. Exposes fields hidden from the normal user API (is_admin, status).
// @tags admin
// @Produce json
// @Security JWTKeyAuth
// @Param s query string false "Search string matched against username and email."
// @Param page query int false "Page number, defaults to 1."
// @Param per_page query int false "Items per page, defaults to the service setting."
// @Success 200 {array} shared.AdminUser
// @Failure 404 {object} web.HTTPError
// @Router /admin/users [get]
func (*UserList) ReadAll(s *xorm.Session, a web.Auth, search string, page, perPage int) (interface{}, int, int64, error) {
	// The response exposes every user's email address; compliance regimes want
	// admin PII reads logged. Queued here, dispatched by DoReadAll's
	// DispatchPending with the request context.
	if doer, err := user.GetFromAuth(a); err == nil {
		events.DispatchOnCommit(s, &models.AdminUsersListedEvent{Doer: doer})
	}

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

	out := make([]*shared.AdminUser, 0, len(users))
	for _, u := range users {
		out = append(out, shared.NewAdminUser(u, providers))
	}
	return out, len(out), totalCount, nil
}

func (*UserList) CanRead(*xorm.Session, web.Auth) (bool, int, error) { return true, 0, nil }
