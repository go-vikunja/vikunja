package models

// ReadAll gets all users who have access to a namespace
func (un *NamespaceUser) ReadAll(u *User) (interface{}, error) {
	// Check if the user has access to the namespace
	l, err := GetNamespaceByID(un.NamespaceID)
	if err != nil {
		return nil, err
	}
	if !l.CanRead(u) {
		return nil, ErrNeedToHaveNamespaceReadAccess{}
	}

	// Get all users
	all := []*userWithRight{}
	err = x.
		Join("INNER", "users_namespace", "user_id = users.id").
		Where("users_namespace.namespace_id = ?", un.NamespaceID).
		Find(&all)

	return all, err
}
