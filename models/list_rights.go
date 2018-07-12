package models

// IsAdmin returns whether the user has admin rights on the list or not
func (l *List) IsAdmin(user *User) bool {
	// Owners are always admins
	if l.Owner.ID == user.ID {
		return true
	}

	// Check Team rights
	// aka "is the user in a team which has admin rights?"
	// TODO

	// Check Namespace rights
	// TODO

	// Check individual rights
	// TODO

	return false
}

// CanWrite return whether the user can write on that list or not
func (l *List) CanWrite(user *User) bool {
	// Owners always have write access
	if l.Owner.ID == user.ID {
		return true
	}

	// Admins always have write access
	if l.IsAdmin(user) {
		return true
	}

	// Check Namespace rights
	// TODO
	// TODO find a way to prioritize: what happens if a user has namespace write access but is not in that list?

	// Check Team rights
	// TODO

	// Check individual rights
	// TODO

	return false
}

// CanRead checks if a user has read access to a list
func (l *List) CanRead(user *User) bool {
	// Owners always have read access
	if l.Owner.ID == user.ID {
		return true
	}

	// Admins always have read access
	if l.IsAdmin(user) {
		return true
	}

	// Check Namespace rights
	exists, _ := x.Select("list.*").
		Table("namespaces").
		Join("INNER", "list", "list.namespace_id = namespaces.id").
		Join("INNER", "team_namespaces", "team_namespaces.namespace_id = namespaces.id").
		Join("INNER", "team_members", "team_members.team_id = team_namespaces.team_id").
		Where("team_members.user_id = ?", user.ID).
		And("list.id = ?", l.ID).
		Get(&List{})

	if exists {
		return true
	}

	// Check Team rights
	// TODO

	// Check individual rights
	// TODO

	return false
}
