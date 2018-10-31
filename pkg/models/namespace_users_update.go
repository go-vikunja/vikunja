package models

// Update updates a user <-> namespace relation
func (nu *NamespaceUser) Update() (err error) {

	// Check if the right is valid
	if err := nu.Right.isValid(); err != nil {
		return err
	}

	_, err = x.
		Where("namespace_id = ? AND user_id = ?", nu.NamespaceID, nu.UserID).
		Cols("right").
		Update(nu)
	return
}
