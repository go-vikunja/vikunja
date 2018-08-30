package models

// Delete deletes a list <-> user relation
func (lu *ListUser) Delete() (err error) {

	// Check if the user exists
	_, err = GetUserByID(lu.UserID)
	if err != nil {
		return
	}

	// Check if the user has access to the list
	has, err := x.Where("user_id = ? AND list_id = ?", lu.UserID, lu.ListID).
		Get(&ListUser{})
	if err != nil {
		return
	}
	if !has {
		return ErrUserDoesNotHaveAccessToList{ListID: lu.ListID, UserID: lu.UserID}
	}

	_, err = x.Where("user_id = ? AND list_id = ?", lu.UserID, lu.ListID).
		Delete(&ListUser{})
	return
}
