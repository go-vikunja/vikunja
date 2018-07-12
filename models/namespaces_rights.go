package models

// NamespaceRight defines the rights teams can have for namespaces
type NamespaceRight int

// define unknown namespace right
const (
	NamespaceRightUnknown = -1
)

// Enumerate all the namespace rights
const (
	// Can read lists in a namespace
	NamespaceRightRead NamespaceRight = iota
	// Cat write items in a namespace like lists and todo items
	NamespaceRightWrite
	// Can manage a namespace, can do everything
	NamespaceRightAdmin
)

// IsNamespaceAdmin returns whether the usre has admin rights in a namespace
func (user *User) IsNamespaceAdmin(namespace *Namespace) (err error) {
	// Owners always have admin rights
	if user.ID == namespace.Owner.ID {
		return nil
	}

	// Check if that user is in a team which has admin rights to that namespace

	return ErrUserNeedsToBeNamespaceAdmin{UserID: user.ID, NamespaceID: namespace.ID}
}

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
func (n *Namespace) CanUpdate(user *User, id int64) bool {
	nn, _ := GetNamespaceByID(id)
	return nn.IsAdmin(user)
}

// CanCreate checks if the user can create a new namespace
func (n *Namespace) CanCreate(user *User, id int64) bool {
	// This is currently a dummy function, later on we could imagine global limits etc.
	return true
}
