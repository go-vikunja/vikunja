package models

// Delete deletes a user from a team
func (tm *TeamMember) Delete() (err error) {
	_, err = x.Where("team_id = ? AND user_id = ?", tm.TeamID, tm.UserID).Delete(&TeamMember{})
	return
}
