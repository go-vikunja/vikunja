package models

// Delete deletes a user from a team
// @Summary Remove a user from a team
// @Description Remove a user from a team. This will also revoke any access this user might have via that team.
// @tags team
// @Produce json
// @Security ApiKeyAuth
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

	_, err = x.Where("team_id = ? AND user_id = ?", tm.TeamID, tm.UserID).Delete(&TeamMember{})
	return
}
