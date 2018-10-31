package models

// Update updates a team <-> list relation
func (tl *TeamList) Update() (err error) {

	// Check if the right is valid
	if err := tl.Right.isValid(); err != nil {
		return err
	}

	_, err = x.
		Where("list_id = ? AND team_id = ?", tl.ListID, tl.TeamID).
		Cols("right").
		Update(tl)
	return
}
