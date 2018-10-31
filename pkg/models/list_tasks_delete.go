package models

// Delete implements the delete method for listTask
func (i *ListTask) Delete() (err error) {

	// Check if it exists
	_, err = GetListTaskByID(i.ID)
	if err != nil {
		return
	}

	_, err = x.ID(i.ID).Delete(ListTask{})
	return
}
