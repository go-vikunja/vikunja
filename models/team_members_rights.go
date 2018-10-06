package models

// CanCreate checks if the user can add a new tem member
func (tm *TeamMember) CanCreate(user *User) bool {
	return tm.IsAdmin(user)
}

// CanDelete checks if the user can delete a new team member
func (tm *TeamMember) CanDelete(user *User) bool {
	return tm.IsAdmin(user)
}

// IsAdmin checks if the user is team admin
func (tm *TeamMember) IsAdmin(user *User) bool {
	// A user can add a member to a team if he is admin of that team
	exists, err := x.Where("user_id = ? AND team_id = ? AND admin = ?", user.ID, tm.TeamID, true).
		Get(&TeamMember{})
	if err != nil {
		Log.Error("Error occurred during IsAdmin for TeamMember: %s", err)
		return false
	}
	return exists
}
