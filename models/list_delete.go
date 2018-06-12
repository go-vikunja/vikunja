package models

func DeleteListByID(listID int64, doer *User) (err error) {

	// Check if the list exists
	list, err := GetListByID(listID)
	if err != nil {
		return
	}

	if list.Owner.ID != doer.ID {
		return ErrNeedToBeListOwner{ListID:listID, UserID:doer.ID}
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
