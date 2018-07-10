package models

// Delete implements the delete method of CRUDable
func (l *List) Delete(id int64, doer *User) (err error) {
	// Check if the list exists
	list, err := GetListByID(id)
	if err != nil {
		return
	}

	// Check rights
	user, _, err := GetUserByID(doer.ID)
	if err != nil {
		return
	}

	if !list.IsAdmin(&user) {
		return ErrNeedToBeListAdmin{ListID: id, UserID: user.ID}
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
