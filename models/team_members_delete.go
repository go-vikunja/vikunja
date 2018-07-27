package models

// Delete deletes a user from a team
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
