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

// ReadAll implements the method to read all teams of a namespace
// @Summary Get teams on a namespace
// @Description Returns a namespace with all teams which have access on a given namespace.
// @tags sharing
// @Accept json
// @Produce json
// @Param id path int true "Namespace ID"
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param s query string false "Search teams by its name."
// @Security JWTKeyAuth
// @Success 200 {array} models.TeamWithRight "The teams with the right they have."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "No right to see the namespace."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id}/teams [get]
func (tn *TeamNamespace) ReadAll(search string, a web.Auth, page int) (interface{}, error) {
	user, err := getUserWithError(a)
	if err != nil {
		return nil, err
	}

	// Check if the user can read the namespace
	n := Namespace{ID: tn.NamespaceID}
	canRead, err := n.CanRead(user)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, ErrNeedToHaveNamespaceReadAccess{NamespaceID: tn.NamespaceID, UserID: user.ID}
	}

	// Get the teams
	all := []*TeamWithRight{}

	err = x.Table("teams").
		Join("INNER", "team_namespaces", "team_id = teams.id").
		Where("team_namespaces.namespace_id = ?", tn.NamespaceID).
		Limit(getLimitFromPageIndex(page)).
		Where("teams.name LIKE ?", "%"+search+"%").
		Find(&all)

	return all, err
}
