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

import _ "code.vikunja.io/web" // For swaggerdocs generation

// Delete deletes a team <-> list relation based on the list & team id
// @Summary Delete a team from a list
// @Description Delets a team from a list. The team won't have access to the list anymore.
// @tags sharing
// @Produce json
// @Security ApiKeyAuth
// @Param listID path int true "List ID"
// @Param teamID path int true "Team ID"
// @Success 200 {object} models.Message "The team was successfully deleted."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
// @Failure 404 {object} code.vikunja.io/web.HTTPError "Team or list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/teams/{teamID} [delete]
func (tl *TeamList) Delete() (err error) {

	// Check if the team exists
	_, err = GetTeamByID(tl.TeamID)
	if err != nil {
		return
	}

	// Check if the team has access to the list
	has, err := x.Where("team_id = ? AND list_id = ?", tl.TeamID, tl.ListID).
		Get(&TeamList{})
	if err != nil {
		return
	}
	if !has {
		return ErrTeamDoesNotHaveAccessToList{TeamID: tl.TeamID, ListID: tl.ListID}
	}

	// Delete the relation
	_, err = x.Where("team_id = ?", tl.TeamID).
		And("list_id = ?", tl.ListID).
		Delete(TeamList{})

	return
}
