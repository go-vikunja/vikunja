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
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// ProjectDuplicateService handles the duplication of projects and all their related data.
// This service orchestrates the complex process of copying a project, including tasks,
// attachments, labels, assignees, comments, relations, views, and permissions.
type ProjectDuplicateService struct {
	DB             *xorm.Engine
	ProjectService *ProjectService
	TaskService    *TaskService
}

// NewProjectDuplicateService creates a new ProjectDuplicateService.
func NewProjectDuplicateService(db *xorm.Engine) *ProjectDuplicateService {
	return &ProjectDuplicateService{
		DB:             db,
		ProjectService: NewProjectService(db),
		TaskService:    NewTaskService(db),
	}
}

// Duplicate creates a complete copy of a project and all its related data.
// This includes tasks, attachments, labels, assignees, comments, relations, views,
// kanban data, user/team permissions, and link shares.
//
// The user needs read access to the source project and write access to the parent
// project where the new project will be created.
func (pds *ProjectDuplicateService) Duplicate(s *xorm.Session, projectID int64, parentProjectID int64, u *user.User) (*models.Project, error) {
	// Permission checks: Read access to source project
	canRead, err := pds.ProjectService.HasPermission(s, projectID, u, models.PermissionRead)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, ErrAccessDenied
	}

	// Permission checks: Write access to parent project (if specified)
	if parentProjectID != 0 {
		canCreate, err := pds.ProjectService.HasPermission(s, parentProjectID, u, models.PermissionCreate)
		if err != nil {
			return nil, err
		}
		if !canCreate {
			return nil, ErrAccessDenied
		}
	}

	// Get the source project
	sourceProject, err := models.GetProjectSimpleByID(s, projectID)
	if err != nil {
		return nil, err
	}

	log.Debugf("Duplicating project %d", projectID)

	// Create the new project
	newProject := &models.Project{
		Title:           sourceProject.Title + " - duplicate",
		Description:     sourceProject.Description,
		ParentProjectID: parentProjectID,
		OwnerID:         u.ID,
		Position:        sourceProject.Position,
		HexColor:        sourceProject.HexColor,
		IsFavorite:      false, // Reset favorite status for new project
		IsArchived:      false, // Reset archived status for new project
	}

	// Create the project using ProjectService
	createdProject, err := pds.ProjectService.Create(s, newProject, u)
	if err != nil {
		// If there is no available unique project identifier, reset it and try again
		if models.IsErrProjectIdentifierIsNotUnique(err) {
			newProject.Identifier = ""
			createdProject, err = pds.ProjectService.Create(s, newProject, u)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	log.Debugf("Duplicated project %d into new project %d", projectID, createdProject.ID)

	// Duplicate all tasks and their related data
	taskIDMap, err := pds.duplicateTasksAndRelatedData(s, projectID, createdProject.ID, u)
	if err != nil {
		return nil, err
	}

	log.Debugf("Duplicated all tasks from project %d into %d", projectID, createdProject.ID)

	// Duplicate views and kanban data
	err = pds.duplicateProjectViews(s, projectID, createdProject.ID, u, taskIDMap)
	if err != nil {
		return nil, err
	}

	log.Debugf("Duplicated all views, buckets and positions from project %d into %d", projectID, createdProject.ID)

	// Duplicate project metadata (background, permissions, shares)
	err = pds.duplicateProjectMetadata(s, projectID, createdProject.ID, u)
	if err != nil {
		return nil, err
	}

	log.Debugf("Duplicated all metadata from project %d into %d", projectID, createdProject.ID)

	// Reload the project with full details
	err = createdProject.ReadOne(s, u)
	if err != nil {
		return nil, err
	}

	return createdProject, nil
}

// duplicateTasksAndRelatedData handles the duplication of all tasks and their related data.
// Returns a map of old task ID -> new task ID for use in other duplication functions.
func (pds *ProjectDuplicateService) duplicateTasksAndRelatedData(s *xorm.Session, sourceProjectID int64, targetProjectID int64, u *user.User) (map[int64]int64, error) {
	// This method will be implemented in the next step
	// For now, return an empty map to allow compilation
	return make(map[int64]int64), nil
}

// duplicateProjectViews handles the duplication of project views and kanban data.
func (pds *ProjectDuplicateService) duplicateProjectViews(s *xorm.Session, sourceProjectID int64, targetProjectID int64, u *user.User, taskIDMap map[int64]int64) error {
	// This method will be implemented in a later step
	return nil
}

// duplicateProjectMetadata handles the duplication of project background, permissions, and shares.
func (pds *ProjectDuplicateService) duplicateProjectMetadata(s *xorm.Session, sourceProjectID int64, targetProjectID int64, u *user.User) error {
	// This method will be implemented in a later step
	return nil
}
