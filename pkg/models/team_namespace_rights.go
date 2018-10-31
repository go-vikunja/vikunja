package models

import (
	"code.vikunja.io/api/pkg/log"
)

// CanCreate checks if one can create a new team <-> namespace relation
func (tn *TeamNamespace) CanCreate(user *User) bool {
	n, err := GetNamespaceByID(tn.NamespaceID)
	if err != nil {
		log.Log.Error("Error occurred during CanCreate for TeamNamespace: %s", err)
		return false
	}
	return n.IsAdmin(user)
}

// CanDelete checks if a user can remove a team from a namespace. Only namespace admins can do that.
func (tn *TeamNamespace) CanDelete(user *User) bool {
	n, err := GetNamespaceByID(tn.NamespaceID)
	if err != nil {
		log.Log.Error("Error occurred during CanDelete for TeamNamespace: %s", err)
		return false
	}
	return n.IsAdmin(user)
}

// CanUpdate checks if a user can update a team from a  Only namespace admins can do that.
func (tn *TeamNamespace) CanUpdate(user *User) bool {
	n, err := GetNamespaceByID(tn.NamespaceID)
	if err != nil {
		log.Log.Error("Error occurred during CanUpdate for TeamNamespace: %s", err)
		return false
	}
	return n.IsAdmin(user)
}
