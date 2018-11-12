package models

// Delete deletes a namespace <-> user relation
// @Summary Delete a user from a namespace
// @Description Delets a user from a namespace. The user won't have access to the namespace anymore.
// @tags sharing
// @Produce json
// @Security ApiKeyAuth
// @Param namespaceID path int true "Namespace ID"
// @Param userID path int true "user ID"
// @Success 200 {object} models.Message "The user was successfully deleted."
// @Failure 403 {object} models.HTTPError "The user does not have access to the namespace"
// @Failure 404 {object} models.HTTPError "user or namespace does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{namespaceID}/users/{userID} [delete]
func (nu *NamespaceUser) Delete() (err error) {

	// Check if the user exists
	_, err = GetUserByID(nu.UserID)
	if err != nil {
		return
	}

	// Check if the user has access to the namespace
	has, err := x.Where("user_id = ? AND namespace_id = ?", nu.UserID, nu.NamespaceID).
		Get(&NamespaceUser{})
	if err != nil {
		return
	}
	if !has {
		return ErrUserDoesNotHaveAccessToNamespace{NamespaceID: nu.NamespaceID, UserID: nu.UserID}
	}

	_, err = x.Where("user_id = ? AND namespace_id = ?", nu.UserID, nu.NamespaceID).
		Delete(&NamespaceUser{})
	return
}
