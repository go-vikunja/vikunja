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

	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// TeamProject defines the relation between a team and a project
type TeamProject struct {
	// The unique, numeric id of this project <-> team relation.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id"`
	// The team id.
	TeamID int64 `xorm:"bigint not null INDEX" json:"team_id" param:"team"`
	// The project id.
	ProjectID int64 `xorm:"bigint not null INDEX" json:"-" param:"project"`
	// The permission this team has. 0 = Read only, 1 = Read & Write, 2 = Admin. See the docs for more details.
	Permission Permission `xorm:"bigint INDEX not null default 0" json:"permission" valid:"length(0|2)" maximum:"2" default:"0"`

	// A timestamp when this relation was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this relation was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName makes beautiful table names
func (*TeamProject) TableName() string {
	return "team_projects"
}

// TeamWithPermission represents a team, combined with permissions.
type TeamWithPermission struct {
	Team       `xorm:"extends"`
	Permission Permission `json:"permission"`
}

// ProjectTeamServiceProvider is a function type that returns a project team service instance
// This is used to avoid import cycles between models and services packages
type ProjectTeamServiceProvider func() interface {
	Create(s *xorm.Session, teamProject *TeamProject, doer web.Auth) error
	Delete(s *xorm.Session, teamProject *TeamProject) error
	GetAll(s *xorm.Session, projectID int64, doer web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error)
	Update(s *xorm.Session, teamProject *TeamProject) error
}

// projectTeamServiceProvider is the registered service provider function
var projectTeamServiceProvider ProjectTeamServiceProvider

// RegisterProjectTeamService registers a service provider for project team operations
// This should be called during application initialization by the services package
func RegisterProjectTeamService(provider ProjectTeamServiceProvider) {
	projectTeamServiceProvider = provider
}

// getProjectTeamService returns the registered project team service instance
func getProjectTeamService() interface {
	Create(s *xorm.Session, teamProject *TeamProject, doer web.Auth) error
	Delete(s *xorm.Session, teamProject *TeamProject) error
	GetAll(s *xorm.Session, projectID int64, doer web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error)
	Update(s *xorm.Session, teamProject *TeamProject) error
} {
	if projectTeamServiceProvider != nil {
		return projectTeamServiceProvider()
	}
	// This should never happen in production, only in tests that don't initialize the service
	panic("ProjectTeamService not registered - ensure services.InitializeDependencies() is called during startup")
}

// Create creates a new team <-> project relation
// @Summary Add a team to a project
// @Description Gives a team access to a project.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Project ID"
// @Param project body models.TeamProject true "The team you want to add to the project."
// @Success 201 {object} models.TeamProject "The created team<->project relation."
// @Failure 400 {object} web.HTTPError "Invalid team project object provided."
// @Failure 404 {object} web.HTTPError "The team does not exist."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id}/teams [put]
// @Deprecated: Use services.ProjectTeamService.Create instead.
func (tl *TeamProject) Create(s *xorm.Session, a web.Auth) (err error) {
	// DEPRECATED: Business logic moved to service layer
	// This method now delegates to ProjectTeamService.Create
	// Direct imports avoided to prevent cycles - get service from registry or instantiate
	service := getProjectTeamService()
	return service.Create(s, tl, a)
}

// Delete deletes a team <-> project relation based on the project & team id
// @Summary Delete a team from a project
// @Description Delets a team from a project. The team won't have access to the project anymore.
// @tags sharing
// @Produce json
// @Security JWTKeyAuth
// @Param projectID path int true "Project ID"
// @Param teamID path int true "Team ID"
// @Success 200 {object} models.Message "The team was successfully deleted."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project"
// @Failure 404 {object} web.HTTPError "Team or project does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{projectID}/teams/{teamID} [delete]
// @Deprecated: Use services.ProjectTeamService.Delete instead.
func (tl *TeamProject) Delete(s *xorm.Session, _ web.Auth) (err error) {
	// DEPRECATED: Business logic moved to service layer
	// This method now delegates to ProjectTeamService.Delete
	service := getProjectTeamService()
	return service.Delete(s, tl)
}

// ReadAll implements the method to read all teams of a project
// @Summary Get teams on a project
// @Description Returns a project with all teams which have access on a given project.
// @tags sharing
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search teams by its name."
// @Security JWTKeyAuth
// @Success 200 {array} models.TeamWithPermission "The teams with their permission."
// @Failure 403 {object} web.HTTPError "No permission to see the project."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{id}/teams [get]
// @Deprecated: Use services.ProjectTeamService.GetAll instead.
func (tl *TeamProject) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	// DEPRECATED: Business logic moved to service layer
	// This method now delegates to ProjectTeamService.GetAll
	service := getProjectTeamService()
	return service.GetAll(s, tl.ProjectID, a, search, page, perPage)
}

// Update updates a team <-> project relation
// @Summary Update a team <-> project relation
// @Description Update a team <-> project relation. Mostly used to update the permission that team has.
// @tags sharing
// @Accept json
// @Produce json
// @Param projectID path int true "Project ID"
// @Param teamID path int true "Team ID"
// @Param project body models.TeamProject true "The team you want to update."
// @Security JWTKeyAuth
// @Success 200 {object} models.TeamProject "The updated team <-> project relation."
// @Failure 403 {object} web.HTTPError "The user does not have admin-access to the project"
// @Failure 404 {object} web.HTTPError "Team or project does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{projectID}/teams/{teamID} [post]
// @Deprecated: Use services.ProjectTeamService.Update instead.
func (tl *TeamProject) Update(s *xorm.Session, _ web.Auth) (err error) {
	// DEPRECATED: Business logic moved to service layer
	// This method now delegates to ProjectTeamService.Update
	service := getProjectTeamService()
	return service.Update(s, tl)
}
