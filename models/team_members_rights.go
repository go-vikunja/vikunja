package models

func (tm *TeamMember) CanCreate(user *User) bool {

	// A user can add a member to a team if he is admin of that team
	exists, _ := x.Where("user_id = ? AND team_id = ? AND admin = ?", user.ID, tm.TeamID, true).
		Get(&TeamMember{})
	return exists
}