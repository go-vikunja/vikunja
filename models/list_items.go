package models

// ListItem represents an item in a todolist
type ListItem struct {
	ID           int64  `xorm:"int(11) autoincr not null unique pk" json:"id"`
	Text         string `xorm:"varchar(250)" json:"text"`
	Description  string `xorm:"varchar(250)" json:"description"`
	Done         bool   `json:"done"`
	DueDateUnix  int64  `xorm:"int(11)" json:"dueDate"`
	ReminderUnix int64  `xorm:"int(11)" json:"reminderDate"`
	CreatedByID  int64  `xorm:"int(11)" json:"createdByID"` // ID of the user who put that item on the list
	ListID       int64  `xorm:"int(11)" json:"listID"`
	Created      int64  `xorm:"created" json:"created"`
	Updated      int64  `xorm:"updated" json:"updated"`

	CreatedBy User `xorm:"-"`
}

// TableName returns the table name for listitems
func (ListItem) TableName() string {
	return "items"
}

func GetItemsByListID(listID int64) (items []*ListItem, err error) {
	err = x.Where("list_id = ?", listID).Find(&items)
	return
}
