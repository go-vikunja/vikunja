package models

// CanCreate checks if one can create a new team <-> namespace relation
func (tn *TeamNamespace) CanCreate(user *User) bool {
	n, _ := GetNamespaceByID(tn.NamespaceID)
	return n.IsAdmin(user)
}

// CanDelete checks if a user can remove a team from a namespace. Only namespace admins can do that.
func (tn *TeamNamespace) CanDelete(user *User) bool {
	n, _ := GetNamespaceByID(tn.NamespaceID)
	return n.IsAdmin(user)
}