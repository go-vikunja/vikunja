package models

// Delete implements the delete method for listItem
func (i *ListItem) Delete() (err error) {

	// Check if it exists
	_, err = GetListItemByID(i.ID)
	if err != nil {
		return
	}

	_, err = x.ID(i.ID).Delete(ListItem{})
	return
}
