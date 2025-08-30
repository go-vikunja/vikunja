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

package services

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"xorm.io/xorm"
)

// ProjectService is a service for managing projects.
type ProjectService struct {
	DB *xorm.Engine
}

// NewProjectService creates a new ProjectService.
func NewProjectService(db *xorm.Engine) *ProjectService {
	return &ProjectService{DB: db}
}

// HasPermission checks if a user has a specific permission on a project.
func (ps *ProjectService) HasPermission(s *xorm.Session, projectID int64, u *user.User, permission models.Permission) (bool, error) {
	// For now, delegate to the existing model method
	// TODO: Move the permission logic to the service layer
	project := &models.Project{ID: projectID}

	switch permission {
	case models.PermissionRead:
		canRead, _, err := project.CanRead(s, u)
		return canRead, err
	case models.PermissionWrite:
		return project.CanWrite(s, u)
	default:
		return false, nil
	}
}
