package models

// ListItem represents an item in a todolist
type ListItem struct {
	ID           int64  `xorm:"int(11) autoincr not null unique pk" json:"id"`
	Text         string `xorm:"varchar(250)" json:"text"`
	Description  string `xorm:"varchar(250)" json:"description"`
	Done         bool   `json:"done"`
	DueDateUnix  int64  `xorm:"int(11)" json:"dueDate"`
	ReminderUnix int64  `xorm:"int(11)" json:"reminderDate"`
	CreatedByID  int64  `xorm:"int(11)" json:"-"` // ID of the user who put that item on the list
	ListID       int64  `xorm:"int(11)" json:"listID"`
	Created      int64  `xorm:"created" json:"created"`
	Updated      int64  `xorm:"updated" json:"updated"`

	CreatedBy User `xorm:"-" json:"createdBy"`
}

// TableName returns the table name for listitems
func (ListItem) TableName() string {
	return "items"
}

// GetItemsByListID gets all todoitems for a list
func GetItemsByListID(listID int64) (items []*ListItem, err error) {
	err = x.Where("list_id = ?", listID).Find(&items)
	if err != nil {
		return
	}

	// Get all users and put them into the array
	var userIDs []int64
	for _, i := range items {
		found := false
		for _, u := range userIDs {
			if i.CreatedByID == u {
				found = true
				break
			}
		}

		if !found {
			userIDs = append(userIDs, i.CreatedByID)
		}
	}

	var users []User
	err = x.In("id", userIDs).Find(&users)
	if err != nil {
		return
	}

	for in, item := range items {
		for _, user := range users {
			if item.CreatedByID == user.ID {
				items[in].CreatedBy = user
				break
			}
		}

		// obsfucate the user password
		items[in].CreatedBy.Password = ""
	}

	return
}

func GetListItemByID(listItemID int64) (listItem ListItem, err error) {
	exists, err := x.ID(listItemID).Get(&listItem)
	if err != nil {
		return ListItem{}, err
	}

	if !exists {
		return ListItem{}, ErrListItemDoesNotExist{listItemID}
	}

	user, _, err := GetUserByID(listItem.CreatedByID)
	if err != nil {
		return
	}
	listItem.CreatedBy = user

	return
}

// DeleteListItemByID deletes a list item by its ID
func DeleteListItemByID(itemID int64, doer *User) (err error) {

	// Check if it exists
	listitem, err := GetListItemByID(itemID)
	if err != nil {
		return
	}

	// Check if the user hat the right to delete that item
	if listitem.CreatedByID != doer.ID {
		return ErrNeedToBeItemOwner{ItemID: itemID, UserID: doer.ID}
	}

	_, err = x.ID(itemID).Delete(ListItem{})
	return
}
