package models

// Create creates a new list <-> user relation
// @Summary Add a user to a list
// @Description Gives a user access to a list.
// @tags sharing
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "List ID"
// @Param list body models.ListUser true "The user you want to add to the list."
// @Success 200 {object} models.ListUser "The created user<->list relation."
// @Failure 400 {object} models.HTTPError "Invalid user list object provided."
// @Failure 404 {object} models.HTTPError "The user does not exist."
// @Failure 403 {object} models.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id}/users [put]
func (ul *ListUser) Create(u *User) (err error) {

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
