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
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/builder"
	"xorm.io/xorm"
)

func init() {
	InitTeamService()
}

// TeamService handles all operations related to teams
type TeamService struct {
	DB *xorm.Engine
}

// NewTeamService creates a new TeamService
func NewTeamService(db *xorm.Engine) *TeamService {
	return &TeamService{
		DB: db,
	}
}

// Create creates a new team with the doer as admin
func (ts *TeamService) Create(s *xorm.Session, team *models.Team, doer *user.User, firstUserShouldBeAdmin bool) (*models.Team, error) {
	// Check if we have a name
	if team.Name == "" {
		return nil, models.ErrTeamNameCannotBeEmpty{}
	}

	team.ID = 0
	team.CreatedByID = doer.ID
	team.CreatedBy = doer

	_, err := s.Insert(team)
	if err != nil {
		return nil, err
	}

	// Add the creator as team member
	tm := &models.TeamMember{
		TeamID:   team.ID,
		Username: doer.Username,
		UserID:   doer.ID,
		Admin:    firstUserShouldBeAdmin,
	}

	_, err = s.Insert(tm)
	if err != nil {
		return nil, err
	}

	err = events.Dispatch(&models.TeamCreatedEvent{
		Team: team,
		Doer: doer,
	})
	if err != nil {
		return nil, err
	}

	// Load full team details
	fullTeam, err := ts.GetByID(s, team.ID)
	if err != nil {
		return nil, err
	}

	return fullTeam, nil
}

// GetByID retrieves a team by its ID with full details
func (ts *TeamService) GetByID(s *xorm.Session, teamID int64) (*models.Team, error) {
	if teamID < 1 {
		return nil, models.ErrTeamDoesNotExist{TeamID: teamID}
	}

	team := &models.Team{}
	exists, err := s.Where("id = ?", teamID).Get(team)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, models.ErrTeamDoesNotExist{TeamID: teamID}
	}

	// Add more info (members, created_by)
	teams := []*models.Team{team}
	err = ts.AddDetailsToTeams(s, teams)
	if err != nil {
		return nil, err
	}

	return team, nil
}

// Get is an alias for GetByID for consistency with other services
func (ts *TeamService) Get(s *xorm.Session, teamID int64) (*models.Team, error) {
	return ts.GetByID(s, teamID)
}

// GetByIDSimple retrieves a team by ID without loading additional details
// This is useful for permission checks and other operations that don't need full team data
func (ts *TeamService) GetByIDSimple(s *xorm.Session, teamID int64) (*models.Team, error) {
	if teamID < 1 {
		return nil, models.ErrTeamDoesNotExist{TeamID: teamID}
	}

	team := &models.Team{}
	exists, err := s.Where("id = ?", teamID).Get(team)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, models.ErrTeamDoesNotExist{TeamID: teamID}
	}

	return team, nil
}

// GetAll retrieves all teams the user has access to
func (ts *TeamService) GetAll(s *xorm.Session, auth web.Auth, search string, page int, perPage int, includePublic bool) (teams []*models.Team, resultCount int, totalItems int64, err error) {
	if _, is := auth.(*models.LinkSharing); is {
		return nil, 0, 0, models.ErrGenericForbidden{}
	}

	limit, start := getLimitFromPageIndex(page, perPage)
	teams = []*models.Team{}

	query := s.Distinct("teams.*").
		Table("teams").
		Join("INNER", "team_members", "team_members.team_id = teams.id").
		Where(db.ILIKE("teams.name", search))

	// If public teams are enabled, we want to include them in the result
	if config.ServiceEnablePublicTeams.GetBool() && includePublic {
		query = query.Where(
			builder.Or(
				builder.Eq{"teams.is_public": true},
				builder.Eq{"team_members.user_id": auth.GetID()},
			),
		)
	} else {
		query = query.Where("team_members.user_id = ?", auth.GetID())
	}

	if limit > 0 {
		query = query.Limit(limit, start)
	}

	err = query.Find(&teams)
	if err != nil {
		return nil, 0, 0, err
	}

	err = ts.AddDetailsToTeams(s, teams)
	if err != nil {
		return nil, 0, 0, err
	}

	numberOfTotalItems, err := s.
		Table("teams").
		Join("INNER", "team_members", "team_members.team_id = teams.id").
		Where("team_members.user_id = ?", auth.GetID()).
		Where("teams.name LIKE ?", "%"+search+"%").
		Count(&models.Team{})

	return teams, len(teams), numberOfTotalItems, err
}

