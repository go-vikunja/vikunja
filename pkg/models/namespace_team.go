// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/timeutil"
	"code.vikunja.io/web"
)

// TeamNamespace defines the relationship between a Team and a Namespace
type TeamNamespace struct {
	// The unique, numeric id of this namespace <-> team relation.
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"id"`
	// The team id.
	TeamID int64 `xorm:"int(11) not null INDEX" json:"team_id" param:"team"`
	// The namespace id.
	NamespaceID int64 `xorm:"int(11) not null INDEX" json:"-" param:"namespace"`
	// The right this team has. 0 = Read only, 1 = Read & Write, 2 = Admin. See the docs for more details.
	Right Right `xorm:"int(11) INDEX not null default 0" json:"right" valid:"length(0|2)" maximum:"2" default:"0"`

	// A timestamp when this relation was created. You cannot change this value.
	Created timeutil.TimeStamp `xorm:"created not null" json:"created"`
	// A timestamp when this relation was last updated. You cannot change this value.
	Updated timeutil.TimeStamp `xorm:"updated not null" json:"updated"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName makes beautiful table names
func (TeamNamespace) TableName() string {
	return "team_namespaces"
}

// Create creates a new team <-> namespace relation
// @Summary Add a team to a namespace
// @Description Gives a team access to a namespace.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Namespace ID"
// @Param namespace body models.TeamNamespace true "The team you want to add to the namespace."
// @Success 200 {object} models.TeamNamespace "The created team<->namespace relation."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid team namespace object provided."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "The team does not exist."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The team does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id}/teams [put]
func (tn *TeamNamespace) Create(a web.Auth) (err error) {

	// Check if the rights are valid
	if err = tn.Right.isValid(); err != nil {
		return
	}

	// Check if the team exists
	_, err = GetTeamByID(tn.TeamID)
	if err != nil {
		return
	}

	// Check if the namespace exists
	_, err = GetNamespaceByID(tn.NamespaceID)
	if err != nil {
		return
	}

	// Check if the team already has access to the namespace
	exists, err := x.Where("team_id = ?", tn.TeamID).
		And("namespace_id = ?", tn.NamespaceID).
		Get(&TeamNamespace{})
	if err != nil {
		return
	}
	if exists {
		return ErrTeamAlreadyHasAccess{tn.TeamID, tn.NamespaceID}
	}

	// Insert the new team
	_, err = x.Insert(tn)
	return
}

// Delete deletes a team <-> namespace relation based on the namespace & team id
// @Summary Delete a team from a namespace
// @Description Delets a team from a namespace. The team won't have access to the namespace anymore.
// @tags sharing
// @Produce json
// @Security JWTKeyAuth
// @Param namespaceID path int true "Namespace ID"
// @Param teamID path int true "team ID"
// @Success 200 {object} models.Message "The team was successfully deleted."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The team does not have access to the namespace"
// @Failure 404 {object} code.vikunja.io/web.HTTPError "team or namespace does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{namespaceID}/teams/{teamID} [delete]
func (tn *TeamNamespace) Delete() (err error) {

	// Check if the team exists
	_, err = GetTeamByID(tn.TeamID)
	if err != nil {
		return
	}

	// Check if the team has access to the namespace
	has, err := x.Where("team_id = ? AND namespace_id = ?", tn.TeamID, tn.NamespaceID).
		Get(&TeamNamespace{})
	if err != nil {
		return
	}
	if !has {
		return ErrTeamDoesNotHaveAccessToNamespace{TeamID: tn.TeamID, NamespaceID: tn.NamespaceID}
	}

	// Delete the relation
	_, err = x.Where("team_id = ?", tn.TeamID).
		And("namespace_id = ?", tn.NamespaceID).
		Delete(TeamNamespace{})

	return
}

// ReadAll implements the method to read all teams of a namespace
// @Summary Get teams on a namespace
// @Description Returns a namespace with all teams which have access on a given namespace.
// @tags sharing
// @Accept json
// @Produce json
// @Param id path int true "Namespace ID"
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search teams by its name."
// @Security JWTKeyAuth
// @Success 200 {array} models.TeamWithRight "The teams with the right they have."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "No right to see the namespace."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id}/teams [get]
func (tn *TeamNamespace) ReadAll(a web.Auth, search string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	// Check if the user can read the namespace
	n := Namespace{ID: tn.NamespaceID}
	canRead, err := n.CanRead(a)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		return nil, 0, 0, ErrNeedToHaveNamespaceReadAccess{NamespaceID: tn.NamespaceID, UserID: a.GetID()}
	}

	// Get the teams
	all := []*TeamWithRight{}

	limit, start := getLimitFromPageIndex(page, perPage)

	query := x.Table("teams").
		Join("INNER", "team_namespaces", "team_id = teams.id").
		Where("team_namespaces.namespace_id = ?", tn.NamespaceID).
		Where("teams.name LIKE ?", "%"+search+"%")
	if limit > 0 {
		query = query.Limit(limit, start)
	}
	err = query.Find(&all)
	if err != nil {
		return nil, 0, 0, err
	}

	teams := []*Team{}
	for _, t := range all {
		teams = append(teams, &t.Team)
	}

	err = addMoreInfoToTeams(teams)
	if err != nil {
		return
	}

	numberOfTotalItems, err = x.Table("teams").
		Join("INNER", "team_namespaces", "team_id = teams.id").
		Where("team_namespaces.namespace_id = ?", tn.NamespaceID).
		Where("teams.name LIKE ?", "%"+search+"%").
		Count(&TeamWithRight{})

	return all, len(all), numberOfTotalItems, err
}

// Update updates a team <-> namespace relation
// @Summary Update a team <-> namespace relation
// @Description Update a team <-> namespace relation. Mostly used to update the right that team has.
// @tags sharing
// @Accept json
// @Produce json
// @Param namespaceID path int true "Namespace ID"
// @Param teamID path int true "Team ID"
// @Param namespace body models.TeamNamespace true "The team you want to update."
// @Security JWTKeyAuth
// @Success 200 {object} models.TeamNamespace "The updated team <-> namespace relation."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The team does not have admin-access to the namespace"
// @Failure 404 {object} code.vikunja.io/web.HTTPError "Team or namespace does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{namespaceID}/teams/{teamID} [post]
func (tn *TeamNamespace) Update() (err error) {

	// Check if the right is valid
	if err := tn.Right.isValid(); err != nil {
		return err
	}

	_, err = x.
		Where("namespace_id = ? AND team_id = ?", tn.TeamID, tn.TeamID).
		Cols("right").
		Update(tn)
	return
}
