package models

// Delete implements the delete method of CRUDable
func (l *List) Delete(id int64) (err error) {
	// Check if the list exists
	_, err = GetListByID(id)
	if err != nil {
		return
	}

	// Delete the list
	_, err = x.ID(id).Delete(&List{})
	if err != nil {
		return
	}

	// Delete all todoitems on that list
	_, err = x.Where("list_id = ?", id).Delete(&ListItem{})
	return
}
