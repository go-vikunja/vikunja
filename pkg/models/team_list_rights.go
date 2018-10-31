package models

import (
	"code.vikunja.io/api/pkg/log"
)

// CanCreate checks if the user can create a team <-> list relation
func (tl *TeamList) CanCreate(u *User) bool {
	l := List{ID: tl.ListID}
	if err := l.GetSimpleByID(); err != nil {
		log.Log.Error("Error occurred during CanCreate for TeamList: %s", err)
		return false
	}
	return l.IsAdmin(u)
}

// CanDelete checks if the user can delete a team <-> list relation
func (tl *TeamList) CanDelete(user *User) bool {
	l := List{ID: tl.ListID}
	if err := l.GetSimpleByID(); err != nil {
		log.Log.Error("Error occurred during CanDelete for TeamList: %s", err)
		return false
	}
	return l.IsAdmin(user)
}

// CanUpdate checks if the user can update a team <-> list relation
func (tl *TeamList) CanUpdate(user *User) bool {
	l := List{ID: tl.ListID}
	if err := l.GetSimpleByID(); err != nil {
		log.Log.Error("Error occurred during CanUpdate for TeamList: %s", err)
		return false
	}
	return l.IsAdmin(user)
}
