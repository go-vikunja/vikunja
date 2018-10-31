package models

import (
	"code.vikunja.io/api/pkg/log"
)

// CanDelete checks if the user can delete an task
func (i *ListTask) CanDelete(doer *User) bool {
	// Get the task
	lI, err := GetListTaskByID(i.ID)
	if err != nil {
		log.Log.Error("Error occurred during CanDelete for ListTask: %s", err)
		return false
	}

	// A user can delete an task if he has write acces to its list
	l := &List{ID: lI.ListID}
	l.ReadOne()
	return l.CanWrite(doer)
}

// CanUpdate determines if a user has the right to update a list task
func (i *ListTask) CanUpdate(doer *User) bool {
	// Get the task
	lI, err := GetListTaskByID(i.ID)
	if err != nil {
		log.Log.Error("Error occurred during CanDelete for ListTask: %s", err)
		return false
	}

	// A user can update an task if he has write acces to its list
	l := &List{ID: lI.ListID}
	l.ReadOne()
	return l.CanWrite(doer)
}

// CanCreate determines if a user has the right to create a list task
func (i *ListTask) CanCreate(doer *User) bool {
	// A user can create an task if he has write acces to its list
	l := &List{ID: i.ListID}
	l.ReadOne()
	return l.CanWrite(doer)
}
