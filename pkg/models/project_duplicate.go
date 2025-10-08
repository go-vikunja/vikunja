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

// ProjectDuplicateServiceProvider is the interface for the project duplicate service
// This interface allows models to call service layer methods without import cycles
type ProjectDuplicateServiceProvider interface {
	Duplicate(s *xorm.Session, projectID int64, parentProjectID int64, u *user.User) (*Project, error)
}

var projectDuplicateService ProjectDuplicateServiceProvider

// RegisterProjectDuplicateService registers the project duplicate service
func RegisterProjectDuplicateService(service ProjectDuplicateServiceProvider) {
	projectDuplicateService = service
}

// getProjectDuplicateService returns the registered project duplicate service
func getProjectDuplicateService() ProjectDuplicateServiceProvider {
	if projectDuplicateService == nil {
		panic("ProjectDuplicateService not registered. Make sure to call RegisterProjectDuplicateService during initialization.")
	}
	return projectDuplicateService
}

// ProjectDuplicate holds everything needed to duplicate a project
type ProjectDuplicate struct {
	// The project id of the project to duplicate
	ProjectID int64 `json:"-" param:"projectid"`
	// The target parent project
	ParentProjectID int64 `json:"parent_project_id,omitempty"`

	// The copied project
	Project *Project `json:"duplicated_project,omitempty"`

	web.Permissions `json:"-"`
	web.CRUDable    `json:"-"`
}

// CanCreate checks if a user has the permission to duplicate a project
// @Deprecated Use ProjectDuplicateService.Duplicate instead (permission checks are built-in)
func (pd *ProjectDuplicate) CanCreate(s *xorm.Session, a web.Auth) (canCreate bool, err error) {
	// Project Exists + user has read access to project
	pd.Project = &Project{ID: pd.ProjectID}
	canRead, _, err := pd.Project.CanRead(s, a)
	if err != nil || !canRead {
		return canRead, err
	}

	if pd.ParentProjectID == 0 { // no parent project
		return canRead, err
	}

	// Parent project exists + user has write access to it (-> can create new projects)
	parent := &Project{ID: pd.ParentProjectID}
	return parent.CanCreate(s, a)
}

// Create duplicates a project
// @Summary Duplicate an existing project
// @Description Copies the project, tasks, files, kanban data, assignees, comments, attachments, labels, relations, backgrounds, user/team permissions and link shares from one project to a new one. The user needs read access in the project and write access in the parent of the new project.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param projectID path int true "The project ID to duplicate"
// @Param project body models.ProjectDuplicate true "The target parent project which should hold the copied project."
// @Success 201 {object} models.ProjectDuplicate "The created project."
// @Failure 400 {object} web.HTTPError "Invalid project duplicate object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project or its parent."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{projectID}/duplicate [put]
// @Deprecated Use ProjectDuplicateService.Duplicate instead
func (pd *ProjectDuplicate) Create(s *xorm.Session, doer web.Auth) (err error) {
	// Delegate to service layer
	service := getProjectDuplicateService()

	// Get user from auth
	doerUser, err := GetUserOrLinkShareUser(s, doer)
	if err != nil {
		return err
	}

	// Call service layer
	duplicatedProject, err := service.Duplicate(s, pd.ProjectID, pd.ParentProjectID, doerUser)
	if err != nil {
		return err
	}

	// Set the duplicated project in the response
	pd.Project = duplicatedProject

	return nil
}
