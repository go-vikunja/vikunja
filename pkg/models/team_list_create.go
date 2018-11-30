//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import "code.vikunja.io/web"

// Create creates a new team <-> list relation
// @Summary Add a team to a list
// @Description Gives a team access to a list.
// @tags sharing
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "List ID"
// @Param list body models.TeamList true "The team you want to add to the list."
// @Success 200 {object} models.TeamList "The created team<->list relation."
// @Failure 400 {object} models.HTTPError "Invalid team list object provided."
// @Failure 404 {object} models.HTTPError "The team does not exist."
// @Failure 403 {object} models.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id}/teams [put]
func (tl *TeamList) Create(a web.Auth) (err error) {

	// Check if the rights are valid
	if err = tl.Right.isValid(); err != nil {
		return
	}

	// Check if the team exists
	_, err = GetTeamByID(tl.TeamID)
	if err != nil {
		return
	}

	// Check if the list exists
	l := &List{ID: tl.ListID}
	if err := l.GetSimpleByID(); err != nil {
		return err
	}

	// Check if the team is already on the list
	exists, err := x.Where("team_id = ?", tl.TeamID).
		And("list_id = ?", tl.ListID).
		Get(&TeamList{})
	if err != nil {
		return
	}
	if exists {
		return ErrTeamAlreadyHasAccess{tl.TeamID, tl.ListID}
	}

	// Insert the new team
	_, err = x.Insert(tl)
	return
}
