package models

// CreateOrUpdateListItem adds or updates a todo item to a list
func CreateOrUpdateListItem(item *ListItem) (newItem *ListItem, err error) {

	// Check if the list exists
	_, err = GetListByID(item.ListID)
	if err != nil {
		return
	}

	// Check if the user exists
	user, _, err := GetUserByID(item.CreatedBy.ID)
	if err != nil {
		return
	}
	item.CreatedByID = item.CreatedBy.ID
	item.CreatedBy = user

	if item.ID != 0 {
		_, err = x.ID(item.ID).Update(item)
		if err != nil {
			return
		}
	} else {
		_, err = x.Insert(item)
		if err != nil {
			return
		}

		// Check if we have at least a text
		if item.Text == "" {
			return newItem, ErrListItemCannotBeEmpty{}
		}
	}

	// Get the new/updated item
	finalItem, err := GetListItemByID(item.ID)

	return &finalItem, err
}
