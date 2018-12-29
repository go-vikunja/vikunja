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
	ID            int64   `xorm:"int(11) autoincr not null unique pk" json:"id" param:"listtask"`
	Text          string  `xorm:"varchar(250)" json:"text" valid:"runelength(3|250)"`
	Description   string  `xorm:"varchar(250)" json:"description" valid:"runelength(0|250)"`
	Done          bool    `xorm:"INDEX" json:"done"`
	DueDateUnix   int64   `xorm:"int(11) INDEX" json:"dueDate"`
	RemindersUnix []int64 `xorm:"JSON TEXT" json:"reminderDates"`
	CreatedByID   int64   `xorm:"int(11)" json:"-"` // ID of the user who put that task on the list
	ListID        int64   `xorm:"int(11) INDEX" json:"listID" param:"list"`
	RepeatAfter   int64   `xorm:"int(11) INDEX" json:"repeatAfter"`
	ParentTaskID  int64   `xorm:"int(11) INDEX" json:"parentTaskID"`
	Priority      int64   `xorm:"int(11)" json:"priority"`
	StartDateUnix int64   `xorm:"int(11) INDEX" json:"startDate"`
	EndDateUnix   int64   `xorm:"int(11) INDEX" json:"endDate"`
	Assignees     []*User `xorm:"-" json:"assignees"`

	Sorting           string `xorm:"-" json:"-" param:"sort"` // Parameter to sort by
	StartDateSortUnix int64  `xorm:"-" json:"-" param:"startdatefilter"`
	EndDateSortUnix   int64  `xorm:"-" json:"-" param:"enddatefilter"`

	Subtasks []*ListTask `xorm:"-" json:"subtasks"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	CreatedBy User `xorm:"-" json:"createdBy" valid:"-"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName returns the table name for listtasks
func (ListTask) TableName() string {
	return "tasks"
}

// ListTaskAssginee represents an assignment of a user to a task
type ListTaskAssginee struct {
	ID      int64 `xorm:"int(11) autoincr not null unique pk"`
	TaskID  int64 `xorm:"int(11) not null"`
	UserID  int64 `xorm:"int(11) not null"`
	Created int64 `xorm:"created"`
}

// TableName makes a pretty table name
func (ListTaskAssginee) TableName() string {
	return "task_assignees"
}

// ListTaskAssigneeWithUser is a helper type to deal with user joins
type ListTaskAssigneeWithUser struct {
	TaskID int64
	User   `xorm:"extends"`
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

	// make a map so we can put in subtasks more easily
	taskMap := make(map[int64]*ListTask, len(tasks))

	// Get all users & task ids and put them into the array
	var userIDs []int64
	var taskIDs []int64
	for _, i := range tasks {
		taskIDs = append(taskIDs, i.ID)
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

		taskMap[i.ID] = i
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

	var users []User
	err = x.In("id", userIDs).Find(&users)
	if err != nil {
		return
	}

	// Add all user objects to the appropriate tasks
	for _, task := range taskMap {

		// Make created by user objects
		for _, u := range users {
			if task.CreatedByID == u.ID {
				taskMap[task.ID].CreatedBy = u
				break
			}
		}

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

func getRawTaskAssigneesForTasks(taskIDs []int64) (taskAssignees []*ListTaskAssigneeWithUser, err error) {
	taskAssignees = []*ListTaskAssigneeWithUser{nil}
	err = x.Table("task_assignees").
		Select("task_id, users.*").
		In("task_id", taskIDs).
		Join("INNER", "users", "task_assignees.user_id = users.id").
		Find(&taskAssignees)
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

	return
}

// GetTasksByIDs returns all tasks for a list of ids
func (bt *BulkTask) GetTasksByIDs() (err error) {
	for _, id := range bt.IDs {
		if id < 1 {
			return ErrListTaskDoesNotExist{id}
		}
	}

	err = x.In("id", bt.IDs).Find(&bt.Tasks)
	if err != nil {
		return err
	}

	// We use a map, to avoid looping over two slices at once
	var usermapids = make(map[int64]bool) // Bool ist just something, doesn't acutually matter
	for _, list := range bt.Tasks {
		usermapids[list.CreatedByID] = true
	}

	// Make a slice from the map
	var userids []int64
	for uid := range usermapids {
		userids = append(userids, uid)
	}

	// Get all users for the tasks
	var users []*User
	err = x.In("id", userids).Find(&users)
	if err != nil {
		return err
	}

	for in, task := range bt.Tasks {
		for _, u := range users {
			if task.CreatedByID == u.ID {
				bt.Tasks[in].CreatedBy = *u
			}
		}
	}

	return
}
