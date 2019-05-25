//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/web"
	"sort"
)

// ListTask represents an task in a todolist
type ListTask struct {
	// The unique, numeric id of this task.
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"id" param:"listtask"`
	// The task text. This is what you'll see in the list.
	Text string `xorm:"varchar(250) not null" json:"text" valid:"runelength(3|250)" minLength:"3" maxLength:"250"`
	// The task description.
	Description string `xorm:"varchar(250)" json:"description" valid:"runelength(0|250)" maxLength:"250"`
	// Whether a task is done or not.
	Done bool `xorm:"INDEX null" json:"done"`
	// The unix timestamp when a task was marked as done.
	DoneAtUnix int64 `xorm:"INDEX null" json:"doneAt"`
	// A unix timestamp when the task is due.
	DueDateUnix int64 `xorm:"int(11) INDEX null" json:"dueDate"`
	// An array of unix timestamps when the user wants to be reminded of the task.
	RemindersUnix []int64 `xorm:"-" json:"reminderDates"`
	CreatedByID   int64   `xorm:"int(11) not null" json:"-"` // ID of the user who put that task on the list
	// The list this task belongs to.
	ListID int64 `xorm:"int(11) INDEX not null" json:"listID" param:"list"`
	// An amount in seconds this task repeats itself. If this is set, when marking the task as done, it will mark itself as "undone" and then increase all remindes and the due date by its amount.
	RepeatAfter int64 `xorm:"int(11) INDEX null" json:"repeatAfter"`
	// If the task is a subtask, this is the id of its parent.
	ParentTaskID int64 `xorm:"int(11) INDEX null" json:"parentTaskID"`
	// The task priority. Can be anything you want, it is possible to sort by this later.
	Priority int64 `xorm:"int(11) null" json:"priority"`
	// When this task starts.
	StartDateUnix int64 `xorm:"int(11) INDEX null" json:"startDate" query:"-"`
	// When this task ends.
	EndDateUnix int64 `xorm:"int(11) INDEX null" json:"endDate" query:"-"`
	// An array of users who are assigned to this task
	Assignees []*User `xorm:"-" json:"assignees"`
	// An array of labels which are associated with this task.
	Labels []*Label `xorm:"-" json:"labels"`
	// The task color in hex
	HexColor string `xorm:"varchar(6) null" json:"hexColor" valid:"runelength(0|6)" maxLength:"6"`

	// The UID is currently not used for anything other than caldav, which is why we don't expose it over json
	UID string `xorm:"varchar(250) null" json:"-"`

	Sorting           string `xorm:"-" json:"-" query:"sort"` // Parameter to sort by
	StartDateSortUnix int64  `xorm:"-" json:"-" query:"startdate"`
	EndDateSortUnix   int64  `xorm:"-" json:"-" query:"enddate"`

	// An array of subtasks.
	Subtasks []*ListTask `xorm:"-" json:"subtasks"`

	// A unix timestamp when this task was created. You cannot change this value.
	Created int64 `xorm:"created not null" json:"created"`
	// A unix timestamp when this task was last updated. You cannot change this value.
	Updated int64 `xorm:"updated not null" json:"updated"`

	// The user who initially created the task.
	CreatedBy User `xorm:"-" json:"createdBy" valid:"-"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName returns the table name for listtasks
func (ListTask) TableName() string {
	return "tasks"
}

// TaskReminder holds a reminder on a task
type TaskReminder struct {
	ID           int64 `xorm:"int(11) autoincr not null unique pk"`
	TaskID       int64 `xorm:"int(11) not null INDEX"`
	ReminderUnix int64 `xorm:"int(11) not null INDEX"`
	Created      int64 `xorm:"created not null"`
}

// TableName returns a pretty table name
func (TaskReminder) TableName() string {
	return "task_reminders"
}

// GetTasksByListID gets all todotasks for a list
func GetTasksByListID(listID int64) (tasks []*ListTask, err error) {
	// make a map so we can put in a lot of other stuff more easily
	taskMap := make(map[int64]*ListTask, len(tasks))
	err = x.Where("list_id = ?", listID).Find(&taskMap)
	if err != nil {
		return
	}

	tasks, err = addMoreInfoToTasks(taskMap)
	return
}

// GetTaskByIDSimple returns a raw task without extra data by the task ID
func GetTaskByIDSimple(taskID int64) (task ListTask, err error) {
	if taskID < 1 {
		return ListTask{}, ErrListTaskDoesNotExist{taskID}
	}

	return GetTaskSimple(&ListTask{ID: taskID})
}

// GetTaskSimple returns a raw task without extra data
func GetTaskSimple(t *ListTask) (task ListTask, err error) {
	task = *t
	exists, err := x.Get(&task)
	if err != nil {
		return ListTask{}, err
	}

	if !exists {
		return ListTask{}, ErrListTaskDoesNotExist{t.ID}
	}
	return
}

// GetTaskByID returns all tasks a list has
func GetTaskByID(listTaskID int64) (listTask ListTask, err error) {
	listTask, err = GetTaskByIDSimple(listTaskID)
	if err != nil {
		return
	}

	u, err := GetUserByID(listTask.CreatedByID)
	if err != nil {
		return
	}
	listTask.CreatedBy = u

	// Get assignees
	taskAssignees, err := getRawTaskAssigneesForTasks([]int64{listTaskID})
	if err != nil {
		return
	}
	for _, u := range taskAssignees {
		if u != nil {
			listTask.Assignees = append(listTask.Assignees, &u.User)
		}
	}

	// Get task labels
	taskLabels, err := getLabelsByTaskIDs(&LabelByTaskIDsOptions{
		TaskIDs: []int64{listTaskID},
	})
	if err != nil {
		return
	}
	for _, label := range taskLabels {
		listTask.Labels = append(listTask.Labels, &label.Label)
	}

	return
}

// GetTasksByIDs returns all tasks for a list of ids
func (bt *BulkTask) GetTasksByIDs() (err error) {
	for _, id := range bt.IDs {
		if id < 1 {
			return ErrListTaskDoesNotExist{id}
		}
	}

	taskMap := make(map[int64]*ListTask, len(bt.Tasks))
	err = x.In("id", bt.IDs).Find(&taskMap)
	if err != nil {
		return
	}

	bt.Tasks, err = addMoreInfoToTasks(taskMap)
	return
}

// GetTasksByUIDs gets all tasks from a bunch of uids
func GetTasksByUIDs(uids []string) (tasks []*ListTask, err error) {
	taskMap := make(map[int64]*ListTask)
	err = x.In("uid", uids).Find(&taskMap)
	if err != nil {
		return
	}

	tasks, err = addMoreInfoToTasks(taskMap)
	return
}

// This function takes a map with pointers and returns a slice with pointers to tasks
// It adds more stuff like assignees/labels/etc to a bunch of tasks
func addMoreInfoToTasks(taskMap map[int64]*ListTask) (tasks []*ListTask, err error) {

	// No need to iterate over users and stuff if the list doesn't has tasks
	if len(taskMap) == 0 {
		return
	}

	// Get all users & task ids and put them into the array
	var userIDs []int64
	var taskIDs []int64
	for _, i := range taskMap {
		taskIDs = append(taskIDs, i.ID)
		userIDs = append(userIDs, i.CreatedByID)
	}

	// Get all assignees
	taskAssignees, err := getRawTaskAssigneesForTasks(taskIDs)
	if err != nil {
		return
	}
	// Put the assignees in the task map
	for _, a := range taskAssignees {
		if a != nil {
			taskMap[a.TaskID].Assignees = append(taskMap[a.TaskID].Assignees, &a.User)
		}
	}

	// Get all labels for all the tasks
	labels, err := getLabelsByTaskIDs(&LabelByTaskIDsOptions{TaskIDs: taskIDs})
	if err != nil {
		return
	}
	for _, l := range labels {
		if l != nil {
			taskMap[l.TaskID].Labels = append(taskMap[l.TaskID].Labels, &l.Label)
		}
	}

	// Get all users of a task
	// aka the ones who created a task
	users := make(map[int64]*User)
	err = x.In("id", userIDs).Find(&users)
	if err != nil {
		return
	}

	// Get all reminders and put them in a map to have it easier later
	reminders := []*TaskReminder{}
	err = x.Table("task_reminders").In("task_id", taskIDs).Find(&reminders)
	if err != nil {
		return
	}

	taskRemindersUnix := make(map[int64][]int64)
	for _, r := range reminders {
		taskRemindersUnix[r.TaskID] = append(taskRemindersUnix[r.TaskID], r.ReminderUnix)
	}

	// Add all user objects to the appropriate tasks
	for _, task := range taskMap {

		// Make created by user objects
		taskMap[task.ID].CreatedBy = *users[task.CreatedByID]

		// Add the reminders
		taskMap[task.ID].RemindersUnix = taskRemindersUnix[task.ID]

		// Reorder all subtasks
		if task.ParentTaskID != 0 {
			taskMap[task.ParentTaskID].Subtasks = append(taskMap[task.ParentTaskID].Subtasks, task)
			delete(taskMap, task.ID)
		}
	}

	// make a complete slice from the map
	tasks = []*ListTask{}
	for _, t := range taskMap {
		tasks = append(tasks, t)
	}

	// Sort the output. In Go, contents on a map are put on that map in no particular order.
	// Because of this, tasks are not sorted anymore in the output, this leads to confiusion.
	// To avoid all this, we need to sort the slice afterwards
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})

	return
}
