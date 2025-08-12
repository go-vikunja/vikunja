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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
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
func (tl *TeamProject) Create(s *xorm.Session, a web.Auth) (err error) {

	// Check if the permissions are valid
	if err = tl.Permission.isValid(); err != nil {
		return
	}

	// Check if the team exists
	team, err := GetTeamByID(s, tl.TeamID)
	if err != nil {
		return err
	}

	// Check if the project exists
	l, err := GetProjectSimpleByID(s, tl.ProjectID)
	if err != nil {
		return err
	}

	// Check if the team is already on the project
	exists, err := s.Where("team_id = ?", tl.TeamID).
		And("project_id = ?", tl.ProjectID).
		Get(&TeamProject{})
	if err != nil {
		return
	}
	if exists {
		return ErrTeamAlreadyHasAccess{tl.TeamID, tl.ProjectID}
	}

	// Insert the new team
	tl.ID = 0
	_, err = s.Insert(tl)
	if err != nil {
		return err
	}

	err = events.Dispatch(&ProjectSharedWithTeamEvent{
		Project: l,
		Team:    team,
		Doer:    a,
	})
	if err != nil {
		return err
	}

	err = updateProjectLastUpdated(s, l)
	return
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
func (tl *TeamProject) Delete(s *xorm.Session, _ web.Auth) (err error) {

	// Check if the team exists
	_, err = GetTeamByID(s, tl.TeamID)
	if err != nil {
		return
	}

	// Check if the team has access to the project
	has, err := s.
		Where("team_id = ? AND project_id = ?", tl.TeamID, tl.ProjectID).
		Get(&TeamProject{})
	if err != nil {
		return
	}
	if !has {
		return ErrTeamDoesNotHaveAccessToProject{TeamID: tl.TeamID, ProjectID: tl.ProjectID}
	}

	// Delete the relation
	_, err = s.Where("team_id = ?", tl.TeamID).
		And("project_id = ?", tl.ProjectID).
		Delete(&TeamProject{})
	if err != nil {
		return err
	}

	err = updateProjectLastUpdated(s, &Project{ID: tl.ProjectID})
	return
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
func (tl *TeamProject) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, totalItems int64, err error) {
	// Check if the user can read the project
	l := &Project{ID: tl.ProjectID}
	canRead, _, err := l.CanRead(s, a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrNeedToHaveProjectReadAccess{ProjectID: tl.ProjectID, UserID: a.GetID()}
	}

	limit, start := getLimitFromPageIndex(page, perPage)

	// Get the teams
	all := []*TeamWithPermission{}
	query := s.
		Table("teams").
		Join("INNER", "team_projects", "team_id = teams.id").
		Where("team_projects.project_id = ?", tl.ProjectID).
		Where(db.ILIKE("teams.name", search))
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&all)
	if err != nil {
		return nil, 0, 0, err
	}

	teams := []*Team{}
	for i := range all {
		teams = append(teams, &all[i].Team)
	}

	err = addMoreInfoToTeams(s, teams)
	if err != nil {
		return
	}

	totalItems, err = s.
		Table("teams").
		Join("INNER", "team_projects", "team_id = teams.id").
		Where("team_projects.project_id = ?", tl.ProjectID).
		Where("teams.name LIKE ?", "%"+search+"%").
		Count(&TeamWithPermission{})
	if err != nil {
		return nil, 0, 0, err
	}

	return all, len(all), totalItems, err
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
func (tl *TeamProject) Update(s *xorm.Session, _ web.Auth) (err error) {

	// Check if the permission is valid
	if err := tl.Permission.isValid(); err != nil {
		return err
	}

	_, err = s.
		Where("project_id = ? AND team_id = ?", tl.ProjectID, tl.TeamID).
		Cols("permission").
		Update(tl)
	if err != nil {
		return err
	}

	err = updateProjectLastUpdated(s, &Project{ID: tl.ProjectID})
	return
}
