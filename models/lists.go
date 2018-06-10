package models

// List represents a list of items
type List struct {
	ID          int64  `xorm:"int(11) autoincr not null unique pk" json:"id"`
	Title       string `xorm:"varchar(250)" json:"title"`
	Description string `xorm:"varchar(1000)" json:"description"`
	OwnerID     int64  `xorm:"int(11)" json:"ownerID"`
	Owner       User   `xorm:"-" json:"owner"`
	Created     int64  `xorm:"created" json:"created"`
	Updated     int64  `xorm:"updated" json:"updated"`

	Items []*ListItem `xorm:"-"`
}

// GetListByID returns a list by its ID
func GetListByID(id int64) (list List, err error) {
	list.ID = id
	exists, err := x.Get(&list)
	if err != nil {
		return List{}, err
	}

	if !exists {
		return List{}, ErrListDoesNotExist{ID: id}
	}

	// Get the list owner
	user, _, err := GetUserByID(list.OwnerID)
	if err != nil {
		return List{}, err
	}

	list.Owner = user

	items, err := GetItemsByListID(list.ID)
	if err != nil {
		return
	}
	list.Items = items

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
