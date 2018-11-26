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

// Delete deletes a team <-> namespace relation based on the namespace & team id
// @Summary Delete a team from a namespace
// @Description Delets a team from a namespace. The team won't have access to the namespace anymore.
// @tags sharing
// @Produce json
// @Security ApiKeyAuth
// @Param namespaceID path int true "Namespace ID"
// @Param teamID path int true "team ID"
// @Success 200 {object} models.Message "The team was successfully deleted."
// @Failure 403 {object} models.HTTPError "The team does not have access to the namespace"
// @Failure 404 {object} models.HTTPError "team or namespace does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{namespaceID}/teams/{teamID} [delete]
func (tn *TeamNamespace) Delete() (err error) {

	// Check if the team exists
	_, err = GetTeamByID(tn.TeamID)
	if err != nil {
		return
	}

	// Check if the team has access to the namespace
	has, err := x.Where("team_id = ? AND namespace_id = ?", tn.TeamID, tn.NamespaceID).
		Get(&TeamNamespace{})
	if err != nil {
		return
	}
	if !has {
		return ErrTeamDoesNotHaveAccessToNamespace{TeamID: tn.TeamID, NamespaceID: tn.NamespaceID}
	}

	// Delete the relation
	_, err = x.Where("team_id = ?", tn.TeamID).
		And("namespace_id = ?", tn.NamespaceID).
		Delete(TeamNamespace{})

	return
}
