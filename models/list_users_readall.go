package models

// ReadAll gets all users who have access to a list
func (ul *ListUser) ReadAll(user *User) (interface{}, error) {
	// Check if the user has access to the list
	l, err := GetListByID(ul.ListID)
	if err != nil {
		return nil, err
	}
	if !l.CanRead(user) {
		return nil, ErrNeedToHaveListReadAccess{}
	}

	// Get all users
	all := []*User{}
	err = x.
		Select("users.*").
		Join("INNER", "users_list", "user_id = users.id").
		Where("users_list.list_id = ?", ul.ListID).
		Find(&all)

	return all, err
}
