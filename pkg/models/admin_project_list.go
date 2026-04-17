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

package models

import (
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// AdminProjectList is the CRUDable wrapper the admin list-projects route
// plugs into the generic web handler. It only implements ReadAll;
// everything else is gated by the RequireInstanceAdmin middleware.
type AdminProjectList struct {
	IsArchived bool `query:"is_archived"`
}

// ReadAll returns every project on the instance, archived included.
// @Summary List projects (admin)
// @Description Paginated list of every project on the instance, regardless of ownership.
// @tags admin
// @Produce json
// @Security JWTKeyAuth
// @Param page query int false "Page number, defaults to 1."
// @Param per_page query int false "Items per page, defaults to the service setting."
// @Param s query string false "Search projects by title, description or identifier."
// @Success 200 {array} models.Project
// @Failure 404 {object} web.HTTPError
// @Router /admin/projects [get]
func (l *AdminProjectList) ReadAll(s *xorm.Session, _ web.Auth, search string, page, perPage int) (interface{}, int, int64, error) {
	return ListAllProjects(s, search, page, perPage, true)
}

func (*AdminProjectList) ReadOne(*xorm.Session, web.Auth) error              { return nil }
func (*AdminProjectList) Create(*xorm.Session, web.Auth) error               { return nil }
func (*AdminProjectList) Update(*xorm.Session, web.Auth) error               { return nil }
func (*AdminProjectList) Delete(*xorm.Session, web.Auth) error               { return nil }
func (*AdminProjectList) CanCreate(*xorm.Session, web.Auth) (bool, error)    { return false, nil }
func (*AdminProjectList) CanDelete(*xorm.Session, web.Auth) (bool, error)    { return false, nil }
func (*AdminProjectList) CanUpdate(*xorm.Session, web.Auth) (bool, error)    { return false, nil }
func (*AdminProjectList) CanRead(*xorm.Session, web.Auth) (bool, int, error) { return true, 0, nil }
