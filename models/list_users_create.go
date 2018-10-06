package models

// Create creates a new list <-> user relation
func (ul *ListUser) Create(user *User) (err error) {

	// Check if the right is valid
	if err := ul.Right.isValid(); err != nil {
		return err
	}

	// Check if the list exists
	l := &List{ID: ul.ListID}
	if err = l.GetSimpleByID(); err != nil {
		return
	}

	// Check if the user exists
	if _, err = GetUserByID(ul.UserID); err != nil {
		return err
	}

	// Check if the user already has access or is owner of that list
	// We explicitly DONT check for teams here
	if l.OwnerID == ul.UserID {
		return ErrUserAlreadyHasAccess{UserID: ul.UserID, ListID: ul.ListID}
	}

	exist, err := x.Where("list_id = ? AND user_id = ?", ul.ListID, ul.UserID).Get(&ListUser{})
	if err != nil {
		return
	}
	if exist {
		return ErrUserAlreadyHasAccess{UserID: ul.UserID, ListID: ul.ListID}
	}

	// Insert user <-> list relation
	_, err = x.Insert(ul)

	return
}
