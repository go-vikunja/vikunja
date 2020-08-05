// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	user2 "code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
)

// Create implements the create method to assign a user to a team
// @Summary Add a user to a team
// @Description Add a user to a team.
// @tags team
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Team ID"
// @Param team body models.TeamMember true "The user to be added to a team."
// @Success 200 {object} models.TeamMember "The newly created member object"
// @Failure 400 {object} web.HTTPError "Invalid member object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the team"
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{id}/members [put]
func (tm *TeamMember) Create(a web.Auth) (err error) {

	// Check if the team extst
	_, err = GetTeamByID(tm.TeamID)
	if err != nil {
		return
	}

	// Check if the user exists
	user, err := user2.GetUserByUsername(tm.Username)
	if err != nil {
		return
	}
	tm.UserID = user.ID

	// Check if that user is already part of the team
	exists, err := x.Where("team_id = ? AND user_id = ?", tm.TeamID, tm.UserID).
		Get(&TeamMember{})
	if err != nil {
		return
	}
	if exists {
		return ErrUserIsMemberOfTeam{tm.TeamID, tm.UserID}
	}

	// Insert the user
	_, err = x.Insert(tm)
	return
}

// Delete deletes a user from a team
// @Summary Remove a user from a team
// @Description Remove a user from a team. This will also revoke any access this user might have via that team.
// @tags team
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Team ID"
// @Param userID path int true "User ID"
// @Success 200 {object} models.Message "The user was successfully removed from the team."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{id}/members/{userID} [delete]
func (tm *TeamMember) Delete() (err error) {

	total, err := x.Where("team_id = ?", tm.TeamID).Count(&TeamMember{})
	if err != nil {
		return
	}
	if total == 1 {
		return ErrCannotDeleteLastTeamMember{tm.TeamID, tm.UserID}
	}

	// Find the numeric user id
	user, err := user2.GetUserByUsername(tm.Username)
	if err != nil {
		return
	}
	tm.UserID = user.ID

	_, err = x.Where("team_id = ? AND user_id = ?", tm.TeamID, tm.UserID).Delete(&TeamMember{})
	return
}

// Update toggles a team member's admin status
// @Summary Toggle a team member's admin status
// @Description If a user is team admin, this will make them member and vise-versa.
// @tags team
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Team ID"
// @Param userID path int true "User ID"
// @Success 200 {object} models.Message "The member right was successfully changed."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{id}/members/{userID}/admin [post]
func (tm *TeamMember) Update() (err error) {
	// Find the numeric user id
	user, err := user2.GetUserByUsername(tm.Username)
	if err != nil {
		return
	}
	tm.UserID = user.ID

	// Get the full member object and change the admin right
	ttm := &TeamMember{}
	_, err = x.
		Where("team_id = ? AND user_id = ?", tm.TeamID, tm.UserID).
		Get(ttm)
	if err != nil {
		return err
	}
	ttm.Admin = !ttm.Admin

	// Do the update
	_, err = x.
		Where("team_id = ? AND user_id = ?", tm.TeamID, tm.UserID).
		Cols("admin").
		Update(ttm)
	tm.Admin = ttm.Admin // Since we're returning the updated rights object
	return
}
