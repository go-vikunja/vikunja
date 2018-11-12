package models

// Delete deletes a team <-> list relation based on the list & team id
// @Summary Delete a team from a list
// @Description Delets a team from a list. The team won't have access to the list anymore.
// @tags sharing
// @Produce json
// @Security ApiKeyAuth
// @Param listID path int true "List ID"
// @Param teamID path int true "Team ID"
// @Success 200 {object} models.Message "The team was successfully deleted."
// @Failure 403 {object} models.HTTPError "The user does not have access to the list"
// @Failure 404 {object} models.HTTPError "Team or list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/teams/{teamID} [delete]
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
