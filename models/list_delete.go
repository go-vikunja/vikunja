package models

func DeleteListByID(listID int64) (err error) {

	// Check if the list exists
	_, err = GetListByID(listID)
	if err != nil {
		return
	}

	// Delete the list
	_, err = x.ID(listID).Delete(&List{})
	if err != nil {
		return
	}

	// Delete all todoitems on that list
	_, err = x.Where("list_id = ?", listID).Delete(&ListItem{})

	return
}
