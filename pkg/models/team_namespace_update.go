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

// Update updates a team <-> namespace relation
// @Summary Update a team <-> namespace relation
// @Description Update a team <-> namespace relation. Mostly used to update the right that team has.
// @tags sharing
// @Accept json
// @Produce json
// @Param namespaceID path int true "Namespace ID"
// @Param teamID path int true "Team ID"
// @Param namespace body models.TeamNamespace true "The team you want to update."
// @Security JWTKeyAuth
// @Success 200 {object} models.TeamNamespace "The updated team <-> namespace relation."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The team does not have admin-access to the namespace"
// @Failure 404 {object} code.vikunja.io/web.HTTPError "Team or namespace does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{namespaceID}/teams/{teamID} [post]
func (tn *TeamNamespace) Update() (err error) {

	// Check if the right is valid
	if err := tn.Right.isValid(); err != nil {
		return err
	}

	_, err = x.
		Where("namespace_id = ? AND team_id = ?", tn.TeamID, tn.TeamID).
		Cols("right").
		Update(tn)
	return
}
