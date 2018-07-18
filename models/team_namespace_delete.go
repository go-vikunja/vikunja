package models

// Delete deletes a team <-> namespace relation based on the namespace & team id
func (tn *TeamNamespace) Delete() (err error) {

	// Check if the namespace exists
	_, err = GetNamespaceByID(tn.NamespaceID)
	if err != nil {
		return
	}

	// Check if the team exists
	_, err = GetTeamByID(tn.TeamID)
	if err != nil {
		return
	}

	// Delete the relation
	_, err = x.Where("team_id = ?", tn.TeamID).
		And("namespace_id = ?", tn.NamespaceID).
		Delete(tn)

	return
}
