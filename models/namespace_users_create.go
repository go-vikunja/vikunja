package models

// Create creates a new namespace <-> user relation
func (un *NamespaceUser) Create(user *User) (err error) {

	// Reset the id
	un.ID = 0

	// Check if the right is valid
	if err := un.Right.isValid(); err != nil {
		return err
	}

	// Check if the namespace exists
	l, err := GetNamespaceByID(un.NamespaceID)
	if err != nil {
		return
	}

	// Check if the user exists
	if _, err = GetUserByID(un.UserID); err != nil {
		return err
	}

	// Check if the user already has access or is owner of that namespace
	// We explicitly DO NOT check for teams here
	if l.OwnerID == un.UserID {
		return ErrUserAlreadyHasNamespaceAccess{UserID: un.UserID, NamespaceID: un.NamespaceID}
	}

	exist, err := x.Where("namespace_id = ? AND user_id = ?", un.NamespaceID, un.UserID).Get(&NamespaceUser{})
	if err != nil {
		return
	}
	if exist {
		return ErrUserAlreadyHasNamespaceAccess{UserID: un.UserID, NamespaceID: un.NamespaceID}
	}

	// Insert user <-> namespace relation
	_, err = x.Insert(un)

	return
}
