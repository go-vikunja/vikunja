package models

// Delete deletes a team <-> list relation based on the list & team id
func (tl *TeamList) Delete() (err error) {

	// Check if the team exists
	_, err = GetTeamByID(tl.TeamID)
	if err != nil {
		return
	}

	// Check if the team has access to the list
	has, err := x.Where("team_id = ? AND list_id = ?", tl.TeamID, tl.ListID).
		Get(&TeamList{})
	if err != nil {
		return
	}
	if !has {
		return ErrTeamDoesNotHaveAccessToList{TeamID: tl.TeamID, ListID: tl.ListID}
	}

	// Delete the relation
	_, err = x.Where("team_id = ?", tl.TeamID).
		And("list_id = ?", tl.ListID).
		Delete(TeamList{})

	return
}
