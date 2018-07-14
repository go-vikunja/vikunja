package models

// Update implements the update method via the interface
func (n *Namespace) Update(id int64) (err error) {
	// Check if we have at least a name
	if n.Name == "" {
		return ErrNamespaceNameCannotBeEmpty{NamespaceID: id}
	}
	n.ID = id

	// Check if the namespace exists
	currentNamespace, err := GetNamespaceByID(id)
	if err != nil {
		return
	}

	// Check if the (new) owner exists
	if currentNamespace.OwnerID != n.OwnerID {
		n.Owner, _, err = GetUserByID(n.OwnerID)
		if err != nil {
			return
		}
	}

	// Do the actual update
	_, err = x.ID(currentNamespace.ID).Update(n)
	return
}
