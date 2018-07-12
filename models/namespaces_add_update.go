package models

// Create implements the creation method via the interface
func (n *Namespace) Create(doer *User, _ int64) (err error) {
	// Check if we have at least a name
	if n.Name == "" {
		return ErrNamespaceNameCannotBeEmpty{NamespaceID: 0, UserID: doer.ID}
	}

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
