package models

// Delete deletes a team
func (t *Team) Delete(id int64) (err error) {

	// Check if the team exists
	_, err = GetTeamByID(id)
	if err != nil {
		return
	}

	// Delete the team
	_, err = x.ID(id).Delete(&Team{})
	if err != nil {
		return
	}

	// Delete team members
	_, err = x.Where("team_id = ?", id).Delete(&TeamMember{})
	if err != nil {
		return
	}

	// Delete team <-> namespace relations
	_, err = x.Where("team_id = ?", id).Delete(&TeamNamespace{})
	if err != nil {
		return
	}

	// Delete team <-> lists relations
	_, err = x.Where("team_id = ?", id).Delete(&TeamList{})
	return
}
