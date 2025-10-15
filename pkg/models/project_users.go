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
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// ProjectUser represents a project <-> user relation
type ProjectUser struct {
	// The unique, numeric id of this project <-> user relation.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id"`
	// The username.
	Username string `xorm:"-" json:"username" param:"user"`
	// Used internally to reference the user
	UserID int64 `xorm:"bigint not null INDEX" json:"-"`
	// The project id.
	ProjectID int64 `xorm:"bigint not null INDEX" json:"-" param:"project"`
	// The permission this user has. 0 = Read only, 1 = Read & Write, 2 = Admin. See the docs for more details.
	Permission Permission `xorm:"bigint INDEX not null default 0" json:"permission" valid:"length(0|2)" maximum:"2" default:"0"`

	// A timestamp when this relation was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this relation was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName is the table name for ProjectUser
func (*ProjectUser) TableName() string {
	return "users_projects"
}

// UserWithPermission represents a user in combination with the permission it can have on a project
type UserWithPermission struct {
	user.User  `xorm:"extends"`
	Permission Permission `json:"permission"`
}

// ProjectUserServiceProvider is a function type that returns a project user service instance
// This is used to avoid import cycles between models and services packages
type ProjectUserServiceProvider func() interface {
	Create(s *xorm.Session, projectUser *ProjectUser, doer *user.User) error
	Delete(s *xorm.Session, projectUser *ProjectUser) error
	GetAll(s *xorm.Session, projectID int64, doer *user.User, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error)
	Update(s *xorm.Session, projectUser *ProjectUser) error
}

// projectUserServiceProvider is the registered service provider function
var projectUserServiceProvider ProjectUserServiceProvider

// RegisterProjectUserService registers a service provider for project user operations
// This should be called during application initialization by the services package
func RegisterProjectUserService(provider ProjectUserServiceProvider) {
	projectUserServiceProvider = provider
}

// getProjectUserService returns the registered project user service instance
func getProjectUserService() interface {
	Create(s *xorm.Session, projectUser *ProjectUser, doer *user.User) error
	Delete(s *xorm.Session, projectUser *ProjectUser) error
	GetAll(s *xorm.Session, projectID int64, doer *user.User, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error)
	Update(s *xorm.Session, projectUser *ProjectUser) error
} {
	if projectUserServiceProvider != nil {
		return projectUserServiceProvider()
	}
	// This should never happen in production, only in tests that don't initialize the service
	panic("ProjectUserService not registered - ensure services.InitializeDependencies() is called during startup")
}

// Create creates a new project <-> user relation
// @Summary Add a user to a project
// @Description Gives a user access to a project.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Project ID"
// @Param project body models.ProjectUser true "The user you want to add to the project."
// @Success 201 {object} models.ProjectUser "The created user<->project relation."
// @Failure 400 {object} web.HTTPError "Invalid user project object provided."
// @Failure 404 {object} web.HTTPError "The user does not exist."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id}/users [put]
// @Deprecated: Use services.ProjectUserService.Create instead.
func (lu *ProjectUser) Create(s *xorm.Session, a web.Auth) (err error) {
	// DEPRECATED: Business logic moved to service layer
	// This method now delegates to ProjectUserService.Create
	service := getProjectUserService()

	// Convert web.Auth to *user.User (supports both regular users and link shares)
	var doer *user.User
	if a != nil {
		doer, err = GetUserOrLinkShareUser(s, a)
		if err != nil {
			return err
		}
	}

	return service.Create(s, lu, doer)
}

// Delete deletes a project <-> user relation
// @Summary Delete a user from a project
// @Description Delets a user from a project. The user won't have access to the project anymore.
// @tags sharing
// @Produce json
// @Security JWTKeyAuth
// @Param projectID path int true "Project ID"
// @Param userID path int true "User ID"
// @Success 200 {object} models.Message "The user was successfully removed from the project."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 404 {object} web.HTTPError "user or project does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{projectID}/users/{userID} [delete]
// @Deprecated: Use services.ProjectUserService.Delete instead.
func (lu *ProjectUser) Delete(s *xorm.Session, _ web.Auth) (err error) {
	// DEPRECATED: Business logic moved to service layer
	// This method now delegates to ProjectUserService.Delete
	service := getProjectUserService()
	return service.Delete(s, lu)
}

// ReadAll gets all users who have access to a project
// @Summary Get users on a project
// @Description Returns a project with all users which have access on a given project.
// @tags sharing
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search users by its name."
// @Security JWTKeyAuth
// @Success 200 {array} models.UserWithPermission "The users with the permission they have."
// @Failure 403 {object} web.HTTPError "No permission to see the project."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id}/users [get]
// @Deprecated: Use services.ProjectUserService.GetAll instead.
func (lu *ProjectUser) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	// DEPRECATED: Business logic moved to service layer
	// This method now delegates to ProjectUserService.GetAll
	service := getProjectUserService()

	// Convert web.Auth to *user.User (supports both regular users and link shares)
	var doer *user.User
	if a != nil {
		doer, err = GetUserOrLinkShareUser(s, a)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	return service.GetAll(s, lu.ProjectID, doer, search, page, perPage)
}

// Update updates a user <-> project relation
// @Summary Update a user <-> project relation
// @Description Update a user <-> project relation. Mostly used to update the permission that user has.
// @tags sharing
// @Accept json
// @Produce json
// @Param projectID path int true "Project ID"
// @Param userID path int true "User ID"
// @Param project body models.ProjectUser true "The user you want to update."
// @Security JWTKeyAuth
// @Success 200 {object} models.ProjectUser "The updated user <-> project relation."
// @Failure 403 {object} web.HTTPError "The user does not have admin-access to the project"
// @Failure 404 {object} web.HTTPError "User or project does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{projectID}/users/{userID} [post]
// @Deprecated: Use services.ProjectUserService.Update instead.
func (pu *ProjectUser) Update(s *xorm.Session, _ web.Auth) (err error) {
	// DEPRECATED: Business logic moved to service layer
	// This method now delegates to ProjectUserService.Update
	service := getProjectUserService()
	return service.Update(s, pu)
}

// Permission delegation functions
var (
	ProjectUserCanCreateFunc func(s *xorm.Session, projectID int64, a web.Auth) (bool, error)
	ProjectUserCanUpdateFunc func(s *xorm.Session, projectID int64, a web.Auth) (bool, error)
	ProjectUserCanDeleteFunc func(s *xorm.Session, projectID int64, a web.Auth) (bool, error)
	ProjectUserCanReadFunc   func(s *xorm.Session, projectID int64, a web.Auth) (bool, error)
)

func (lu *ProjectUser) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if ProjectUserCanCreateFunc != nil {
		return ProjectUserCanCreateFunc(s, lu.ProjectID, a)
	}
	panic("ProjectUserCanCreateFunc not set")
}

func (lu *ProjectUser) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	if ProjectUserCanDeleteFunc != nil {
		return ProjectUserCanDeleteFunc(s, lu.ProjectID, a)
	}
	panic("ProjectUserCanDeleteFunc not set")
}

func (lu *ProjectUser) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	if ProjectUserCanUpdateFunc != nil {
		return ProjectUserCanUpdateFunc(s, lu.ProjectID, a)
	}
	panic("ProjectUserCanUpdateFunc not set")
}

func (lu *ProjectUser) CanRead(s *xorm.Session, a web.Auth) (canRead bool, maxPermission int, err error) {
	if ProjectUserCanReadFunc != nil {
		canRead, err = ProjectUserCanReadFunc(s, lu.ProjectID, a)
		return canRead, int(PermissionAdmin), err
	}
	panic("ProjectUserCanReadFunc not set")
}
