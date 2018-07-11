package models

// CreateOrUpdateNamespace does what it says
func CreateOrUpdateNamespace(namespace *Namespace) (err error) {
	// Check if the namespace exists
	_, err = GetNamespaceByID(namespace.ID)
	if err != nil {
		return
	}

	// Check if the User exists
	namespace.Owner, _, err = GetUserByID(namespace.Owner.ID)
	if err != nil {
		return
	}
	namespace.OwnerID = namespace.Owner.ID

	if namespace.ID == 0 {
		_, err = x.Insert(namespace)
	} else {
		_, err = x.ID(namespace.ID).Update(namespace)
	}

	if err != nil {
		return
	}

	// Get the new one
	*namespace, err = GetNamespaceByID(namespace.ID)

	return
}

func (n *Namespace) Create(doer *User, _ int64) (err error) {
	// Check if we have at least a name
	if n.Name == "" {
		return ErrNamespaceNameCannotBeEmpty{NamespaceID:0, UserID:doer.ID}
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
