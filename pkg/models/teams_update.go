package models

// Update is the handler to create a team
// @Summary Updates a team
// @Description Updates a team.
// @tags team
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Team ID"
// @Param team body models.Team true "The team with updated values you want to update."
// @Success 200 {object} models.Team "The updated team."
// @Failure 400 {object} models.HTTPError "Invalid team object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{id} [post]
func (t *Team) Update() (err error) {
	// Check if we have a name
	if t.Name == "" {
		return ErrTeamNameCannotBeEmpty{}
	}

	// Check if the team exists
	_, err = GetTeamByID(t.ID)
	if err != nil {
		return
	}

	_, err = x.ID(t.ID).Update(t)
	if err != nil {
		return
	}

	// Get the newly updated team
	*t, err = GetTeamByID(t.ID)

	return
}
