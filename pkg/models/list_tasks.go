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
	Text string `xorm:"varchar(250)" json:"text" valid:"runelength(3|250)" minLength:"3" maxLength:"250"`
	// The task description.
	Description string `xorm:"varchar(250)" json:"description" valid:"runelength(0|250)" maxLength:"250"`
	Done        bool   `xorm:"INDEX" json:"done"`
	// A unix timestamp when the task is due.
	DueDateUnix int64 `xorm:"int(11) INDEX" json:"dueDate"`
	// An array of unix timestamps when the user wants to be reminded of the task.
	RemindersUnix []int64 `xorm:"JSON TEXT" json:"reminderDates"`
	CreatedByID   int64   `xorm:"int(11)" json:"-"` // ID of the user who put that task on the list
	// The list this task belongs to.
	ListID int64 `xorm:"int(11) INDEX" json:"listID" param:"list"`
	// An amount in seconds this task repeats itself. If this is set, when marking the task as done, it will mark itself as "undone" and then increase all remindes and the due date by its amount.
	RepeatAfter int64 `xorm:"int(11) INDEX" json:"repeatAfter"`
	// If the task is a subtask, this is the id of its parent.
	ParentTaskID int64 `xorm:"int(11) INDEX" json:"parentTaskID"`
	// The task priority. Can be anything you want, it is possible to sort by this later.
	Priority int64 `xorm:"int(11)" json:"priority"`
	// When this task starts.
	StartDateUnix int64 `xorm:"int(11) INDEX" json:"startDate"`
	// When this task ends.
	EndDateUnix int64 `xorm:"int(11) INDEX" json:"endDate"`
	// An array of users who are assigned to this task
	Assignees []*User `xorm:"-" json:"assignees"`
	// An array of labels which are associated with this task.
	Labels []*Label `xorm:"-" json:"labels"`

	Sorting           string `xorm:"-" json:"-" param:"sort"` // Parameter to sort by
	StartDateSortUnix int64  `xorm:"-" json:"-" param:"startdatefilter"`
	EndDateSortUnix   int64  `xorm:"-" json:"-" param:"enddatefilter"`

	// An array of subtasks.
	Subtasks []*ListTask `xorm:"-" json:"subtasks"`

	// A unix timestamp when this task was created. You cannot change this value.
	Created int64 `xorm:"created" json:"created"`
	// A unix timestamp when this task was last updated. You cannot change this value.
	Updated int64 `xorm:"updated" json:"updated"`

	// The user who initially created the task.
	CreatedBy User `xorm:"-" json:"createdBy" valid:"-"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName returns the table name for listtasks
func (ListTask) TableName() string {
	return "tasks"
}

// GetTasksByListID gets all todotasks for a list
func GetTasksByListID(listID int64) (tasks []*ListTask, err error) {
	// make a map so we can put in a lot of other stuff more easily
	taskMap := make(map[int64]*ListTask, len(tasks))
	err = x.Where("list_id = ?", listID).Find(&taskMap)
	if err != nil {
		return
	}

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

	// Get all labels for the tasks
	labels, err := getLabelsByTaskIDs("", &User{}, -1, taskIDs, false)
	if err != nil {
		return
	}
	for _, l := range labels {
		if l != nil {
			taskMap[l.TaskID].Labels = append(taskMap[l.TaskID].Labels, &l.Label)
		}
	}

	users := make(map[int64]*User)
	err = x.In("id", userIDs).Find(&users)
	if err != nil {
		return
	}

	// Add all user objects to the appropriate tasks
	for _, task := range taskMap {

		// Make created by user objects
		taskMap[task.ID].CreatedBy = *users[task.CreatedByID]

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

	// Sort the output. In Go, contents on a map are put on that map in no particular order (saved on heap).
	// Because of this, tasks are not sorted anymore in the output, this leads to confiusion.
	// To avoid all this, we need to sort the slice afterwards
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})

	return
}

func getTaskByIDSimple(taskID int64) (task ListTask, err error) {
	if taskID < 1 {
		return ListTask{}, ErrListTaskDoesNotExist{taskID}
	}

	exists, err := x.ID(taskID).Get(&task)
	if err != nil {
		return ListTask{}, err
	}

	if !exists {
		return ListTask{}, ErrListTaskDoesNotExist{taskID}
	}
	return
}

// GetListTaskByID returns all tasks a list has
func GetListTaskByID(listTaskID int64) (listTask ListTask, err error) {
	listTask, err = getTaskByIDSimple(listTaskID)
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
