package models

// ReadAll implements the method to read all teams of a list
func (tl *TeamList) ReadAll(user *User) (interface{}, error) {
	// Check if the user can read the namespace
	l, err := GetListByID(tl.ListID)
	if err != nil {
		return nil, err
	}
	if !l.CanRead(user) {
		return nil, ErrNeedToHaveListReadAccess{ListID: tl.ListID, UserID: user.ID}
	}

	// Get the teams
	all := []*Team{}

	err = x.Select("teams.*").
		Table("teams").
		Join("INNER", "team_list", "team_id = teams.id").
		Where("team_list.list_id = ?", tl.ListID).
		Find(&all)

	return all, err
}
