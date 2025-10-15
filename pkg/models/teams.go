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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// TeamServiceProvider defines the interface for team service operations
// This allows the model layer to delegate to the service layer without import cycles
type TeamServiceProvider interface {
	Create(s *xorm.Session, team *Team, doer *user.User, firstUserShouldBeAdmin bool) (*Team, error)
	GetByID(s *xorm.Session, teamID int64) (*Team, error)
	GetAll(s *xorm.Session, auth web.Auth, search string, page int, perPage int, includePublic bool) (teams []*Team, resultCount int, totalItems int64, err error)
	Update(s *xorm.Session, team *Team) (*Team, error)
	Delete(s *xorm.Session, teamID int64, doer web.Auth) error
	AddDetailsToTeams(s *xorm.Session, teams []*Team) error
	// Permission methods
	CanRead(s *xorm.Session, teamID int64, auth web.Auth) (bool, int, error)
	CanCreate(s *xorm.Session, a web.Auth) (bool, error)
	CanUpdate(s *xorm.Session, teamID int64, auth web.Auth) (bool, error)
	CanDelete(s *xorm.Session, teamID int64, auth web.Auth) (bool, error)
	IsAdmin(s *xorm.Session, teamID int64, auth web.Auth) (bool, error)
	// Team member permission methods
	CanCreateTeamMember(s *xorm.Session, teamID int64, a web.Auth) (bool, error)
	CanDeleteTeamMember(s *xorm.Session, teamID int64, username string, a web.Auth) (bool, error)
	CanUpdateTeamMember(s *xorm.Session, teamID int64, a web.Auth) (bool, error)
}

var teamService TeamServiceProvider

// RegisterTeamService registers the team service for use by the model layer
func RegisterTeamService(service TeamServiceProvider) {
	teamService = service
}

// getTeamService returns the registered team service
func getTeamService() TeamServiceProvider {
	if teamService == nil {
		panic("TeamService not registered. Call RegisterTeamService during initialization.")
	}
	return teamService
}

