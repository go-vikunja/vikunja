package models

// CreateOrUpdateList updates a list or creates it if it doesn't exist
func CreateOrUpdateList(list *List) (err error) {
	// Check if it exists
	_, err = GetListByID(list.ID)
	if err != nil {
		return
	}

	// Check if the user exists
	list.Owner, _, err = GetUserByID(list.Owner.ID)
	if err != nil {
		return
	}
	list.OwnerID = list.Owner.ID

	if list.ID == 0 {
		_, err = x.Insert(list)
	} else {
		_, err = x.ID(list.ID).Update(list)
	}

	if err != nil {
		return
	}

	*list, err = GetListByID(list.ID)

	return

}
