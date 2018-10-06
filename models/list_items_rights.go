package models

// CanDelete checks if the user can delete an task
func (i *ListTask) CanDelete(doer *User) bool {
	// Get the task
	lI, _ := GetListTaskByID(i.ID)

	// A user can delete an task if he has write acces to its list
	list := &List{ID: lI.ListID}
	list.ReadOne()
	return list.CanWrite(doer)
}

// CanUpdate determines if a user has the right to update a list task
func (i *ListTask) CanUpdate(doer *User) bool {
	// Get the task
	lI, _ := GetListTaskByID(i.ID)

	// A user can update an task if he has write acces to its list
	list := &List{ID: lI.ListID}
	list.ReadOne()
	return list.CanWrite(doer)
}

// CanCreate determines if a user has the right to create a list task
func (i *ListTask) CanCreate(doer *User) bool {
	// A user can create an task if he has write acces to its list
	list := &List{ID: i.ListID}
	list.ReadOne()
	return list.CanWrite(doer)
}
