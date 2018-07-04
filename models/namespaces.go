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

func (user *User) IsNamespaceAdmin(namespace *Namespace) (ok bool, err error) {
	// Owners always have admin rights
	if user.ID == namespace.Owner.ID {
		return true, nil
	}

	// Check if that user is in a team which has admin rights to that namespace

	return
}

func (user *User) HasNamespaceAccess(namespace *Namespace) (has bool, err error) {
	// Owners always have access
	if user.ID == namespace.Owner.ID {
		return true, nil
	}

	// Check if the user is in a team which has access to the namespace

	return
}

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
