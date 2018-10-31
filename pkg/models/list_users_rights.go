package models

import (
	"code.vikunja.io/api/pkg/log"
)

// CanCreate checks if the user can create a new user <-> list relation
func (lu *ListUser) CanCreate(doer *User) bool {
	// Get the list and check if the user has write access on it
	l := List{ID: lu.ListID}
	if err := l.GetSimpleByID(); err != nil {
		log.Log.Error("Error occurred during CanCreate for ListUser: %s", err)
		return false
	}
	return l.CanWrite(doer)
}

// CanDelete checks if the user can delete a user <-> list relation
func (lu *ListUser) CanDelete(doer *User) bool {
	// Get the list and check if the user has write access on it
	l := List{ID: lu.ListID}
	if err := l.GetSimpleByID(); err != nil {
		log.Log.Error("Error occurred during CanDelete for ListUser: %s", err)
		return false
	}
	return l.CanWrite(doer)
}

// CanUpdate checks if the user can update a user <-> list relation
func (lu *ListUser) CanUpdate(doer *User) bool {
	// Get the list and check if the user has write access on it
	l := List{ID: lu.ListID}
	if err := l.GetSimpleByID(); err != nil {
		log.Log.Error("Error occurred during CanUpdate for ListUser: %s", err)
		return false
	}
	return l.CanWrite(doer)
}