// Update updates a team's details
func (ts *TeamService) Update(s *xorm.Session, team *models.Team) (*models.Team, error) {
	// Check if we have a name
	if team.Name == "" {
		return nil, models.ErrTeamNameCannotBeEmpty{}
	}

	// Check if the team exists
	_, err := ts.GetByID(s, team.ID)
	if err != nil {
		return nil, err
	}

	_, err = s.ID(team.ID).UseBool("is_public").Update(team)
	if err != nil {
		return nil, err
	}

	// Get the newly updated team
	updatedTeam, err := ts.GetByID(s, team.ID)
	if err != nil {
		return nil, err
	}

	return updatedTeam, nil
}

// Delete deletes a team and all its associations
func (ts *TeamService) Delete(s *xorm.Session, teamID int64, doer web.Auth) error {
	// Get the team first for event dispatch
	team, err := ts.GetByID(s, teamID)
	if err != nil {
		return err
	}

	// Delete the team
	_, err = s.ID(teamID).Delete(&models.Team{})
	if err != nil {
		return err
	}

	// Delete team members
	_, err = s.Where("team_id = ?", teamID).Delete(&models.TeamMember{})
	if err != nil {
		return err
	}

	// Delete team <-> projects relations
	_, err = s.Where("team_id = ?", teamID).Delete(&models.TeamProject{})
	if err != nil {
		return err
	}

	return events.Dispatch(&models.TeamDeletedEvent{
		Team: team,
		Doer: doer,
	})
}

// AddDetailsToTeams adds more info (members, created_by) to teams
func (ts *TeamService) AddDetailsToTeams(s *xorm.Session, teams []*models.Team) error {
	if len(teams) == 0 {
		return nil
	}

	// Put the teams in a map to make assigning more info to it more efficient
	teamMap := make(map[int64]*models.Team, len(teams))
	var teamIDs []int64
	var ownerIDs []int64
	for _, team := range teams {
		teamMap[team.ID] = team
		teamIDs = append(teamIDs, team.ID)
		ownerIDs = append(ownerIDs, team.CreatedByID)
	}

	// Get all owners and team members
	users := make(map[int64]*models.TeamUser)
	err := s.
		Select("*").
		Table("users").
		Join("LEFT", "team_members", "team_members.user_id = users.id").
		Join("LEFT", "teams", "team_members.team_id = teams.id").
		Or(
			builder.In("team_id", teamIDs),
			builder.And(
				builder.In("users.id", ownerIDs),
				builder.Expr("teams.created_by_id = users.id"),
				builder.In("teams.id", teamIDs),
			),
		).
		Find(&users)
	if err != nil {
		return err
	}

	for _, u := range users {
		if _, exists := teamMap[u.TeamID]; !exists {
			continue
		}
		u.Email = ""
		teamMap[u.TeamID].Members = append(teamMap[u.TeamID].Members, u)
	}

	// We need to do this in a second loop as owners might not be the last ones in the project
	for _, team := range teamMap {
		if teamUser, has := users[team.CreatedByID]; has {
			team.CreatedBy = &teamUser.User
		}
	}

	return nil
}

