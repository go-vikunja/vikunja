package models

// Delete deletes a team <-> namespace relation based on the namespace & team id
// @Summary Delete a team from a namespace
// @Description Delets a team from a namespace. The team won't have access to the namespace anymore.
// @tags sharing
// @Produce json
// @Security ApiKeyAuth
// @Param namespaceID path int true "Namespace ID"
// @Param teamID path int true "team ID"
// @Success 200 {object} models.Message "The team was successfully deleted."
// @Failure 403 {object} models.HTTPError "The team does not have access to the namespace"
// @Failure 404 {object} models.HTTPError "team or namespace does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{namespaceID}/teams/{teamID} [delete]
func (tn *TeamNamespace) Delete() (err error) {

	// Check if the team exists
	_, err = GetTeamByID(tn.TeamID)
	if err != nil {
		return
	}

	// Check if the team has access to the namespace
	has, err := x.Where("team_id = ? AND namespace_id = ?", tn.TeamID, tn.NamespaceID).
		Get(&TeamNamespace{})
	if err != nil {
		return
	}
	if !has {
		return ErrTeamDoesNotHaveAccessToNamespace{TeamID: tn.TeamID, NamespaceID: tn.NamespaceID}
	}

	// Delete the relation
	_, err = x.Where("team_id = ?", tn.TeamID).
		And("namespace_id = ?", tn.NamespaceID).
		Delete(TeamNamespace{})

	return
}
