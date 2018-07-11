package models

// Delete implements the delete method for listItem
func (i *ListItem) Delete(id int64, doer *User) (err error) {

	// Check if it exists
	listitem, err := GetListItemByID(id)
	if err != nil {
		return
	}

	// Check if the user hat the right to delete that item
	_, err = listItemPreCheck(i, doer, listitem.ListID)
	if err != nil {
		return
	}

	_, err = x.ID(id).Delete(ListItem{})
	return
}
