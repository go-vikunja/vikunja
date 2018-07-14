package models

// Create implements the creation method via the interface
func (n *Namespace) Create(doer *User, _ int64) (err error) {
	// Check if we have at least a name
	if n.Name == "" {
		return ErrNamespaceNameCannotBeEmpty{NamespaceID: 0, UserID: doer.ID}
	}
	n.ID = 0 // This would otherwise prevent the creation of new lists after one was created

	// Check if the User exists
	n.Owner, _, err = GetUserByID(doer.ID)
	if err != nil {
		return
	}
	n.OwnerID = n.Owner.ID

	// Insert
	_, err = x.Insert(n)
	return
}
