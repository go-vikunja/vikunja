package models

// Namespace holds informations about a namespace
type Namespace struct {
	ID          int64  `xorm:"int(11) autoincr not null unique pk" json:"id"`
	Name        string `xorm:"varchar(250) autoincr not null" json:"name"`
	Description string `xorm:"varchar(700) autoincr not null" json:"description"`
	OwnerID     int64  `xorm:"int(11) autoincr not null" json:"owner_id"`

	Owner User `xorm:"-"`

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

func (user User) IsNamespaceAdmin(namespace Namespace) (ok bool, err error) {
	// Owners always have admin rights
	if user.ID == namespace.Owner.ID {
		return true, nil
	}

	// Check if that user is in a team which has admin rights to that namespace


	return
}

// CreateOrUpdateNamespace does what it says
func CreateOrUpdateNamespace(namespace *Namespace) (err error) {
	// Check if the User exists
	_, _, err = GetUserByID(namespace.Owner.ID)
	if err != nil {
		return
	}

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
