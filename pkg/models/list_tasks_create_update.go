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
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/web"
	"github.com/imdario/mergo"
)

// Create is the implementation to create a list task
// @Summary Create a task
// @Description Inserts a task into a list.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "List ID"
// @Param task body models.ListTask true "The task object"
// @Success 200 {object} models.ListTask "The created task object."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid task object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the list"
// @Failure 500 {object} models.Message "Internal error"
// @Router /lists/{id} [put]
func (t *ListTask) Create(a web.Auth) (err error) {
	doer, err := getUserWithError(a)
	if err != nil {
		return err
	}

	t.ID = 0

	// Check if we have at least a text
	if t.Text == "" {
		return ErrListTaskCannotBeEmpty{}
	}

	// Check if the list exists
	l := &List{ID: t.ListID}
	if err = l.GetSimpleByID(); err != nil {
		return
	}

	u, err := GetUserByID(doer.ID)
	if err != nil {
		return err
	}

	t.CreatedByID = u.ID
	t.CreatedBy = u
	if _, err = x.Insert(t); err != nil {
		return err
	}

	// Update the assignees
	if err := t.updateTaskAssignees(t.Assignees); err != nil {
		return err
	}

	metrics.UpdateCount(1, metrics.TaskCountKey)
	return
}

// Update updates a list task
// @Summary Update a task
// @Description Updates a task. This includes marking it as done.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param id path int true "Task ID"
// @Param task body models.ListTask true "The task object"
// @Success 200 {object} models.ListTask "The updated task object."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid task object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the task (aka its list)"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{id} [post]
func (t *ListTask) Update() (err error) {
	// Check if the task exists
	ot, err := GetListTaskByID(t.ID)
	if err != nil {
		return
	}

	// When a repeating task is marked as done, we update all deadlines and reminders and set it as undone
	updateDone(&ot, t)

	// Update the assignees
	if err := ot.updateTaskAssignees(t.Assignees); err != nil {
		return err
	}

	// For whatever reason, xorm dont detect if done is updated, so we need to update this every time by hand
	// Which is why we merge the actual task struct with the one we got from the
	// The user struct overrides values in the actual one.
	if err := mergo.Merge(&ot, t, mergo.WithOverride); err != nil {
		return err
	}

	// And because a false is considered to be a null value, we need to explicitly check that case here.
	if t.Done == false {
		ot.Done = false
	}

	_, err = x.ID(t.ID).
		Cols("text",
			"description",
			"done",
			"due_date_unix",
			"reminders_unix",
			"repeat_after",
			"parent_task_id",
			"priority",
			"start_date_unix",
			"end_date_unix").
		Update(ot)
	*t = ot
	return
}

func updateDone(oldTask *ListTask, newTask *ListTask) {
	if !oldTask.Done && newTask.Done && oldTask.RepeatAfter > 0 {
		oldTask.DueDateUnix = oldTask.DueDateUnix + oldTask.RepeatAfter // assuming we'll save the old task (merged)

		for in, r := range oldTask.RemindersUnix {
			oldTask.RemindersUnix[in] = r + oldTask.RepeatAfter
		}

		newTask.Done = false
	}
}

// Create a bunch of task assignees
func (t *ListTask) updateTaskAssignees(assignees []*User) (err error) {

	// Get old assignees to delete
	var found bool
	var assigneesToDelete []int64
	for _, oldAssignee := range t.Assignees {
		found = false
		for _, newAssignee := range assignees {
			if newAssignee.ID == oldAssignee.ID {
				found = true // If a new assignee is already in the list with old assignees
				break
			}
		}

		// Put all assignees which are only on the old list to the trash
		if !found {
			assigneesToDelete = append(assigneesToDelete, oldAssignee.ID)
		}
	}

	// Delete all assignees not passed
	if len(assigneesToDelete) > 0 {
		_, err = x.In("user_id", assigneesToDelete).
			And("task_id = ?", t.ID).
			Delete(ListTaskAssginee{})
		if err != nil {
			return err
		}
	}

	// Get the list to perform later checks
	list := List{ID: t.ListID}
	err = list.ReadOne()
	if err != nil {
		return
	}

	// Loop through our users and add them
AddNewAssignee:
	for _, u := range assignees {
		// Check if the user is already assigned and assign him only if not
		for _, oldAssignee := range t.Assignees {
			if oldAssignee.ID == u.ID {
				// continue outer loop
				continue AddNewAssignee
			}
		}

		// Check if the user exists and has access to the list
		newAssignee, err := GetUserByID(u.ID)
		if err != nil {
			return err
		}
		if !list.CanRead(&newAssignee) {
			return ErrUserDoesNotHaveAccessToList{list.ID, u.ID}
		}

		_, err = x.Insert(ListTaskAssginee{
			TaskID: t.ID,
			UserID: u.ID,
		})
		if err != nil {
			return err
		}
	}

	return
}
