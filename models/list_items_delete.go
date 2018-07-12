package models

// Delete implements the delete method for listItem
func (i *ListItem) Delete(id int64) (err error) {

	// Check if it exists
	_, err = GetListItemByID(id)
	if err != nil {
		return
	}

	_, err = x.ID(id).Delete(ListItem{})
	return
}
