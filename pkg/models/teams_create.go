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

// Create is the handler to create a team
// @Summary Creates a new team
// @Description Creates a new team in a given namespace. The user needs write-access to the namespace.
// @tags team
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param team body models.Team true "The team you want to create."
// @Success 200 {object} models.Team "The created team."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid team object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams [put]
func (t *Team) Create(a web.Auth) (err error) {
	doer, err := getUserWithError(a)
	if err != nil {
		return err
	}

	// Check if we have a name
	if t.Name == "" {
		return ErrTeamNameCannotBeEmpty{}
	}

	t.CreatedByID = doer.ID
	t.CreatedBy = *doer

	_, err = x.Insert(t)
	if err != nil {
		return
	}

	// Insert the current user as member and admin
	tm := TeamMember{TeamID: t.ID, UserID: doer.ID, Admin: true}
	err = tm.Create(doer)
	return
}
