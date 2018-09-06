package models

// IsAdmin returns true or false if the user is admin on that namespace or not
func (n *Namespace) IsAdmin(user *User) bool {
	// Owners always have admin rights
	if user.ID == n.Owner.ID {
		return true
	}

	// Check user rights
	if n.checkUserRights(user, UserRightAdmin){
		return true
	}

	// Check if that user is in a team which has admin rights to that namespace
	return n.checkTeamRights(user, TeamRightAdmin)
}

// CanWrite checks if a user has write access to a namespace
func (n *Namespace) CanWrite(user *User) bool {
	// Admins always have write access
	if n.IsAdmin(user) {
		return true
	}

	// Check user rights
	if n.checkUserRights(user, UserRightWrite){
		return true
	}

	// Check if that user is in a team which has write rights to that namespace
	return n.checkTeamRights(user, TeamRightWrite)
}

// CanRead checks if a user has read access to that namespace
func (n *Namespace) CanRead(user *User) bool {
	// Admins always have read access
	if n.IsAdmin(user) {
		return true
	}

	// Check user rights
	if n.checkUserRights(user, UserRightRead){
		return true
	}

	// Check if the user is in a team which has access to the namespace
	return n.checkTeamRights(user, TeamRightRead)
}

// CanUpdate checks if the user can update the namespace
func (n *Namespace) CanUpdate(user *User) bool {
	nn, _ := GetNamespaceByID(n.ID)
	return nn.IsAdmin(user)
}

// CanDelete checks if the user can delete a namespace
func (n *Namespace) CanDelete(user *User) bool {
	nn, _ := GetNamespaceByID(n.ID)
	return nn.IsAdmin(user)
}

// CanCreate checks if the user can create a new namespace
func (n *Namespace) CanCreate(user *User) bool {
	// This is currently a dummy function, later on we could imagine global limits etc.
	return true
}

func (n *Namespace) checkTeamRights(user *User, r TeamRight) bool {
	exists, err := x.Select("namespaces.*").
		Table("namespaces").
		Join("LEFT", "team_namespaces", "namespaces.id = team_namespaces.namespace_id").
		Join("LEFT", "team_members", "team_members.team_id = team_namespaces.team_id").
		Where("namespaces.id = ? " +
			"AND ((team_members.user_id = ?  AND team_namespaces.right = ?) " +
			"OR namespaces.owner_id = ? ", n.ID, user.ID, r, user.ID).
		Get(&Namespace{})

	if err != nil {
		return false
	}

	return exists
}

func (n *Namespace) checkUserRights(user *User, r UserRight) bool {
	exists, err := x.Select("namespaces.*").
		Table("namespaces").
		Join("LEFT", "users_namespace", "users_namespace.namespace_id = namespaces.id").
		Where("namespaces.id = ? " +
			"OR namespaces.owner_id = ? " +
			"OR (users_namespace.user_id = ? AND users_namespace.right = ?))", n.ID, user.ID, user.ID, r).
		Get(&Namespace{})

	if err != nil {
		return false
	}

	return exists
}
