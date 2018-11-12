package models

// Delete deletes a namespace
// @Summary Deletes a namespace
// @Description Delets a namespace
// @tags namespace
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Namespace ID"
// @Success 200 {object} models.Message "The namespace was successfully deleted."
// @Failure 400 {object} models.HTTPError "Invalid namespace object provided."
// @Failure 403 {object} models.HTTPError "The user does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id} [delete]
func (n *Namespace) Delete() (err error) {

	// Check if the namespace exists
	_, err = GetNamespaceByID(n.ID)
	if err != nil {
		return
	}

	// Delete the namespace
	_, err = x.ID(n.ID).Delete(&Namespace{})
	if err != nil {
		return
	}

	// Delete all lists with their tasks
	lists, err := GetListsByNamespaceID(n.ID)
	var listIDs []int64
	// We need to do that for here because we need the list ids to delete two times:
	// 1) to delete the lists itself
	// 2) to delete the list tasks
	for _, l := range lists {
		listIDs = append(listIDs, l.ID)
	}

	// Delete tasks
	_, err = x.In("list_id", listIDs).Delete(&ListTask{})
	if err != nil {
		return
	}

	// Delete the lists
	_, err = x.In("id", listIDs).Delete(&List{})
	if err != nil {
		return
	}

	return
}
