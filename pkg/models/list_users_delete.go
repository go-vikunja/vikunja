package models

// Delete deletes a list <-> user relation
// @Summary Delete a user from a list
// @Description Delets a user from a list. The user won't have access to the list anymore.
// @tags sharing
// @Produce json
// @Security ApiKeyAuth
// @Param listID path int true "List ID"
// @Param userID path int true "User ID"
// @Success 200 {object} models.Message "The user was successfully removed from the list."
// @Failure 403 {object} models.HTTPError "The user does not have access to the list"
// @Failure 404 {object} models.HTTPError "user or list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/users/{userID} [delete]
func (lu *ListUser) Delete() (err error) {

	// Check if the user exists
	_, err = GetUserByID(lu.UserID)
	if err != nil {
		return
	}

	// Check if the user has access to the list
	has, err := x.Where("user_id = ? AND list_id = ?", lu.UserID, lu.ListID).
		Get(&ListUser{})
	if err != nil {
		return
	}
	if !has {
		return ErrUserDoesNotHaveAccessToList{ListID: lu.ListID, UserID: lu.UserID}
	}

	_, err = x.Where("user_id = ? AND list_id = ?", lu.UserID, lu.ListID).
		Delete(&ListUser{})
	return
}
