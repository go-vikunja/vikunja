package models

// Delete deletes a team
func (t *Team) Delete() (err error) {

	// Check if the team exists
	_, err = GetTeamByID(t.ID)
	if err != nil {
		return
	}

	// Delete the team
	_, err = x.ID(t.ID).Delete(&Team{})
	if err != nil {
		return
	}

	// Delete team members
	_, err = x.Where("team_id = ?", t.ID).Delete(&TeamMember{})
	if err != nil {
		return
	}

	// Delete team <-> namespace relations
	_, err = x.Where("team_id = ?", t.ID).Delete(&TeamNamespace{})
	if err != nil {
		return
	}

	// Delete team <-> lists relations
	_, err = x.Where("team_id = ?", t.ID).Delete(&TeamList{})
	return
}
