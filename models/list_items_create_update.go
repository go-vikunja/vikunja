package models

// Create is the implementation to create a list item
func (i *ListItem) Create(doer *User, lID int64) (err error) {
	i.ListID = lID
	i.ID = 0

	return createOrUpdateListItem(i, doer, lID)
}

// Update updates a list item
func (i *ListItem) Update(ID int64, doer *User) (err error) {
	i.ID = ID

	// Get the full item
	fullItem, err := GetListItemByID(ID)
	if err != nil {
		return
	}

	return createOrUpdateListItem(i, doer, fullItem.ListID)
}

// Helper function for creation or updating of new lists as both methods share most of their logic
func createOrUpdateListItem(i *ListItem, doer *User, lID int64) (err error) {
	// Check rights
	user, _, err := GetUserByID(doer.ID)
	if err != nil {
		return
	}

	// Get the list to check if the user has the right to write to that list
	list, err := GetListByID(lID) // TODO: Get the list with one query by item ID
	if err != nil {
		return
	}

	if !list.CanWrite(&user) {
		return ErrNeedToBeListWriter{ListID: i.ListID, UserID: user.ID}
	}

	// Check if we have at least a text
	if i.Text == "" {
		return ErrListItemCannotBeEmpty{}
	}

	// Do the update
	if i.ID != 0 {
		_, err = x.ID(i.ID).Update(i)
	} else {
		i.CreatedByID = user.ID
		i.CreatedBy = user
		_, err = x.Insert(i)
	}

	return
}
