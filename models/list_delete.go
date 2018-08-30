package models

// Delete implements the delete method of CRUDable
func (l *List) Delete() (err error) {
	// Check if the list exists
	_, err = GetListByID(l.ID)
	if err != nil {
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
