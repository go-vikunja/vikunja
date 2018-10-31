package models

import (
	"code.vikunja.io/api/pkg/log"
)

// IsAdmin returns true or false if the user is admin on that namespace or not
func (n *Namespace) IsAdmin(u *User) bool {
	// Owners always have admin rights
	if u.ID == n.Owner.ID {
		return true
	}

	// Check user rights
	if n.checkUserRights(u, UserRightAdmin) {
		return true
	}

	// Check if that user is in a team which has admin rights to that namespace
	return n.checkTeamRights(u, TeamRightAdmin)
}

// CanWrite checks if a user has write access to a namespace
func (n *Namespace) CanWrite(u *User) bool {
	// Admins always have write access
	if n.IsAdmin(u) {
		return true
	}

	// Check user rights
	if n.checkUserRights(u, UserRightWrite) {
		return true
	}

	// Check if that user is in a team which has write rights to that namespace
	return n.checkTeamRights(u, TeamRightWrite)
}

// CanRead checks if a user has read access to that namespace
func (n *Namespace) CanRead(u *User) bool {
	// Admins always have read access
	if n.IsAdmin(u) {
		return true
	}

	// Check user rights
	if n.checkUserRights(u, UserRightRead) {
		return true
	}

	// Check if the user is in a team which has access to the namespace
	return n.checkTeamRights(u, TeamRightRead)
}

// CanUpdate checks if the user can update the namespace
func (n *Namespace) CanUpdate(u *User) bool {
	nn, err := GetNamespaceByID(n.ID)
	if err != nil {
		log.Log.Error("Error occurred during CanUpdate for Namespace: %s", err)
		return false
	}
	return nn.IsAdmin(u)
}

// CanDelete checks if the user can delete a namespace
func (n *Namespace) CanDelete(u *User) bool {
	nn, err := GetNamespaceByID(n.ID)
	if err != nil {
		log.Log.Error("Error occurred during CanDelete for Namespace: %s", err)
		return false
	}
	return nn.IsAdmin(u)
}

// CanCreate checks if the user can create a new namespace
func (n *Namespace) CanCreate(u *User) bool {
	// This is currently a dummy function, later on we could imagine global limits etc.
	return true
}

func (n *Namespace) checkTeamRights(u *User, r TeamRight) bool {
	exists, err := x.Select("namespaces.*").
		Table("namespaces").
		Join("LEFT", "team_namespaces", "namespaces.id = team_namespaces.namespace_id").
		Join("LEFT", "team_members", "team_members.team_id = team_namespaces.team_id").
		Where("namespaces.id = ? "+
			"AND (team_members.user_id = ?  AND team_namespaces.right = ?) "+
			"OR namespaces.owner_id = ? ", n.ID, u.ID, r, u.ID).
		Get(&Namespace{})
	if err != nil {
		log.Log.Error("Error occurred during checkTeamRights for Namespace: %s, TeamRight: %d", err, r)
		return false
	}

	return exists
}

func (n *Namespace) checkUserRights(u *User, r UserRight) bool {
	exists, err := x.Select("namespaces.*").
		Table("namespaces").
		Join("LEFT", "users_namespace", "users_namespace.namespace_id = namespaces.id").
		Where("namespaces.id = ? AND ("+
			"namespaces.owner_id = ? "+
			"OR (users_namespace.user_id = ? AND users_namespace.right = ?))", n.ID, u.ID, u.ID, r).
		Get(&Namespace{})
	if err != nil {
		log.Log.Error("Error occurred during checkUserRights for Namespace: %s, UserRight: %d", err, r)
		return false
	}

	return exists
}
