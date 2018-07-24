package models

// IsAdmin returns true or false if the user is admin on that namespace or not
func (n *Namespace) IsAdmin(user *User) bool {

	// Owners always have admin rights
	if user.ID == n.Owner.ID {
		return true
	}

	// Check if that user is in a team which has admin rights to that namespace
	// TODO

	return false
}

// CanWrite checks if a user has write access to a namespace
func (n *Namespace) CanWrite(user *User) bool {
	// Owners always have access
	if user.ID == n.Owner.ID {
		return true
	}

	return true
}

// CanRead checks if a user has read access to that namespace
func (n *Namespace) CanRead(user *User) bool {
	// Owners always have access
	if user.ID == n.Owner.ID {
		return true
	}

	// Admins always have read access
	if n.IsAdmin(user) {
		return true
	}

	// Check if the user is in a team which has access to the namespace
	all := Namespace{}
	// TODO respect individual rights
	exists, _ := x.Select("namespaces.*").
		Table("namespaces").
		Join("LEFT", "team_namespaces", "namespaces.id = team_namespaces.namespace_id").
		Join("LEFT", "team_members", "team_members.team_id = team_namespaces.team_id").
		Where("team_members.user_id = ?", user.ID).
		Or("namespaces.owner_id = ?", user.ID).
		And("namespaces.id = ?", n.ID).
		GroupBy("namespaces.id").
		Get(&all)

	return exists
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
