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

	// Check if the team is already on the list
	exists, err := x.Where("team_id = ?", tl.TeamID).
		And("list_id = ?", tl.ListID).
		Get(&TeamList{})
	if err != nil {
		return
	}
	if exists {
		return ErrTeamAlreadyHasAccess{tl.TeamID, tl.ListID}
	}

	// Insert the new team
	_, err = x.Insert(tl)
	return
}
