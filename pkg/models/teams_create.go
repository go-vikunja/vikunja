package models

// Create is the handler to create a team
// @Summary Creates a new team
// @Description Creates a new team in a given namespace. The user needs write-access to the namespace.
// @tags team
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param team body models.Team true "The team you want to create."
// @Success 200 {object} models.Team "The created team."
// @Failure 400 {object} models.HTTPError "Invalid team object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams [put]
func (t *Team) Create(doer *User) (err error) {
	// Check if we have a name
	if t.Name == "" {
		return ErrTeamNameCannotBeEmpty{}
	}

	t.CreatedByID = doer.ID
	t.CreatedBy = *doer

	_, err = x.Insert(t)
	if err != nil {
		return
	}

	// Insert the current user as member and admin
	tm := TeamMember{TeamID: t.ID, UserID: doer.ID, Admin: true}
	err = tm.Create(doer)
	return
}
