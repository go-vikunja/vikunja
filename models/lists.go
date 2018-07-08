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
}

func (l *List) AfterLoad() {

	// Get the owner
	l.Owner, _, _ = GetUserByID(l.OwnerID)

	// Get the list items
	l.Items, _ = GetItemsByListID(l.ID)
}

// GetListByID returns a list by its ID
func GetListByID(id int64) (list List, err error) {
	exists, err := x.ID(id).Get(&list) // tName ist hässlich, geht das nicht auch anders?
	if err != nil {
		return list, err
	}

	if !exists {
		return list, ErrListDoesNotExist{ID: id}
	}

	return list, nil
}

// GetListsByUser gets all lists a user owns
func GetListsByUser(user *User) (lists []*List, err error) {
	fullUser, _, err := GetUserByID(user.ID)
	if err != nil {
		return
	}

	err = x.Where("owner_id = ?", user.ID).Find(&lists)
	if err != nil {
		return
	}

	for in := range lists {
		lists[in].Owner = fullUser
	}

	return
}

func GetListsByNamespaceID(nID int64) (lists []*List, err error) {
	err = x.Where("namespace_id = ?", nID).Find(&lists)
	return lists, err
}

func (list *List) ReadAll(user *User) (interface{}, error) {
	lists := Lists{}
	err := lists.ReadAll(user)
	return lists, err
}

type Lists []List

func (lists *Lists) ReadAll(user *User) (err error) {
	fullUser, _, err := GetUserByID(user.ID)
	if err != nil {
		return
	}

	err = x.Select("list.*").
		Join("LEFT", "team_list", "list.id = team_list.list_id").
		Join("LEFT", "team_members", "team_members.team_id = team_list.team_id").
		Where("team_members.user_id = ?", fullUser.ID).
		Or("list.owner_id = ?", fullUser.ID).
		Find(lists)
	if err != nil {
		return
	}

	return
}

func (l *List) ReadOne(id int64) (err error) {
	*l, err = GetListByID(id)
	return
}
