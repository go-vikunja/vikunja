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

// Create creates a new list <-> user relation
// @Summary Add a user to a list
// @Description Gives a user access to a list.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "List ID"
// @Param list body models.ListUser true "The user you want to add to the list."
// @Success 200 {object} models.ListUser "The created user<->list relation."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid user list object provided."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "The user does not exist."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id}/users [put]
func (lu *ListUser) Create(a web.Auth) (err error) {

	// Check if the right is valid
	if err := lu.Right.isValid(); err != nil {
		return err
	}

	// Check if the list exists
	l := &List{ID: lu.ListID}
	if err = l.GetSimpleByID(); err != nil {
		return
	}

	// Check if the user exists
	if _, err = GetUserByID(lu.UserID); err != nil {
		return err
	}

	// Check if the user already has access or is owner of that list
	// We explicitly DONT check for teams here
	if l.OwnerID == lu.UserID {
		return ErrUserAlreadyHasAccess{UserID: lu.UserID, ListID: lu.ListID}
	}

	exist, err := x.Where("list_id = ? AND user_id = ?", lu.ListID, lu.UserID).Get(&ListUser{})
	if err != nil {
		return
	}
	if exist {
		return ErrUserAlreadyHasAccess{UserID: lu.UserID, ListID: lu.ListID}
	}

	// Insert user <-> list relation
	_, err = x.Insert(lu)
	if err != nil {
		return err
	}

	err = updateListLastUpdated(l)
	return
}
