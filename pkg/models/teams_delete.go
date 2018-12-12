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

import (
	"code.vikunja.io/api/pkg/metrics"
	_ "code.vikunja.io/web" // For swaggerdocs generation
)

// Delete deletes a team
// @Summary Deletes a team
// @Description Delets a team. This will also remove the access for all users in that team.
// @tags team
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Team ID"
// @Success 200 {object} models.Message "The team was successfully deleted."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid team object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{id} [delete]
func (t *Team) Delete() (err error) {

	// Check if the team exists
	_, err = GetTeamByID(t.ID)
	if err != nil {
		return
	}

	// Delete the team
	_, err = x.ID(t.ID).Delete(&Team{})
	if err != nil {
		return
	}

	// Delete team members
	_, err = x.Where("team_id = ?", t.ID).Delete(&TeamMember{})
	if err != nil {
		return
	}

	// Delete team <-> namespace relations
	_, err = x.Where("team_id = ?", t.ID).Delete(&TeamNamespace{})
	if err != nil {
		return
	}

	// Delete team <-> lists relations
	_, err = x.Where("team_id = ?", t.ID).Delete(&TeamList{})
	if err != nil {
		return
	}

	metrics.UpdateCount(-1, metrics.TeamCountKey)
	return
}
