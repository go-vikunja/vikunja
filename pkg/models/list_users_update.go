package models

// Update updates a user <-> list relation
// @Summary Update a user <-> list relation
// @Description Update a user <-> list relation. Mostly used to update the right that user has.
// @tags sharing
// @Accept json
// @Produce json
// @Param listID path int true "List ID"
// @Param userID path int true "User ID"
// @Param list body models.ListUser true "The user you want to update."
// @Security ApiKeyAuth
// @Success 200 {object} models.ListUser "The updated user <-> list relation."
// @Failure 403 {object} models.HTTPError "The user does not have admin-access to the list"
// @Failure 404 {object} models.HTTPError "User or list does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{listID}/users/{userID} [post]
func (lu *ListUser) Update() (err error) {

	// Check if the right is valid
	if err := lu.Right.isValid(); err != nil {
		return err
	}

	_, err = x.
		Where("list_id = ? AND user_id = ?", lu.ListID, lu.UserID).
		Cols("right").
		Update(lu)
	return
}
