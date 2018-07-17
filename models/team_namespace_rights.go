package models

// CanCreate checks if one can create a new team <-> namespace relation
func (tn *TeamNamespace) CanCreate(user *User, _ int64) bool {
	n, _ := GetNamespaceByID(tn.NamespaceID)
	return n.IsAdmin(user)
}
