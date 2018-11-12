package models

// Update updates a user <-> namespace relation
// @Summary Update a user <-> namespace relation
// @Description Update a user <-> namespace relation. Mostly used to update the right that user has.
// @tags sharing
// @Accept json
// @Produce json
// @Param namespaceID path int true "Namespace ID"
// @Param userID path int true "User ID"
// @Param namespace body models.NamespaceUser true "The user you want to update."
// @Security ApiKeyAuth
// @Success 200 {object} models.NamespaceUser "The updated user <-> namespace relation."
// @Failure 403 {object} models.HTTPError "The user does not have admin-access to the namespace"
// @Failure 404 {object} models.HTTPError "User or namespace does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{namespaceID}/users/{userID} [post]
func (nu *NamespaceUser) Update() (err error) {

	// Check if the right is valid
	if err := nu.Right.isValid(); err != nil {
		return err
	}

	_, err = x.
		Where("namespace_id = ? AND user_id = ?", nu.NamespaceID, nu.UserID).
		Cols("right").
		Update(nu)
	return
}
