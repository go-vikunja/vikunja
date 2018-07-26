package models

// Create implements the create method to assign a user to a team
func (tm *TeamMember) Create(doer *User) (err error) {
	// Check if the team extst
	_, err = GetTeamByID(tm.TeamID)
	if err != nil {
		return
	}

	// Check if the user exists
	_, _, err = GetUserByID(tm.UserID)
	if err != nil {
		return
	}

	// Insert the user
	_, err = x.Insert(tm)
	return
}
