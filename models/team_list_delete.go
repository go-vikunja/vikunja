package models

// Delete deletes a team <-> list relation based on the list & team id
func (tl *TeamList) Delete() (err error) {

	// Check if the list exists
	_, err = GetListByID(tl.ListID)
	if err != nil {
		return
	}

	// Check if the team exists
	_, err = GetTeamByID(tl.TeamID)
	if err != nil {
		return
	}

	// Delete the relation
	_, err = x.Where("team_id = ?", tl.TeamID).
		And("list_id = ?", tl.ListID).
		Delete(TeamList{})

	return
}
