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

// Update updates a user <-> list relation
// @Summary Update a user <-> list relation
// @Description Update a user <-> list relation. Mostly used to update the right that user has.
// @tags sharing
// @Accept json
// @Produce json
// @Param listID path int true "List ID"
// @Param userID path int true "User ID"
// @Param list body models.ListUser true "The user you want to update."
// @Security JWTKeyAuth
// @Success 200 {object} models.ListUser "The updated user <-> list relation."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have admin-access to the list"
// @Failure 404 {object} code.vikunja.io/web.HTTPError "User or list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/users/{userID} [post]
func (lu *ListUser) Update() (err error) {

	// Check if the right is valid
	if err := lu.Right.isValid(); err != nil {
		return err
	}

	_, err = x.
		Where("list_id = ? AND user_id = ?", lu.ListID, lu.UserID).
		Cols("right").
		Update(lu)
	if err != nil {
		return err
	}

	err = updateListLastUpdated(&List{ID: lu.ListID})
	return
}
