package models

// Delete deletes a team
// @Summary Deletes a team
// @Description Delets a team. This will also remove the access for all users in that team.
// @tags team
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Team ID"
// @Success 200 {object} models.Message "The team was successfully deleted."
// @Failure 400 {object} models.HTTPError "Invalid team object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{id} [delete]
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