// GetTeamsByIDs retrieves multiple teams by their IDs
func (ts *TeamService) GetTeamsByIDs(s *xorm.Session, teamIDs []int64) ([]*models.Team, error) {
	if len(teamIDs) == 0 {
		return []*models.Team{}, nil
	}

	teams := []*models.Team{}
	err := s.In("id", teamIDs).Find(&teams)
	if err != nil {
		return nil, err
	}

	err = ts.AddDetailsToTeams(s, teams)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

// CanRead checks if a user has read access to the team
func (ts *TeamService) CanRead(s *xorm.Session, teamID int64, auth web.Auth) (bool, int, error) {
	// Check if the user is in the team
	tm := &models.TeamMember{}
	can, err := s.
		Where("team_id = ?", teamID).
		And("user_id = ?", auth.GetID()).
		Get(tm)

	maxPermissions := 0
	if tm.Admin {
		maxPermissions = int(models.PermissionAdmin)
	}

	return can, maxPermissions, err
}

// CanUpdate checks if the user can update a team
func (ts *TeamService) CanUpdate(s *xorm.Session, teamID int64, auth web.Auth) (bool, error) {
	return ts.IsAdmin(s, teamID, auth)
}

// CanWrite is an alias for CanUpdate for consistency
func (ts *TeamService) CanWrite(s *xorm.Session, teamID int64, auth web.Auth) (bool, error) {
	return ts.CanUpdate(s, teamID, auth)
}

// CanDelete checks if a user can delete a team
func (ts *TeamService) CanDelete(s *xorm.Session, teamID int64, auth web.Auth) (bool, error) {
	return ts.IsAdmin(s, teamID, auth)
}

// CanCreate checks if the user can create a new team
func (ts *TeamService) CanCreate(_ *xorm.Session, a web.Auth) (bool, error) {
	if _, is := a.(*models.LinkSharing); is {
		return false, nil
	}

	// This is currently a dummy function, later on we could imagine global limits etc.
	return true, nil
}

// IsAdmin returns true when the user is admin of a team
func (ts *TeamService) IsAdmin(s *xorm.Session, teamID int64, auth web.Auth) (bool, error) {
	// Don't do anything if we're dealing with a link share auth here
	if _, is := auth.(*models.LinkSharing); is {
		return false, nil
	}

	// Check if the team exists to be able to return a proper error message if not
	_, err := ts.GetByID(s, teamID)
	if err != nil {
		return false, err
	}

	return s.Where("team_id = ?", teamID).
		And("user_id = ?", auth.GetID()).
		And("admin = ?", true).
		Get(&models.TeamMember{})
}

// HasPermission checks if a user has a specific permission level for a team
func (ts *TeamService) HasPermission(s *xorm.Session, teamID int64, u *user.User, permission models.Permission) (bool, error) {
	// Link shares cannot have team permissions
	if u == nil {
		return false, nil
	}

	switch permission {
	case models.PermissionRead:
		can, _, err := ts.CanRead(s, teamID, u)
		return can, err
	case models.PermissionWrite, models.PermissionAdmin:
		return ts.IsAdmin(s, teamID, u)
	default:
		return false, models.ErrInvalidPermission{Permission: permission}
	}
}

// AddMember adds a user to a team
func (ts *TeamService) AddMember(s *xorm.Session, teamID int64, username string, admin bool, doer *user.User) (*models.TeamMember, error) {
	// Check if the team exists
	team, err := ts.GetByID(s, teamID)
	if err != nil {
		return nil, err
	}

	// Check if the user exists
	member, err := user.GetUserByUsername(s, username)
	if err != nil {
		return nil, err
	}

	// Check if that user is already part of the team
	exists, err := s.
		Where("team_id = ? AND user_id = ?", teamID, member.ID).
		Get(&models.TeamMember{})
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, models.ErrUserIsMemberOfTeam{TeamID: teamID, UserID: member.ID}
	}

	tm := &models.TeamMember{
		TeamID:   teamID,
		Username: username,
		UserID:   member.ID,
		Admin:    admin,
	}

	_, err = s.Insert(tm)
	if err != nil {
		return nil, err
	}

	err = events.Dispatch(&models.TeamMemberAddedEvent{
		Team:   team,
		Member: member,
		Doer:   doer,
	})
	if err != nil {
		return nil, err
	}

	return tm, nil
}

// RemoveMember removes a user from a team
func (ts *TeamService) RemoveMember(s *xorm.Session, teamID int64, username string) error {
	// Check if the team exists
	team, err := ts.GetByID(s, teamID)
	if err != nil {
		return err
	}

	// Check if team is external (from OIDC/LDAP)
	if team.ExternalID != "" {
		return models.ErrCannotRemoveUserFromExternalTeam{TeamID: teamID}
	}

	// Find the numeric user id first
	member, err := user.GetUserByUsername(s, username)
	if err != nil {
		return err
	}

	// Check if this is the last member
	total, err2 := s.Where("team_id = ?", teamID).Count(&models.TeamMember{})
	if err2 != nil {
		return err2
	}
	if total == 1 {
		return models.ErrCannotDeleteLastTeamMember{TeamID: teamID, UserID: member.ID}
	}

	_, err = s.Where("team_id = ? AND user_id = ?", teamID, member.ID).Delete(&models.TeamMember{})
	return err
}

