package models

// CanDelete checks if the user can delete an item
func (i *ListItem) CanDelete(doer *User, id int64) bool {
	// Get the item
	lI, _ := GetListItemByID(id)

	// A user can delete an item if he has write acces to its list
	list, _ := GetListByID(lI.ListID)
	return list.CanWrite(doer)
}

// CanUpdate determines if a user has the right to update a list item
func (i *ListItem) CanUpdate(doer *User, id int64) bool {
	// Get the item
	lI, _ := GetListItemByID(id)

	// A user can update an item if he has write acces to its list
	list, _ := GetListByID(lI.ListID)
	return list.CanWrite(doer)
}

// CanCreate determines if a user has the right to create a list item
func (i *ListItem) CanCreate(doer *User, lID int64) bool {
	// A user can create an item if he has write acces to its list
	list, _ := GetListByID(lID)
	return list.CanWrite(doer)
}
