package models

// CreateOrUpdateListItem adds or updates a todo item to a list
func CreateOrUpdateListItem(item *ListItem) (newItem *ListItem, err error) {

	// Check if the list exists
	_, err = GetListByID(item.ListID)
	if err != nil {
		return
	}

	// Check if the user exists
	item.CreatedBy, _, err = GetUserByID(item.CreatedBy.ID)
	if err != nil {
		return
	}
	item.CreatedByID = item.CreatedBy.ID

	if item.ID != 0 {
		_, err = x.ID(item.ID).Update(item)
		if err != nil {
			return
		}
	} else {
		// Check if we have at least a text
		if item.Text == "" {
			return newItem, ErrListItemCannotBeEmpty{}
		}

		_, err = x.Insert(item)
		if err != nil {
			return
		}
	}

	// Get the new/updated item
	finalItem, err := GetListItemByID(item.ID)

	return &finalItem, err
}

// Create is the implementation to create a list item
func (i *ListItem) Create(doer *User, lID int64) (err error) {
	i.ListID = lID

	// Check rights
	user, _, err := GetUserByID(doer.ID)
	if err != nil {
		return
	}
	i.CreatedBy = user // Needed because we return the full item object
	i.CreatedByID = user.ID

	// Get the list to check if the user has the right to write to that list
	list, err := GetListByID(lID)
	if err != nil {
		return
	}

	if !list.CanWrite(&user) {
		return ErrNeedToBeListWriter{ListID: lID, UserID: user.ID}
	}

	// Check if we have at least a text
	if i.Text == "" {
		return ErrListItemCannotBeEmpty{}
	}

	_, err = x.Insert(i)
	if err != nil {
		return
	}

	return
}
