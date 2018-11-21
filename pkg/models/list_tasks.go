package models

// ListTask represents an task in a todolist
type ListTask struct {
	ID            int64   `xorm:"int(11) autoincr not null unique pk" json:"id" param:"listtask"`
	Text          string  `xorm:"varchar(250)" json:"text" valid:"runelength(5|250)"`
	Description   string  `xorm:"varchar(250)" json:"description" valid:"runelength(0|250)"`
	Done          bool    `xorm:"INDEX" json:"done"`
	DueDateUnix   int64   `xorm:"int(11) INDEX" json:"dueDate"`
	RemindersUnix []int64 `xorm:"JSON TEXT" json:"reminderDates"`
	CreatedByID   int64   `xorm:"int(11)" json:"-"` // ID of the user who put that task on the list
	ListID        int64   `xorm:"int(11) INDEX" json:"listID" param:"list"`

	Created int64 `xorm:"created" json:"created" valid:"range(0|0)"`
	Updated int64 `xorm:"updated" json:"updated" valid:"range(0|0)"`

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
		for _, u := range users {
			if task.CreatedByID == u.ID {
				tasks[in].CreatedBy = u
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
	if listTaskID < 1 {
		return ListTask{}, ErrListTaskDoesNotExist{listTaskID}
	}

	exists, err := x.ID(listTaskID).Get(&listTask)
	if err != nil {
		return ListTask{}, err
	}

	if !exists {
		return ListTask{}, ErrListTaskDoesNotExist{listTaskID}
	}

	u, err := GetUserByID(listTask.CreatedByID)
	if err != nil {
		return
	}
	listTask.CreatedBy = u

	return
}