// UpdateMemberAdmin toggles a team member's admin status
func (ts *TeamService) UpdateMemberAdmin(s *xorm.Session, teamID int64, username string) (bool, error) {
	// Find the numeric user id
	member, err := user.GetUserByUsername(s, username)
	if err != nil {
		return false, err
	}

	// Get the full member object and change the admin permission
	tm := &models.TeamMember{}
	_, err = s.
		Where("team_id = ? AND user_id = ?", teamID, member.ID).
		Get(tm)
	if err != nil {
		return false, err
	}

	tm.Admin = !tm.Admin

	// Do the update
	_, err = s.
		Where("team_id = ? AND user_id = ?", teamID, member.ID).
		Cols("admin").
		Update(tm)

	return tm.Admin, err
}

// UpdateMemberPermission sets a team member's admin status to a specific value
func (ts *TeamService) UpdateMemberPermission(s *xorm.Session, teamID int64, userID int64, admin bool) error {
	// Update the admin status directly
	_, err := s.
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Cols("admin").
		Update(&models.TeamMember{Admin: admin})

	return err
}

// GetMembers retrieves all members of a team with pagination
func (ts *TeamService) GetMembers(s *xorm.Session, teamID int64, search string, page int, perPage int) (members []*models.TeamUser, resultCount int, totalItems int64, err error) {
	// Check if team exists
	_, err = ts.GetByIDSimple(s, teamID)
	if err != nil {
		return nil, 0, 0, err
	}

	limit, start := getLimitFromPageIndex(page, perPage)

	// Get team members with user details
	members = []*models.TeamUser{}
	query := s.
		Table("users").
		Join("INNER", "team_members", "team_members.user_id = users.id").
		Where("team_members.team_id = ?", teamID)

	if search != "" {
		query = query.Where("users.username LIKE ? OR users.name LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if limit > 0 {
		query = query.Limit(limit, start)
	}

	err = query.Find(&members)
	if err != nil {
		return nil, 0, 0, err
	}

	// Get total count
	totalItems, err = s.
		Table("users").
		Join("INNER", "team_members", "team_members.user_id = users.id").
		Where("team_members.team_id = ?", teamID).
		Count(&user.User{})
	if err != nil {
		return nil, 0, 0, err
	}

	return members, len(members), totalItems, nil
}

// IsMember checks if a user is a member of a team
func (ts *TeamService) IsMember(s *xorm.Session, teamID int64, userID int64) (bool, error) {
	return s.
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Exist(&models.TeamMember{})
}

// MembershipExists checks if a user is a member of a team
func (ts *TeamService) MembershipExists(s *xorm.Session, teamID int64, userID int64) (bool, error) {
	return s.
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Exist(&models.TeamMember{})
}

// CanCreateTeamMember checks if a user can add a new team member
func (ts *TeamService) CanCreateTeamMember(s *xorm.Session, teamID int64, a web.Auth) (bool, error) {
	return ts.IsAdmin(s, teamID, a)
}

// CanDeleteTeamMember checks if a user can delete a team member
func (ts *TeamService) CanDeleteTeamMember(s *xorm.Session, teamID int64, username string, a web.Auth) (bool, error) {
	// Get the user being removed
	u, err := user.GetUserByUsername(s, username)
	if err != nil {
		return false, err
	}

	// Users can remove themselves from a team
	if u.ID == a.GetID() {
		return true, nil
	}

	// Otherwise, must be team admin
	return ts.IsAdmin(s, teamID, a)
}

// CanUpdateTeamMember checks if a user can modify a team member's permissions
func (ts *TeamService) CanUpdateTeamMember(s *xorm.Session, teamID int64, a web.Auth) (bool, error) {
	return ts.IsAdmin(s, teamID, a)
}

// InitTeamService sets up dependency injection for team-related model functions.
// This function wires up the Team permission delegation functions to complete the
// permission migration started in T-PERM-012 and documented in T-PERM-016B.
func InitTeamService() {
	// Wire Team CRUD permission delegation functions - T-PERM-016B follow-up
	models.CheckTeamUpdateFunc = func(s *xorm.Session, teamID int64, a web.Auth) (bool, error) {
		return NewTeamService(s.Engine()).CanUpdate(s, teamID, a)
	}

	models.CheckTeamDeleteFunc = func(s *xorm.Session, teamID int64, a web.Auth) (bool, error) {
		return NewTeamService(s.Engine()).CanDelete(s, teamID, a)
	}
}
