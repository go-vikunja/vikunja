package models

// CanCreate checks if the user can create a new user <-> namespace relation
func (nu *NamespaceUser) CanCreate(doer *User) bool {
	// Get the namespace and check if the user has write access on it
	n, _ := GetNamespaceByID(nu.NamespaceID)
	return n.CanWrite(doer)
}

// CanDelete checks if the user can delete a user <-> namespace relation
func (nu *NamespaceUser) CanDelete(doer *User) bool {
	// Get the namespace and check if the user has write access on it
	n, _ := GetNamespaceByID(nu.NamespaceID)
	return n.CanWrite(doer)
}
