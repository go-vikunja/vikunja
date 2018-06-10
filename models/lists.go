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

	return list, nil
}

//
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

// CreateOrUpdateList updates a list or creates it if it doesn't exist
func CreateOrUpdateList(list *List) (err error) {
	// Check if it exists
	_, err = GetListByID(list.ID)
	if err != nil {
		return
	}

	list.OwnerID = list.Owner.ID

	if list.ID == 0 {
		_, err = x.Insert(list)
	} else {
		_, err = x.ID(list.ID).Update(list)
		return
	}

	return

}
