package models

// Update updates a team <-> list relation
// @Summary Update a team <-> list relation
// @Description Update a team <-> list relation. Mostly used to update the right that team has.
// @tags sharing
// @Accept json
// @Produce json
// @Param listID path int true "List ID"
// @Param teamID path int true "Team ID"
// @Param list body models.TeamList true "The team you want to update."
// @Security ApiKeyAuth
// @Success 200 {object} models.TeamList "The updated team <-> list relation."
// @Failure 403 {object} models.HTTPError "The user does not have admin-access to the list"
// @Failure 404 {object} models.HTTPError "Team or list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/teams/{teamID} [post]
func (tl *TeamList) Update() (err error) {

	// Check if the right is valid
	if err := tl.Right.isValid(); err != nil {
		return err
	}

	_, err = x.
		Where("list_id = ? AND team_id = ?", tl.ListID, tl.TeamID).
		Cols("right").
		Update(tl)
	return
}
