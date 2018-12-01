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

// Delete deletes a namespace <-> user relation
// @Summary Delete a user from a namespace
// @Description Delets a user from a namespace. The user won't have access to the namespace anymore.
// @tags sharing
// @Produce json
// @Security ApiKeyAuth
// @Param namespaceID path int true "Namespace ID"
// @Param userID path int true "user ID"
// @Success 200 {object} models.Message "The user was successfully deleted."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the namespace"
// @Failure 404 {object} code.vikunja.io/web.HTTPError "user or namespace does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{namespaceID}/users/{userID} [delete]
func (nu *NamespaceUser) Delete() (err error) {

	// Check if the user exists
	_, err = GetUserByID(nu.UserID)
	if err != nil {
		return
	}

	// Check if the user has access to the namespace
	has, err := x.Where("user_id = ? AND namespace_id = ?", nu.UserID, nu.NamespaceID).
		Get(&NamespaceUser{})
	if err != nil {
		return
	}
	if !has {
		return ErrUserDoesNotHaveAccessToNamespace{NamespaceID: nu.NamespaceID, UserID: nu.UserID}
	}

	_, err = x.Where("user_id = ? AND namespace_id = ?", nu.UserID, nu.NamespaceID).
		Delete(&NamespaceUser{})
	return
}
