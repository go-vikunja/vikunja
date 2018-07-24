package models

// Create creates a new team <-> list relation
func (tl *TeamList) Create(doer *User) (err error) {

	// Check if the rights are valid
	if err = tl.Right.isValid(); err != nil {
		return
	}

	// Check if the team exists
	_, err = GetTeamByID(tl.TeamID)
	if err != nil {
		return
	}

	// Check if the list exists
	_, err = GetListByID(tl.ListID)
	if err != nil {
		return
	}

	// Insert the new team
	_, err = x.Insert(tl)
	return
}
