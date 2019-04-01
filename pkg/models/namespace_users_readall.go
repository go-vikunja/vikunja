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

// ReadAll gets all users who have access to a namespace
// @Summary Get users on a namespace
// @Description Returns a namespace with all users which have access on a given namespace.
// @tags sharing
// @Accept json
// @Produce json
// @Param id path int true "Namespace ID"
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param s query string false "Search users by its name."
// @Security JWTKeyAuth
// @Success 200 {array} models.UserWithRight "The users with the right they have."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "No right to see the namespace."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id}/users [get]
func (nu *NamespaceUser) ReadAll(search string, a web.Auth, page int) (interface{}, error) {
	u, err := getUserWithError(a)
	if err != nil {
		return nil, err
	}

	// Check if the user has access to the namespace
	l := Namespace{ID: nu.NamespaceID}
	canRead, err := l.CanRead(u)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, ErrNeedToHaveNamespaceReadAccess{}
	}

	// Get all users
	all := []*UserWithRight{}
	err = x.
		Join("INNER", "users_namespace", "user_id = users.id").
		Where("users_namespace.namespace_id = ?", nu.NamespaceID).
		Limit(getLimitFromPageIndex(page)).
		Where("users.username LIKE ?", "%"+search+"%").
		Find(&all)

	return all, err
}
