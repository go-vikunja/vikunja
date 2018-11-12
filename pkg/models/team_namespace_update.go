package models

// Update updates a team <-> namespace relation
// @Summary Update a team <-> namespace relation
// @Description Update a team <-> namespace relation. Mostly used to update the right that team has.
// @tags sharing
// @Accept json
// @Produce json
// @Param namespaceID path int true "Namespace ID"
// @Param teamID path int true "Team ID"
// @Param namespace body models.TeamNamespace true "The team you want to update."
// @Security ApiKeyAuth
// @Success 200 {object} models.TeamNamespace "The updated team <-> namespace relation."
// @Failure 403 {object} models.HTTPError "The team does not have admin-access to the namespace"
// @Failure 404 {object} models.HTTPError "Team or namespace does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{namespaceID}/teams/{teamID} [post]
func (tl *TeamNamespace) Update() (err error) {

	// Check if the right is valid
	if err := tl.Right.isValid(); err != nil {
		return err
	}

	_, err = x.
		Where("namespace_id = ? AND team_id = ?", tl.TeamID, tl.TeamID).
		Cols("right").
		Update(tl)
	return
}
