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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// InitProjectTeamService sets up dependency injection for project team-related model functions.
// This function must be called during initialization to enable service layer delegation.
func InitProjectTeamService() {
	// Set up permission delegation (T-PERM-011)
	models.TeamProjectCanCreateFunc = func(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
		pts := NewProjectTeamService(s.Engine())
		return pts.CanCreate(s, projectID, a)
	}
	models.TeamProjectCanUpdateFunc = func(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
		pts := NewProjectTeamService(s.Engine())
		return pts.CanUpdate(s, projectID, a)
	}
	models.TeamProjectCanDeleteFunc = func(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
		pts := NewProjectTeamService(s.Engine())
		return pts.CanDelete(s, projectID, a)
	}
	models.TeamProjectCanReadFunc = func(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
		pts := NewProjectTeamService(s.Engine())
		return pts.CanRead(s, projectID, a)
	}
}

// ProjectTeamService handles all operations related to project-team permissions
type ProjectTeamService struct {
	DB       *xorm.Engine
	Registry *ServiceRegistry
}

// NewProjectTeamService creates a new ProjectTeamService
// Deprecated: Use ServiceRegistry.ProjectTeams() instead.
func NewProjectTeamService(db *xorm.Engine) *ProjectTeamService {
	registry := NewServiceRegistry(db)
	return registry.ProjectTeams()
}

// Create adds a team to a project with the specified permission level.
// Returns error if the team already has access or permission is invalid.
func (pts *ProjectTeamService) Create(s *xorm.Session, teamProject *models.TeamProject, doer web.Auth) error {
	// Check if the permission is valid
	if err := teamProject.Permission.IsValid(); err != nil {
		return err
	}

	// Check if the team exists
	team, err := pts.Registry.Team().GetByID(s, teamProject.TeamID)
	if err != nil {
		return err
	}

	// Check if the project exists
	project, err := pts.Registry.Project().GetByIDSimple(s, teamProject.ProjectID)
	if err != nil {
		return err
	}

	// Check if the team already has access to the project
	exists, err := s.Where("team_id = ?", teamProject.TeamID).
		And("project_id = ?", teamProject.ProjectID).
		Get(&models.TeamProject{})
	if err != nil {
		return err
	}
	if exists {
		return models.ErrTeamAlreadyHasAccess{TeamID: teamProject.TeamID, ID: teamProject.ProjectID}
	}

	// Insert the new team-project relation
	teamProject.ID = 0
	_, err = s.Insert(teamProject)
	if err != nil {
		return err
	}

	// Dispatch event
	err = events.Dispatch(&models.ProjectSharedWithTeamEvent{
		Project: project,
		Team:    team,
		Doer:    doer,
	})
	if err != nil {
		return err
	}

	// Update project's last updated timestamp
	err = models.UpdateProjectLastUpdated(s, project)
	return err
}

// Delete removes a team's access to a project.
// Returns error if the team doesn't have access to the project.
func (pts *ProjectTeamService) Delete(s *xorm.Session, teamProject *models.TeamProject) error {
	// Check if the team exists
	_, err := pts.Registry.Team().GetByID(s, teamProject.TeamID)
	if err != nil {
		return err
	}

	// Check if the team has access to the project
	has, err := s.
		Where("team_id = ? AND project_id = ?", teamProject.TeamID, teamProject.ProjectID).
		Get(&models.TeamProject{})
	if err != nil {
		return err
	}
	if !has {
		return models.ErrTeamDoesNotHaveAccessToProject{TeamID: teamProject.TeamID, ProjectID: teamProject.ProjectID}
	}

	// Delete the team-project relation
	_, err = s.
		Where("team_id = ?", teamProject.TeamID).
		And("project_id = ?", teamProject.ProjectID).
		Delete(&models.TeamProject{})
	if err != nil {
		return err
	}

	// Update project's last updated timestamp
	err = models.UpdateProjectLastUpdated(s, &models.Project{ID: teamProject.ProjectID})
	return err
}

