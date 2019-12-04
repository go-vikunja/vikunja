// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2019 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/web"
	"github.com/imdario/mergo"
)

// BulkTask is the definition of a bulk update task
type BulkTask struct {
	// A list of task ids to update
	IDs   []int64 `json:"task_ids"`
	Tasks []*Task `json:"-"`
	Task
}

func (bt *BulkTask) checkIfTasksAreOnTheSameList() (err error) {
	// Get the tasks
	err = bt.GetTasksByIDs()
	if err != nil {
		return err
	}

	if len(bt.Tasks) == 0 {
		return ErrBulkTasksNeedAtLeastOne{}
	}

	// Check if all tasks are in the same list
	var firstListID = bt.Tasks[0].ListID
	for _, t := range bt.Tasks {
		if t.ListID != firstListID {
			return ErrBulkTasksMustBeInSameList{firstListID, t.ListID}
		}
	}

	return nil
}

// CanUpdate checks if a user is allowed to update a task
func (bt *BulkTask) CanUpdate(a web.Auth) (bool, error) {

	err := bt.checkIfTasksAreOnTheSameList()
	if err != nil {
		return false, err
	}

	// A user can update an task if he has write acces to its list
	l := &List{ID: bt.Tasks[0].ListID}
	return l.CanWrite(a)
}

// Update updates a bunch of tasks at once
// @Summary Update a bunch of tasks at once
// @Description Updates a bunch of tasks at once. This includes marking them as done. Note: although you could supply another ID, it will be ignored. Use task_ids instead.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param task body models.BulkTask true "The task object. Looks like a normal task, the only difference is it uses an array of list_ids to update."
// @Success 200 {object} models.Task "The updated task object."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid task object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the task (aka its list)"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/bulk [post]
func (bt *BulkTask) Update() (err error) {

	sess := x.NewSession()
	defer sess.Close()

	err = sess.Begin()
	if err != nil {
		return
	}

	for _, oldtask := range bt.Tasks {

		// When a repeating task is marked as done, we update all deadlines and reminders and set it as undone
		updateDone(oldtask, &bt.Task)

		// Update the assignees
		if err := oldtask.updateTaskAssignees(bt.Assignees); err != nil {
			return err
		}

		// For whatever reason, xorm dont detect if done is updated, so we need to update this every time by hand
		// Which is why we merge the actual task struct with the one we got from the
		// The user struct overrides values in the actual one.
		if err := mergo.Merge(oldtask, &bt.Task, mergo.WithOverride); err != nil {
			return err
		}

		// And because a false is considered to be a null value, we need to explicitly check that case here.
		if !bt.Task.Done {
			oldtask.Done = false
		}

		_, err = sess.ID(oldtask.ID).
			Cols("text",
				"description",
				"done",
				"due_date_unix",
				"reminders_unix",
				"repeat_after",
				"priority",
				"start_date_unix",
				"end_date_unix").
			Update(oldtask)
		if err != nil {
			return sess.Rollback()
		}
	}

	err = sess.Commit()
	if err != nil {
		return
	}

	return
}
