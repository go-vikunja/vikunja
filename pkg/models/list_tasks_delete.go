package models

// Delete implements the delete method for listTask
// @Summary Delete a task
// @Description Deletes a task from a list. This does not mean "mark it done".
// @tags task
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Task ID"
// @Success 200 {object} models.Message "The created task object."
// @Failure 400 {object} models.HTTPError "Invalid task ID provided."
// @Failure 403 {object} models.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{id} [delete]
func (i *ListTask) Delete() (err error) {

	// Check if it exists
	_, err = GetListTaskByID(i.ID)
	if err != nil {
		return
	}

	_, err = x.ID(i.ID).Delete(ListTask{})
	return
}
