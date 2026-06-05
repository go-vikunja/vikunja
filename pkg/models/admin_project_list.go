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
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// AdminProjectList overrides ReadAll to return every project on the instance;
// non-ReadAll methods inherit from Project and are gated by RequireInstanceAdmin.
type AdminProjectList struct {
	Project
}

// ReassignProjectOwner refuses owners scheduled for deletion because DeleteUser cascades to their projects.
func ReassignProjectOwner(s *xorm.Session, projectID, newOwnerID int64) (*Project, error) {
	p, err := GetProjectSimpleByID(s, projectID)
	if err != nil {
		return nil, err
	}

	newOwner, err := user.GetUserByID(s, newOwnerID)
	if err != nil {
		return nil, err
	}
	if !newOwner.DeletionScheduledAt.IsZero() {
		return nil, ErrInvalidData{Message: "new owner is scheduled for deletion"}
	}

	p.OwnerID = newOwnerID
	if _, err := s.ID(p.ID).Cols("owner_id").Update(p); err != nil {
		return nil, err
	}
	p.Owner = newOwner
	return p, nil
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
