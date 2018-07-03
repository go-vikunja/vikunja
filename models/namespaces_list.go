package models

// CreateOrUpdateNamespace does what it says
func CreateOrUpdateNamespace(namespace *Namespace) (err error) {
	// Check if the User exists
	_, _, err = GetUserByID(namespace.Owner.ID)
	if err != nil {
		return
	}

	namespace.OwnerID = namespace.Owner.ID

	if namespace.ID == 0 {
		_, err = x.Insert(namespace)
		if err != nil {
			return
		}
	} else {
		_, err = x.ID(namespace.ID).Update(namespace)
		if err != nil {
			return
		}
	}

	return
}

// GetAllNamespacesByUserID does what it says
func GetAllNamespacesByUserID(userID int64) (namespaces []Namespace, err error) {

	// First, get all namespaces which that user owns
	err = x.Where("owner_id = ?", userID).Find(&namespaces)
	if err != nil {
		return namespaces, err
	}

	// Get all namespaces of teams that user is part of
	/*err = x.Table("namespaces").
	Join("INNER", ).
	Find(namespaces)*/

	// Get user objects
	// I couldn't come up with a more performant way to do this...
	for in, n := range namespaces {
		namespaces[in].Owner, _, err = GetUserByID(n.OwnerID)
		if err != nil {
			return nil, err
		}
	}

	return
}
