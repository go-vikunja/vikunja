package models

// ListTask represents an task in a todolist
type ListTask struct {
	ID           int64  `xorm:"int(11) autoincr not null unique pk" json:"id" param:"listtask"`
	Text         string `xorm:"varchar(250)" json:"text"`
	Description  string `xorm:"varchar(250)" json:"description"`
	Done         bool   `json:"done"`
	DueDateUnix  int64  `xorm:"int(11)" json:"dueDate"`
	ReminderUnix int64  `xorm:"int(11)" json:"reminderDate"`
	CreatedByID  int64  `xorm:"int(11)" json:"-"` // ID of the user who put that task on the list
	ListID       int64  `xorm:"int(11)" json:"listID" param:"list"`
	Created      int64  `xorm:"created" json:"created"`
	Updated      int64  `xorm:"updated" json:"updated"`

	CreatedBy User `xorm:"-" json:"createdBy"`

	CRUDable `xorm:"-" json:"-"`
	Rights   `xorm:"-" json:"-"`
}

// TableName returns the table name for listtasks
func (ListTask) TableName() string {
	return "tasks"
}

// GetTasksByListID gets all todotasks for a list
func GetTasksByListID(listID int64) (tasks []*ListTask, err error) {
	err = x.Where("list_id = ?", listID).Find(&tasks)
	if err != nil {
		return
	}

	// No need to iterate over users if the list doesn't has tasks
	if len(tasks) == 0 {
		return
	}

	// Get all users and put them into the array
	var userIDs []int64
	for _, i := range tasks {
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

	for in, task := range tasks {
		for _, user := range users {
			if task.CreatedByID == user.ID {
				tasks[in].CreatedBy = user
				break
			}
		}

		// obsfucate the user password
		tasks[in].CreatedBy.Password = ""
	}

	return
}

// GetListTaskByID returns all tasks a list has
func GetListTaskByID(listTaskID int64) (listTask ListTask, err error) {
	exists, err := x.ID(listTaskID).Get(&listTask)
	if err != nil {
		return ListTask{}, err
	}

	if !exists {
		return ListTask{}, ErrListTaskDoesNotExist{listTaskID}
	}

	user, err := GetUserByID(listTask.CreatedByID)
	if err != nil {
		return
	}
	listTask.CreatedBy = user

	return
}
