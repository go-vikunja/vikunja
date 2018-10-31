package models

// Delete deletes a team <-> namespace relation based on the namespace & team id
func (tn *TeamNamespace) Delete() (err error) {

	// Check if the team exists
	_, err = GetTeamByID(tn.TeamID)
	if err != nil {
		return
	}

	// Check if the team has access to the namespace
	has, err := x.Where("team_id = ? AND namespace_id = ?", tn.TeamID, tn.NamespaceID).
		Get(&TeamNamespace{})
	if err != nil {
		return
	}
	if !has {
		return ErrTeamDoesNotHaveAccessToNamespace{TeamID: tn.TeamID, NamespaceID: tn.NamespaceID}
	}

	// Delete the relation
	_, err = x.Where("team_id = ?", tn.TeamID).
		And("namespace_id = ?", tn.NamespaceID).
		Delete(TeamNamespace{})

	return
}
