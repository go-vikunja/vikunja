package models

// CanCreate checks if the user can create a team <-> list relation
func (tl *TeamList) CanCreate(user *User) bool {
	l, _ := GetListByID(tl.ListID)
	return l.IsAdmin(user)
}

// CanDelete checks if the user can delete a team <-> list relation
func (tl *TeamList) CanDelete(user *User) bool {
	l, _ := GetListByID(tl.ListID)
	return l.IsAdmin(user)
}
