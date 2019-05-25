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

// Update updates a user <-> namespace relation
// @Summary Update a user <-> namespace relation
// @Description Update a user <-> namespace relation. Mostly used to update the right that user has.
// @tags sharing
// @Accept json
// @Produce json
// @Param namespaceID path int true "Namespace ID"
// @Param userID path int true "User ID"
// @Param namespace body models.NamespaceUser true "The user you want to update."
// @Security JWTKeyAuth
// @Success 200 {object} models.NamespaceUser "The updated user <-> namespace relation."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have admin-access to the namespace"
// @Failure 404 {object} code.vikunja.io/web.HTTPError "User or namespace does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{namespaceID}/users/{userID} [post]
func (nu *NamespaceUser) Update() (err error) {

	// Check if the right is valid
	if err := nu.Right.isValid(); err != nil {
		return err
	}

	// Check if the user exists
	user, err := GetUserByUsername(nu.Username)
	if err != nil {
		return err
	}
	nu.UserID = user.ID

	_, err = x.
		Where("namespace_id = ? AND user_id = ?", nu.NamespaceID, nu.UserID).
		Cols("right").
		Update(nu)
	return
}
