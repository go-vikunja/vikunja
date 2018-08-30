package models

// Update implements the update method via the interface
func (n *Namespace) Update() (err error) {
	// Check if we have at least a name
	if n.Name == "" {
		return ErrNamespaceNameCannotBeEmpty{NamespaceID: n.ID}
	}

	// Check if the namespace exists
	currentNamespace, err := GetNamespaceByID(n.ID)
	if err != nil {
		return
	}

	// Check if the (new) owner exists
	if currentNamespace.OwnerID != n.OwnerID {
		n.Owner, err = GetUserByID(n.OwnerID)
		if err != nil {
			return
		}
	}

	// Do the actual update
	_, err = x.ID(currentNamespace.ID).Update(n)
	return
}
