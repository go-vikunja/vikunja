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
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"xorm.io/xorm"
)

// ProjectUserService handles all operations related to project-user permissions
type ProjectUserService struct {
	DB             *xorm.Engine
	ProjectService *ProjectService
}

// NewProjectUserService creates a new ProjectUserService
func NewProjectUserService(db *xorm.Engine) *ProjectUserService {
	return &ProjectUserService{
		DB:             db,
		ProjectService: NewProjectService(db),
	}
}

// Create adds a user to a project with the specified permission level.
// Returns error if the user already has access, is the owner, or permission is invalid.
func (pus *ProjectUserService) Create(s *xorm.Session, projectUser *models.ProjectUser, doer *user.User) error {
	// Check if the permission is valid
	if err := projectUser.Permission.IsValid(); err != nil {
		return err
	}

	// Check if the project exists
	project, err := models.GetProjectSimpleByID(s, projectUser.ProjectID)
	if err != nil {
		return err
	}

	// Check if the user exists
	targetUser, err := user.GetUserByUsername(s, projectUser.Username)
	if err != nil {
		return err
	}
	projectUser.UserID = targetUser.ID

	// Check if the user already has access or is owner of that project
	// We explicitly DON'T check for teams here
	if project.OwnerID == projectUser.UserID {
		return models.ErrUserAlreadyHasAccess{UserID: projectUser.UserID, ProjectID: projectUser.ProjectID}
	}

	exist, err := s.Where("project_id = ? AND user_id = ?", projectUser.ProjectID, projectUser.UserID).
		Get(&models.ProjectUser{})
	if err != nil {
		return err
	}
	if exist {
		return models.ErrUserAlreadyHasAccess{UserID: projectUser.UserID, ProjectID: projectUser.ProjectID}
	}

	// Insert the new project-user relation
	projectUser.ID = 0
	_, err = s.Insert(projectUser)
	if err != nil {
		return err
	}

	// Dispatch event
	err = events.Dispatch(&models.ProjectSharedWithUserEvent{
		Project: project,
		User:    targetUser,
		Doer:    doer,
	})
	if err != nil {
		return err
	}

	// Update project's last updated timestamp
	err = models.UpdateProjectLastUpdated(s, project)
	return err
}

// Delete removes a user's access to a project.
// Returns error if the user doesn't have access to the project.
func (pus *ProjectUserService) Delete(s *xorm.Session, projectUser *models.ProjectUser) error {
	if projectUser.UserID == 0 {
		// Check if the user exists
		targetUser, err := user.GetUserByUsername(s, projectUser.Username)
		if err != nil {
			return err
		}
		projectUser.UserID = targetUser.ID
	}

	// Check if the user has access to the project
	has, err := s.
		Where("user_id = ? AND project_id = ?", projectUser.UserID, projectUser.ProjectID).
		Get(&models.ProjectUser{})
	if err != nil {
		return err
	}
	if !has {
		return models.ErrUserDoesNotHaveAccessToProject{ProjectID: projectUser.ProjectID, UserID: projectUser.UserID}
	}

	// Delete the project-user relation
	_, err = s.
		Where("user_id = ? AND project_id = ?", projectUser.UserID, projectUser.ProjectID).
		Delete(&models.ProjectUser{})
	if err != nil {
		return err
	}

	// Update project's last updated timestamp
	err = models.UpdateProjectLastUpdated(s, &models.Project{ID: projectUser.ProjectID})
	return err
}

// GetAll retrieves all users who have access to a project with their permission levels.
// Supports pagination and search by username.
func (pus *ProjectUserService) GetAll(s *xorm.Session, projectID int64, doer *user.User, search string, page int, perPage int) (users []*models.UserWithPermission, resultCount int, totalItems int64, err error) {
	// Check if the user has access to the project
	canRead, err := pus.ProjectService.HasPermission(s, projectID, doer, models.PermissionRead)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, models.ErrNeedToHaveProjectReadAccess{UserID: doer.ID, ProjectID: projectID}
	}

	limit, start := getLimitFromPageIndex(page, perPage)

	// Get all users with their permissions
	users = []*models.UserWithPermission{}
	query := s.
		Select("users.*, users_projects.permission").
		Join("INNER", "users_projects", "users_projects.user_id = users.id").
		Where("users_projects.project_id = ?", projectID).
		Where(db.ILIKE("users.username", search))
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&users)
	if err != nil {
		return nil, 0, 0, err
	}

	// Obfuscate all user emails for privacy
	for _, u := range users {
		u.Email = ""
	}

	// Get total count
	totalItems, err = s.
		Join("INNER", "users_projects", "user_id = users.id").
		Where("users_projects.project_id = ?", projectID).
		Where("users.username LIKE ?", "%"+search+"%").
		Count(&models.UserWithPermission{})
	if err != nil {
		return nil, 0, 0, err
	}

	return users, len(users), totalItems, nil
}

// Update modifies the permission level of a user's access to a project.
// Returns error if the permission is invalid or user doesn't have access.
func (pus *ProjectUserService) Update(s *xorm.Session, projectUser *models.ProjectUser) error {
	if projectUser.UserID == 0 {
		// Check if the user exists
		targetUser, err := user.GetUserByUsername(s, projectUser.Username)
		if err != nil {
			return err
		}
		projectUser.UserID = targetUser.ID
	}

	// Check if the permission is valid
	if err := projectUser.Permission.IsValid(); err != nil {
		return err
	}

	// Update the permission
	_, err := s.
		Where("project_id = ? AND user_id = ?", projectUser.ProjectID, projectUser.UserID).
		Cols("permission").
		Update(projectUser)
	if err != nil {
		return err
	}

	// Update project's last updated timestamp
	err = models.UpdateProjectLastUpdated(s, &models.Project{ID: projectUser.ProjectID})
	return err
}

// HasAccess checks if a user has any level of access to a project.
// Returns true if the user has direct access (not via teams).
func (pus *ProjectUserService) HasAccess(s *xorm.Session, projectID int64, userID int64) (bool, error) {
	has, err := s.
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Exist(&models.ProjectUser{})
	return has, err
}

// GetPermission retrieves the permission level a user has for a project.
// Returns PermissionUnknown if the user doesn't have direct access.
func (pus *ProjectUserService) GetPermission(s *xorm.Session, projectID int64, userID int64) (models.Permission, error) {
	pu := &models.ProjectUser{}
	has, err := s.
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Get(pu)
	if err != nil {
		return models.PermissionUnknown, err
	}
	if !has {
		return models.PermissionUnknown, nil
	}
	return pu.Permission, nil
}
