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

// Delete deletes a list <-> user relation
// @Summary Delete a user from a list
// @Description Delets a user from a list. The user won't have access to the list anymore.
// @tags sharing
// @Produce json
// @Security JWTKeyAuth
// @Param listID path int true "List ID"
// @Param userID path int true "User ID"
// @Success 200 {object} models.Message "The user was successfully removed from the list."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
// @Failure 404 {object} code.vikunja.io/web.HTTPError "user or list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/users/{userID} [delete]
func (lu *ListUser) Delete() (err error) {

	// Check if the user exists
	user, err := GetUserByUsername(lu.Username)
	if err != nil {
		return
	}
	lu.UserID = user.ID

	// Check if the user has access to the list
	has, err := x.Where("user_id = ? AND list_id = ?", lu.UserID, lu.ListID).
		Get(&ListUser{})
	if err != nil {
		return
	}
	if !has {
		return ErrUserDoesNotHaveAccessToList{ListID: lu.ListID, UserID: lu.UserID}
	}

	_, err = x.Where("user_id = ? AND list_id = ?", lu.UserID, lu.ListID).
		Delete(&ListUser{})
	if err != nil {
		return err
	}

	err = updateListLastUpdated(&List{ID: lu.ListID})
	return
}
