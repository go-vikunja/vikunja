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

// Create creates a new team <-> namespace relation
// @Summary Add a team to a namespace
// @Description Gives a team access to a namespace.
// @tags sharing
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Namespace ID"
// @Param namespace body models.TeamNamespace true "The team you want to add to the namespace."
// @Success 200 {object} models.TeamNamespace "The created team<->namespace relation."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid team namespace object provided."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "The team does not exist."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The team does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id}/teams [put]
func (tn *TeamNamespace) Create(a web.Auth) (err error) {

	// Check if the rights are valid
	if err = tn.Right.isValid(); err != nil {
		return
	}

	// Check if the team exists
	_, err = GetTeamByID(tn.TeamID)
	if err != nil {
		return
	}

	// Check if the namespace exists
	_, err = GetNamespaceByID(tn.NamespaceID)
	if err != nil {
		return
	}

	// Check if the team already has access to the namespace
	exists, err := x.Where("team_id = ?", tn.TeamID).
		And("namespace_id = ?", tn.NamespaceID).
		Get(&TeamNamespace{})
	if err != nil {
		return
	}
	if exists {
		return ErrTeamAlreadyHasAccess{tn.TeamID, tn.NamespaceID}
	}

	// Insert the new team
	_, err = x.Insert(tn)
	return
}
