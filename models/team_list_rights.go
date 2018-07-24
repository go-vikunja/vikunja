package models

// CanCreate checks if the use can create a team <-> list relation
func (tl *TeamList) CanCreate(user *User) bool {
	l, _ := GetListByID(tl.ListID)
	return l.IsAdmin(user)
}
