package models

// Delete deletes a namespace <-> user relation
func (nu *NamespaceUser) Delete() (err error) {

	// Check if the user exists
	_, err = GetUserByID(nu.UserID)
	if err != nil {
		return
	}

	// Check if the user has access to the namespace
	has, err := x.Where("user_id = ? AND namespace_id = ?", nu.UserID, nu.NamespaceID).
		Get(&NamespaceUser{})
	if err != nil {
		return
	}
	if !has {
		return ErrUserDoesNotHaveAccessToNamespace{NamespaceID: nu.NamespaceID, UserID: nu.UserID}
	}

	_, err = x.Where("user_id = ? AND namespace_id = ?", nu.UserID, nu.NamespaceID).
		Delete(&NamespaceUser{})
	return
}
