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

// Create creates a new namespace <-> user relation
// @Summary Add a user to a namespace
// @Description Gives a user access to a namespace.
// @tags sharing
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Namespace ID"
// @Param namespace body models.NamespaceUser true "The user you want to add to the namespace."
// @Success 200 {object} models.NamespaceUser "The created user<->namespace relation."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid user namespace object provided."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "The user does not exist."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id}/users [put]
func (un *NamespaceUser) Create(a web.Auth) (err error) {
	// Reset the id
	un.ID = 0

	// Check if the right is valid
	if err := un.Right.isValid(); err != nil {
		return err
	}

	// Check if the namespace exists
	l, err := GetNamespaceByID(un.NamespaceID)
	if err != nil {
		return
	}

	// Check if the user exists
	if _, err = GetUserByID(un.UserID); err != nil {
		return err
	}

	// Check if the user already has access or is owner of that namespace
	// We explicitly DO NOT check for teams here
	if l.OwnerID == un.UserID {
		return ErrUserAlreadyHasNamespaceAccess{UserID: un.UserID, NamespaceID: un.NamespaceID}
	}

	exist, err := x.Where("namespace_id = ? AND user_id = ?", un.NamespaceID, un.UserID).Get(&NamespaceUser{})
	if err != nil {
		return
	}
	if exist {
		return ErrUserAlreadyHasNamespaceAccess{UserID: un.UserID, NamespaceID: un.NamespaceID}
	}

	// Insert user <-> namespace relation
	_, err = x.Insert(un)

	return
}
