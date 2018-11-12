package models

// Create implements the create method to assign a user to a team
// @Summary Add a user to a team
// @Description Add a user to a team.
// @tags team
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Team ID"
// @Param team body models.TeamMember true "The user to be added to a team."
// @Success 200 {object} models.TeamMember "The newly created member object"
// @Failure 400 {object} models.HTTPError "Invalid member object provided."
// @Failure 403 {object} models.HTTPError "The user does not have access to the team"
// @Failure 500 {object} models.Message "Internal error"
// @Router /teams/{id}/members [put]
func (tm *TeamMember) Create(doer *User) (err error) {
	// Check if the team extst
	_, err = GetTeamByID(tm.TeamID)
	if err != nil {
		return
	}

	// Check if the user exists
	_, err = GetUserByID(tm.UserID)
	if err != nil {
		return
	}

	// Check if that user is already part of the team
	exists, err := x.Where("team_id = ? AND user_id = ?", tm.TeamID, tm.UserID).
		Get(&TeamMember{})
	if exists {
		return ErrUserIsMemberOfTeam{tm.TeamID, tm.UserID}
	}

	// Insert the user
	_, err = x.Insert(tm)
	return
}
