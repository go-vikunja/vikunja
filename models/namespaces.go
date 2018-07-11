package models

// Namespace holds informations about a namespace
type Namespace struct {
	ID          int64  `xorm:"int(11) autoincr not null unique pk" json:"id"`
	Name        string `xorm:"varchar(250)" json:"name"`
	Description string `xorm:"varchar(1000)" json:"description"`
	OwnerID     int64  `xorm:"int(11) not null" json:"-"`

	Owner User `xorm:"-" json:"owner"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	CRUDable `xorm:"-" json:"-"`
	Rights   `xorm:"-" json:"-"`
}

// TableName makes beautiful table names
func (Namespace) TableName() string {
	return "namespaces"
}

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

// HasNamespaceAccess checks if the User has namespace read access
func (user *User) HasNamespaceAccess(namespace *Namespace) (err error) {
	// Owners always have access
	if user.ID == namespace.Owner.ID {
		return nil
	}

	// Check if the user is in a team which has access to the namespace

	return ErrUserDoesNotHaveAccessToNamespace{UserID: user.ID, NamespaceID: namespace.ID}
}

// CanWrite checks if a user has write access to a namespace
func (n *Namespace) CanWrite(user *User) bool {
	// Owners always have access
	if user.ID == n.Owner.ID {
		return true
	}

	return true
}

// GetNamespaceByID returns a namespace object by its ID
func GetNamespaceByID(id int64) (namespace Namespace, err error) {
	namespace.ID = id
	exists, err := x.Get(&namespace)
	if err != nil {
		return namespace, err
	}

	if !exists {
		return namespace, ErrNamespaceDoesNotExist{ID: id}
	}

	// Get the namespace Owner
	namespace.Owner, _, err = GetUserByID(namespace.OwnerID)
	if err != nil {
		return namespace, err
	}

	return namespace, err
}

// ReadAll gets all namespaces a user has access to
func (n *Namespace) ReadAll(doer *User) (interface{}, error) {

	all := []*Namespace{}

	// TODO respect individual rights
	err := x.Select("namespaces.*").
		Table("namespaces").
		Join("LEFT", "team_namespaces", "namespaces.id = team_namespaces.namespace_id").
		Join("LEFT", "team_members", "team_members.team_id = team_namespaces.team_id").
		Where("team_members.user_id = ?", doer.ID).
		Or("namespaces.owner_id = ?", doer.ID).
		GroupBy("namespaces.id").
		Find(&all)

	if err != nil {
		return all, err
	}

	// Get all users
	users := []*User{}
	err = x.Select("users.*").
		Table("namespaces").
		Join("LEFT", "team_namespaces", "namespaces.id = team_namespaces.namespace_id").
		Join("LEFT", "team_members", "team_members.team_id = team_namespaces.team_id").
		Join("INNER", "users", "users.id = namespaces.owner_id").
		Where("team_members.user_id = ?", doer.ID).
		Or("namespaces.owner_id = ?", doer.ID).
		GroupBy("users.id").
		Find(&users)

	if err != nil {
		return all, err
	}

	// Put user objects in our namespace list
	for i, n := range all {
		for _, u := range users {
			if n.OwnerID == u.ID {
				all[i].Owner = *u
				break
			}
		}
	}

	return all, nil
}
