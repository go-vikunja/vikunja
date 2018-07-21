package models

// Create is the implementation to create a list item
func (i *ListItem) Create(doer *User) (err error) {
	//i.ListID = lID
	i.ID = 0

	return createOrUpdateListItem(i, doer)
}

// Update updates a list item
func (i *ListItem) Update() (err error) {
	// Check if the item exists
	_, err = GetListItemByID(i.ID)
	if err != nil {
		return
	}

	return createOrUpdateListItem(i, &User{})
}

// Helper function for creation or updating of new lists as both methods share most of their logic
func createOrUpdateListItem(i *ListItem, doer *User) (err error) {

	// Check if we have at least a text
	if i.Text == "" {
		return ErrListItemCannotBeEmpty{}
	}

	// Do the update
	if i.ID != 0 {
		_, err = x.ID(i.ID).Update(i)
	} else {
		user, _, err := GetUserByID(doer.ID)
		if err != nil {
			return err
		}

		i.CreatedByID = user.ID
		i.CreatedBy = user
		_, err = x.Insert(i)
	}

	return
}