// Team holds a team object
type Team struct {
	// The unique, numeric id of this team.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id" param:"team"`
	// The name of this team.
	Name string `xorm:"varchar(250) not null" json:"name" valid:"required,runelength(1|250)" minLength:"1" maxLength:"250"`
	// The team's description.
	Description string `xorm:"longtext null" json:"description"`
	CreatedByID int64  `xorm:"bigint not null INDEX" json:"-"`
	// The team's external id provided by the openid or ldap provider
	ExternalID string `xorm:"varchar(250) null" maxLength:"250" json:"external_id"`
	// Contains the issuer extracted from the vikunja_groups claim if this team was created through oidc
	Issuer string `xorm:"text null" json:"-"`

	// The user who created this team.
	CreatedBy *user.User `xorm:"-" json:"created_by"`
	// An array of all members in this team.
	Members []*TeamUser `xorm:"-" json:"members"`

	// A timestamp when this relation was created. You cannot change this value.
	Created time.Time `xorm:"created" json:"created"`
	// A timestamp when this relation was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated" json:"updated"`

	// Defines wether the team should be publicly discoverable when sharing a project
	IsPublic bool `xorm:"not null default false" json:"is_public"`

	// Query parameter controlling whether to include public projects or not
	IncludePublic bool `xorm:"-" query:"include_public" json:"include_public"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName makes beautiful table names
func (*Team) TableName() string {
	return "teams"
}

// TeamMember defines the relationship between a user and a team
type TeamMember struct {
	// The unique, numeric id of this team member relation.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id"`
	// The team id.
	TeamID int64 `xorm:"bigint not null INDEX" json:"-" param:"team"`
	// The username of the member. We use this to prevent automated user id entering.
	Username string `xorm:"-" json:"username" param:"user"`
	// Used under the hood to manage team members
	UserID int64 `xorm:"bigint not null INDEX" json:"-"`
	// Whether or not the member is an admin of the team. See the docs for more about what a team admin can do
	Admin bool `xorm:"null" json:"admin"`

	// A timestamp when this relation was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName makes beautiful table names
func (*TeamMember) TableName() string {
	return "team_members"
}

// TeamUser is the team member type
type TeamUser struct {
	user.User `xorm:"extends"`
	// Whether the member is an admin of the team. See the docs for more about what a team admin can do
	Admin  bool  `json:"admin"`
	TeamID int64 `json:"-"`
}

// AddMoreInfoToTeams adds more info (members, created_by) to teams
// @Deprecated: Use services.TeamService.AddDetailsToTeams instead
func AddMoreInfoToTeams(s *xorm.Session, teams []*Team) (err error) {
	if teamService != nil {
		return teamService.AddDetailsToTeams(s, teams)
	}

	// Fallback implementation for tests or migration code
	if len(teams) == 0 {
		return nil
	}

	// Put the teams in a map to make assigning more info to it more efficient
	teamMap := make(map[int64]*Team, len(teams))
	var teamIDs []int64
	var ownerIDs []int64
	for _, team := range teams {
		teamMap[team.ID] = team
		teamIDs = append(teamIDs, team.ID)
		ownerIDs = append(ownerIDs, team.CreatedByID)
	}

	// Get all owners and team members
	users := make(map[int64]*TeamUser)
	err = s.
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
		return
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
	return
}

// CreateNewTeam creates a new team and assignes the user that has caused creation
// as the first member of the team
// If firstUserShouldBeAdmin is true, the user will be an admin of the team
// Note: this function has been extracted from the Create method to allow
// an additional parameter to control whether the user should become admin of the team
func (t *Team) CreateNewTeam(s *xorm.Session, a web.Auth, firstUserShouldBeAdmin bool) (err error) {

	doer, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	// Check if we have a name
	if t.Name == "" {
		return ErrTeamNameCannotBeEmpty{}
	}

	t.ID = 0
	t.CreatedByID = doer.ID
	t.CreatedBy = doer

	_, err = s.Insert(t)
	if err != nil {
		return
	}

	tm := TeamMember{TeamID: t.ID, Username: doer.Username, Admin: firstUserShouldBeAdmin}
	if err = tm.Create(s, doer); err != nil {
		return err
	}

	return events.Dispatch(&TeamCreatedEvent{
		Team: t,
		Doer: a,
	})
}

// ReadOne implements the CRUD method to get one team
// @Summary Gets one team
// @Description Returns a team by its ID.
// @tags team
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Team ID"
// @Success 200 {object} models.Team "The team"
// @Failure 403 {object} web.HTTPError "The user does not have access to the team"
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{id} [get]
// @Deprecated: Use services.TeamService.GetByID instead
func (t *Team) ReadOne(s *xorm.Session, _ web.Auth) (err error) {
	if teamService == nil {
		return ErrTeamDoesNotExist{t.ID}
	}

	team, err := teamService.GetByID(s, t.ID)
	if err != nil {
		return err
	}
	*t = *team
	return nil
}

// ReadAll gets all teams the user is part of
// @Summary Get teams
// @Description Returns all teams the current user is part of.
// @tags team
// @Accept json
// @Produce json
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search teams by its name."
// @Security JWTKeyAuth
// @Success 200 {array} models.Team "The teams."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams [get]
// @Deprecated: Use services.TeamService.GetAll instead
func (t *Team) ReadAll(s *xorm.Session, a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	if teamService != nil {
		teams, count, total, err := teamService.GetAll(s, a, search, page, perPage, t.IncludePublic)
		return teams, count, total, err
	}

	// Fallback implementation
	if _, is := a.(*LinkSharing); is {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	limit, start := getLimitFromPageIndex(page, perPage)
	all := []*Team{}

	query := s.Distinct("teams.*").
		Table("teams").
		Join("INNER", "team_members", "team_members.team_id = teams.id").
		Where(db.ILIKE("teams.name", search))

	// If public teams are enabled, we want to include them in the result
	if config.ServiceEnablePublicTeams.GetBool() && t.IncludePublic {
		query = query.Where(
			builder.Or(
				builder.Eq{"teams.is_public": true},
				builder.Eq{"team_members.user_id": a.GetID()},
			),
		)
	} else {
		query = query.Where("team_members.user_id = ?", a.GetID())
	}

	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&all)
	if err != nil {
		return nil, 0, 0, err
	}

	err = AddMoreInfoToTeams(s, all)
	if err != nil {
		return nil, 0, 0, err
	}

	numberOfTotalItems, err = s.
		Table("teams").
		Join("INNER", "team_members", "team_members.team_id = teams.id").
		Where("team_members.user_id = ?", a.GetID()).
		Where("teams.name LIKE ?", "%"+search+"%").
		Count(&Team{})
	return all, len(all), numberOfTotalItems, err
}

// Create is the handler to create a team
// @Summary Creates a new team
// @Description Creates a new team.
// @tags team
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param team body models.Team true "The team you want to create."
// @Success 201 {object} models.Team "The created team."
// @Failure 400 {object} web.HTTPError "Invalid team object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams [put]
// @Deprecated: Use services.TeamService.Create instead
func (t *Team) Create(s *xorm.Session, a web.Auth) (err error) {
	doer, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	if teamService != nil {
		created, err := teamService.Create(s, t, doer, true)
		if err != nil {
			return err
		}
		*t = *created
		return nil
	}

	// Fallback implementation
	err = t.CreateNewTeam(s, a, true)
	if err != nil {
		return err
	}

	return
}

// Delete deletes a team
// @Summary Deletes a team
// @Description Delets a team. This will also remove the access for all users in that team.
// @tags team
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Team ID"
// @Success 200 {object} models.Message "The team was successfully deleted."
// @Failure 400 {object} web.HTTPError "Invalid team object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{id} [delete]
// @Deprecated: Use services.TeamService.Delete instead
func (t *Team) Delete(s *xorm.Session, a web.Auth) (err error) {
	if teamService != nil {
		return teamService.Delete(s, t.ID, a)
	}

	// Fallback implementation
	// Delete the team
	_, err = s.ID(t.ID).Delete(&Team{})
	if err != nil {
		return
	}

	// Delete team members
	_, err = s.Where("team_id = ?", t.ID).Delete(&TeamMember{})
	if err != nil {
		return
	}

	// Delete team <-> projects relations
	_, err = s.Where("team_id = ?", t.ID).Delete(&TeamProject{})
	if err != nil {
		return
	}

	return events.Dispatch(&TeamDeletedEvent{
		Team: t,
		Doer: a,
	})
}

// Update is the handler to create a team
// @Summary Updates a team
// @Description Updates a team.
// @tags team
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Team ID"
// @Param team body models.Team true "The team with updated values you want to update."
// @Success 200 {object} models.Team "The updated team."
// @Failure 400 {object} web.HTTPError "Invalid team object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{id} [post]
// @Deprecated: Use services.TeamService.Update instead
func (t *Team) Update(s *xorm.Session, _ web.Auth) (err error) {
	if teamService != nil {
		updated, err := teamService.Update(s, t)
		if err != nil {
			return err
		}
		*t = *updated
		return nil
	}

	// Fallback implementation
	return ErrTeamDoesNotExist{t.ID}
}

// ===== Permission Methods =====
// These methods delegate to the service layer via function pointers

// CanRead checks if the user can read a team
func (t *Team) CanRead(s *xorm.Session, a web.Auth) (canRead bool, maxPermission int, err error) {
	if CheckTeamReadFunc == nil {
		return false, 0, ErrPermissionDelegationNotInitialized{}
	}
	return CheckTeamReadFunc(s, t.ID, a)
}

// CanCreate checks if the user can create a team
func (t *Team) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckTeamCreateFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	return CheckTeamCreateFunc(s, t, a)
}

// CanUpdate checks if the user can update a team
func (t *Team) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckTeamUpdateFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	return CheckTeamUpdateFunc(s, t.ID, a)
}

// CanDelete checks if the user can delete a team
func (t *Team) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckTeamDeleteFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	return CheckTeamDeleteFunc(s, t.ID, a)
}

// IsAdmin checks if the user is an admin of the team
func (t *Team) IsAdmin(s *xorm.Session, a web.Auth) (isAdmin bool, err error) {
	ts := getTeamService()
	return ts.IsAdmin(s, t.ID, a)
}
