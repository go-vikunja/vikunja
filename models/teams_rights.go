package models

// CanCreate checks if the user can create a new team
func (t *Team) CanCreate(user *User, id int64) bool {
	// This is currently a dummy function, later on we could imagine global limits etc.
	return true
}

// CanUpdate checks if the user can update a team
func (t *Team) CanUpdate(user *User, id int64) bool {

	// Check if the current user is in the team and has admin rights in it
	exists, _ := x.Where("team_id = ?", id).
		And("user_id = ?", user.ID).
		And("is_admin = ?", true).
		Get(&TeamMember{})

	return exists
}

// CanDelete
func (t *Team) CanDelete(user *User, id int64) bool {
	t.ID = id
	return t.IsAdmin(user)
}

// IsAdmin
func (t *Team) IsAdmin(user *User) bool {
	exists, _ := x.Where("team_id = ?", t.ID).
		And("user_id = ?", user.ID).
		And("is_admin = ?", true).
		Get(&TeamMember{})
	return exists
}

func (t *Team) CanRead(user *User) bool {
	// Check if the user is in the team
	exists, _ := x.Where("team_id = ?", t.ID).
		And("user_id = ?", user.ID).
		Get(&TeamMember{})
	return exists
}
