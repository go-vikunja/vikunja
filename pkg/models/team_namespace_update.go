package models

// Update updates a team <-> namespace relation
func (tl *TeamNamespace) Update() (err error) {

	// Check if the right is valid
	if err := tl.Right.isValid(); err != nil {
		return err
	}

	_, err = x.
		Where("namespace_id = ? AND team_id = ?", tl.TeamID, tl.TeamID).
		Cols("right").
		Update(tl)
	return
}
