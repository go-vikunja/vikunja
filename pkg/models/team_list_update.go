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

// Update updates a team <-> list relation
// @Summary Update a team <-> list relation
// @Description Update a team <-> list relation. Mostly used to update the right that team has.
// @tags sharing
// @Accept json
// @Produce json
// @Param listID path int true "List ID"
// @Param teamID path int true "Team ID"
// @Param list body models.TeamList true "The team you want to update."
// @Security ApiKeyAuth
// @Success 200 {object} models.TeamList "The updated team <-> list relation."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have admin-access to the list"
// @Failure 404 {object} code.vikunja.io/web.HTTPError "Team or list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/teams/{teamID} [post]
func (tl *TeamList) Update() (err error) {

	// Check if the right is valid
	if err := tl.Right.isValid(); err != nil {
		return err
	}

	_, err = x.
		Where("list_id = ? AND team_id = ?", tl.ListID, tl.TeamID).
		Cols("right").
		Update(tl)
	return
}
