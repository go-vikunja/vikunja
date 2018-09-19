package models

// Update updates a user <-> list relation
func (lu *ListUser) Update() (err error) {

	// Check if the right is valid
	if err := lu.Right.isValid(); err != nil {
		return err
	}

	_, err = x.
		Where("list_id = ? AND user_id = ?", lu.ListID, lu.UserID).
		Cols("right").
		Update(lu)
	return
}
