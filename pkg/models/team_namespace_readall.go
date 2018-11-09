package models

// ReadAll implements the method to read all teams of a namespace
func (tn *TeamNamespace) ReadAll(user *User, page int) (interface{}, error) {
	// Check if the user can read the namespace
	n, err := GetNamespaceByID(tn.NamespaceID)
	if err != nil {
		return nil, err
	}
	if !n.CanRead(user) {
		return nil, ErrNeedToHaveNamespaceReadAccess{NamespaceID: tn.NamespaceID, UserID: user.ID}
	}

	// Get the teams
	all := []*teamWithRight{}

	err = x.Table("teams").
		Join("INNER", "team_namespaces", "team_id = teams.id").
		Where("team_namespaces.namespace_id = ?", tn.NamespaceID).
		Limit(getLimitFromPageIndex(page)).
		Find(&all)

	return all, err
}
