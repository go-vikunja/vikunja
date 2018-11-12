package models

// Create creates a new namespace <-> user relation
// @Summary Add a user to a namespace
// @Description Gives a user access to a namespace.
// @tags sharing
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Namespace ID"
// @Param namespace body models.NamespaceUser true "The user you want to add to the namespace."
// @Success 200 {object} models.NamespaceUser "The created user<->namespace relation."
// @Failure 400 {object} models.HTTPError "Invalid user namespace object provided."
// @Failure 404 {object} models.HTTPError "The user does not exist."
// @Failure 403 {object} models.HTTPError "The user does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id}/users [put]
func (un *NamespaceUser) Create(u *User) (err error) {

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
