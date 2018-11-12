package models

// Delete implements the delete method of CRUDable
// @Summary Deletes a list
// @Description Delets a list
// @tags list
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "List ID"
// @Success 200 {object} models.Message "The list was successfully deleted."
// @Failure 400 {object} models.HTTPError "Invalid list object provided."
// @Failure 403 {object} models.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id} [delete]
func (l *List) Delete() (err error) {
	// Check if the list exists
	if err = l.GetSimpleByID(); err != nil {
		return
	}

	// Delete the list
	_, err = x.ID(l.ID).Delete(&List{})
	if err != nil {
		return
	}

	// Delete all todotasks on that list
	_, err = x.Where("list_id = ?", l.ID).Delete(&ListTask{})
	return
}
