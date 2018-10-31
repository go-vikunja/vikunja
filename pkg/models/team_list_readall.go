package models

// ReadAll implements the method to read all teams of a list
func (tl *TeamList) ReadAll(u *User) (interface{}, error) {
	// Check if the user can read the namespace
	l := &List{ID: tl.ListID}
	if err := l.GetSimpleByID(); err != nil {
		return nil, err
	}
	if !l.CanRead(u) {
		return nil, ErrNeedToHaveListReadAccess{ListID: tl.ListID, UserID: u.ID}
	}

	// Get the teams
	all := []*teamWithRight{}
	err := x.
		Table("teams").
		Join("INNER", "team_list", "team_id = teams.id").
		Where("team_list.list_id = ?", tl.ListID).
		Find(&all)

	return all, err
}