// GetAll retrieves all teams that have access to a project with their permission levels.
// Supports pagination and search by team name.
func (pts *ProjectTeamService) GetAll(s *xorm.Session, projectID int64, doer web.Auth, search string, page int, perPage int) (teams []*models.TeamWithPermission, resultCount int, totalItems int64, err error) {
	// Check if the user/link share has access to the project
	project := &models.Project{ID: projectID}
	canRead, _, err := project.CanRead(s, doer)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, models.ErrNeedToHaveProjectReadAccess{UserID: doer.GetID(), ProjectID: projectID}
	}

	limit, start := getLimitFromPageIndex(page, perPage)

	// Get all teams with their permissions
	teams = []*models.TeamWithPermission{}
	query := s.
		Table("teams").
		Join("INNER", "team_projects", "team_id = teams.id").
		Where("team_projects.project_id = ?", projectID).
		Where(db.ILIKE("teams.name", search))
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&teams)
	if err != nil {
		return nil, 0, 0, err
	}

	// Extract teams for additional info loading
	teamList := []*models.Team{}
	for i := range teams {
		teamList = append(teamList, &teams[i].Team)
	}

	// Add more info to teams (members, created_by, etc.)
	err = models.AddMoreInfoToTeams(s, teamList)
	if err != nil {
		return nil, 0, 0, err
	}

	// Get total count
	totalItems, err = s.
		Table("teams").
		Join("INNER", "team_projects", "team_id = teams.id").
		Where("team_projects.project_id = ?", projectID).
		Where(db.ILIKE("teams.name", search)).
		Count(&models.TeamWithPermission{})
	if err != nil {
		return nil, 0, 0, err
	}

	return teams, len(teams), totalItems, nil
}

// Update modifies the permission level of a team on a project.
// Returns error if permission is invalid.
func (pts *ProjectTeamService) Update(s *xorm.Session, teamProject *models.TeamProject) error {
	// Check if the permission is valid
	if err := teamProject.Permission.IsValid(); err != nil {
		return err
	}

	// Update the permission
	_, err := s.
		Where("project_id = ? AND team_id = ?", teamProject.ProjectID, teamProject.TeamID).
		Cols("permission").
		Update(teamProject)
	if err != nil {
		return err
	}

	// Update project's last updated timestamp
	err = models.UpdateProjectLastUpdated(s, &models.Project{ID: teamProject.ProjectID})
	return err
}

// HasAccess checks if a team has direct access to a project (not through other means).
func (pts *ProjectTeamService) HasAccess(s *xorm.Session, projectID int64, teamID int64) (bool, error) {
	exists, err := s.
		Where("team_id = ? AND project_id = ?", teamID, projectID).
		Exist(&models.TeamProject{})
	return exists, err
}

// GetPermission retrieves the permission level a team has for a project.
// Returns 0 and nil error if the team doesn't have direct access.
func (pts *ProjectTeamService) GetPermission(s *xorm.Session, projectID int64, teamID int64) (models.Permission, error) {
	tp := &models.TeamProject{}
	exists, err := s.
		Where("team_id = ? AND project_id = ?", teamID, projectID).
		Get(tp)
	if err != nil {
		return models.PermissionRead, err
	}
	if !exists {
		return models.PermissionRead, nil
	}
	return tp.Permission, nil
}

// Permission Methods (T-PERM-011)

// CanCreate checks if the user can create a team <-> project relation.
// Requires admin permission on the project.
// MIGRATION: Migrated from models.TeamProject.CanCreate
func (pts *ProjectTeamService) CanCreate(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	// Link shares aren't allowed to do anything
	if _, isLinkShare := a.(*models.LinkSharing); isLinkShare {
		return false, nil
	}

	return pts.Registry.Project().IsAdmin(s, projectID, a)
}

// CanUpdate checks if the user can update a team <-> project relation.
// Requires admin permission on the project.
// MIGRATION: Migrated from models.TeamProject.CanUpdate
func (pts *ProjectTeamService) CanUpdate(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	// Link shares aren't allowed to do anything
	if _, isLinkShare := a.(*models.LinkSharing); isLinkShare {
		return false, nil
	}

	return pts.Registry.Project().IsAdmin(s, projectID, a)
}

// CanDelete checks if the user can delete a team <-> project relation.
// Requires admin permission on the project.
// MIGRATION: Migrated from models.TeamProject.CanDelete
func (pts *ProjectTeamService) CanDelete(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	// Link shares aren't allowed to do anything
	if _, isLinkShare := a.(*models.LinkSharing); isLinkShare {
		return false, nil
	}

	return pts.Registry.Project().IsAdmin(s, projectID, a)
}

// CanRead checks if the user can read team <-> project relations.
// Requires read permission on the project.
// MIGRATION: Migrated from models.TeamProject (implicit read check)
func (pts *ProjectTeamService) CanRead(s *xorm.Session, projectID int64, a web.Auth) (bool, error) {
	canRead, _, err := pts.Registry.Project().CanRead(s, projectID, a)
	return canRead, err
}
