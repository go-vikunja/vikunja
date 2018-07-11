package models

// List represents a list of items
type List struct {
	ID          int64  `xorm:"int(11) autoincr not null unique pk" json:"id"`
	Title       string `xorm:"varchar(250)" json:"title"`
	Description string `xorm:"varchar(1000)" json:"description"`
	OwnerID     int64  `xorm:"int(11)" json:"-"`
	NamespaceID int64  `xorm:"int(11)" json:"-"`

	Owner User        `xorm:"-" json:"owner"`
	Items []*ListItem `xorm:"-" json:"items"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	CRUDable `xorm:"-" json:"-"`
	Rights   `xorm:"-" json:"-"`
}

// Lists is a multiple of list
type Lists []List

// AfterLoad loads the owner and list items
func (l *List) AfterLoad() {

	// Get the owner
	l.Owner, _, _ = GetUserByID(l.OwnerID)

	// Get the list items
	l.Items, _ = GetItemsByListID(l.ID)
}

// GetListByID returns a list by its ID
func GetListByID(id int64) (list List, err error) {
	exists, err := x.ID(id).Get(&list)
	if err != nil {
		return list, err
	}

	if !exists {
		return list, ErrListDoesNotExist{ID: id}
	}

	return list, nil
}

// GetListsByNamespaceID gets all lists in a namespace
func GetListsByNamespaceID(nID int64) (lists []*List, err error) {
	err = x.Where("namespace_id = ?", nID).Find(&lists)
	return lists, err
}

// ReadAll gets all List a user has access to
func (l *List) ReadAll(user *User) (interface{}, error) {
	lists := Lists{}
	fullUser, _, err := GetUserByID(user.ID)
	if err != nil {
		return lists, err
	}

	// TODO: namespaces...
	err = x.Select("list.*").
		Join("LEFT", "team_list", "list.id = team_list.list_id").
		Join("LEFT", "team_members", "team_members.team_id = team_list.team_id").
		Where("team_members.user_id = ?", fullUser.ID).
		Or("list.owner_id = ?", fullUser.ID).
		Find(&lists)

	return lists, err
}

// ReadOne gets one list by its ID
func (l *List) ReadOne(id int64) (err error) {
	*l, err = GetListByID(id)
	return
}

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
	// Admins always have write access
	if l.IsAdmin(user) {
		return true
	}

	// Owners always have write access
	if l.Owner.ID == user.ID {
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