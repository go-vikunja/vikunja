package models

// Create creates a new team <-> namespace relation
// @Summary Add a team to a namespace
// @Description Gives a team access to a namespace.
// @tags sharing
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Namespace ID"
// @Param namespace body models.TeamNamespace true "The team you want to add to the namespace."
// @Success 200 {object} models.TeamNamespace "The created team<->namespace relation."
// @Failure 400 {object} models.HTTPError "Invalid team namespace object provided."
// @Failure 404 {object} models.HTTPError "The team does not exist."
// @Failure 403 {object} models.HTTPError "The team does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id}/teams [put]
func (tn *TeamNamespace) Create(doer *User) (err error) {

	// Check if the rights are valid
	if err = tn.Right.isValid(); err != nil {
		return
	}

	// Check if the team exists
	_, err = GetTeamByID(tn.TeamID)
	if err != nil {
		return
	}

	// Check if the namespace exists
	_, err = GetNamespaceByID(tn.NamespaceID)
	if err != nil {
		return
	}

	// Check if the team already has access to the namespace
	exists, err := x.Where("team_id = ?", tn.TeamID).
		And("namespace_id = ?", tn.NamespaceID).
		Get(&TeamNamespace{})
	if err != nil {
		return
	}
	if exists {
		return ErrTeamAlreadyHasAccess{tn.TeamID, tn.NamespaceID}
	}

	// Insert the new team
	_, err = x.Insert(tn)
	return
}
