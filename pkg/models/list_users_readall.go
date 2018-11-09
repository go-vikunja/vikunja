package models

// ReadAll gets all users who have access to a list
func (ul *ListUser) ReadAll(u *User, page int) (interface{}, error) {
	// Check if the user has access to the list
	l := &List{ID: ul.ListID}
	if err := l.GetSimpleByID(); err != nil {
		return nil, err
	}
	if !l.CanRead(u) {
		return nil, ErrNeedToHaveListReadAccess{}
	}

	// Get all users
	all := []*userWithRight{}
	err := x.
		Join("INNER", "users_list", "user_id = users.id").
		Where("users_list.list_id = ?", ul.ListID).
		Limit(getLimitFromPageIndex(page)).
		Find(&all)

	return all, err
}
