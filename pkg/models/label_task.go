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
	"github.com/go-xorm/builder"
)

// Delete deletes a label on a task
// @Summary Remove a label from a task
// @Description Remove a label from a task. The user needs to have write-access to the list to be able do this.
// @tags labels
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param task path int true "Task ID"
// @Param label path int true "Label ID"
// @Success 200 {object} models.Label "The label was successfully removed."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "Not allowed to remove the label."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "Label not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/labels/{label} [delete]
func (l *LabelTask) Delete() (err error) {
	_, err = x.Delete(&LabelTask{LabelID: l.LabelID, TaskID: l.TaskID})
	return err
}

// Create adds a label to a task
// @Summary Add a label to a task
// @Description Add a label to a task. The user needs to have write-access to the list to be able do this.
// @tags labels
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param task path int true "Task ID"
// @Param label body models.Label true "The label object"
// @Success 200 {object} models.Label "The created label relation object."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid label object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "Not allowed to add the label."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "The label does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/labels [put]
func (l *LabelTask) Create(a web.Auth) (err error) {
	// Check if the label is already added
	exists, err := x.Exist(&LabelTask{LabelID: l.LabelID, TaskID: l.TaskID})
	if err != nil {
		return err
	}
	if exists {
		return ErrLabelIsAlreadyOnTask{l.LabelID, l.TaskID}
	}

	// Insert it
	_, err = x.Insert(l)
	return err
}

// ReadAll gets all labels on a task
// @Summary Get all labels on a task
// @Description Returns all labels which are assicociated with a given task.
// @tags labels
// @Accept json
// @Produce json
// @Param task path int true "Task ID"
// @Param p query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param s query string false "Search labels by label text."
// @Security JWTKeyAuth
// @Success 200 {array} models.Label "The labels"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/labels [get]
func (l *LabelTask) ReadAll(search string, a web.Auth, page int) (labels interface{}, err error) {
	u, err := getUserWithError(a)
	if err != nil {
		return nil, err
	}

	// Check if the user has the right to see the task
	task, err := GetListTaskByID(l.TaskID)
	if err != nil {
		return nil, err
	}

	if !task.CanRead(a) {
		return nil, ErrNoRightToSeeTask{l.TaskID, u.ID}
	}

	return getLabelsByTaskIDs(search, u, page, []int64{l.TaskID}, false)
}

type labelWithTaskID struct {
	TaskID int64
	Label  `xorm:"extends"`
}

// Helper function to get all labels for a set of tasks
// Used when getting all labels for one task as well when getting all lables
func getLabelsByTaskIDs(search string, u *User, page int, taskIDs []int64, getUnusedLabels bool) (ls []*labelWithTaskID, err error) {
	// Incl unused labels
	var uidOrNil interface{}
	var requestOrNil interface{}
	if getUnusedLabels {
		uidOrNil = u.ID
		requestOrNil = "label_task.label_id != null OR labels.created_by_id = ?"
	}

	// Get all labels associated with these labels
	var labels []*labelWithTaskID
	err = x.Table("labels").
		Select("labels.*, label_task.task_id").
		Join("LEFT", "label_task", "label_task.label_id = labels.id").
		Where(requestOrNil, uidOrNil).
		Or(builder.In("label_task.task_id", taskIDs)).
		And("labels.title LIKE ?", "%"+search+"%").
		GroupBy("labels.id").
		Limit(getLimitFromPageIndex(page)).
		Find(&labels)
	if err != nil {
		return nil, err
	}

	// Get all created by users
	var userids []int64
	for _, l := range labels {
		userids = append(userids, l.CreatedByID)
	}
	users := make(map[int64]*User)
	err = x.In("id", userids).Find(&users)
	if err != nil {
		return nil, err
	}

	// Put it all together
	for in, l := range labels {
		labels[in].CreatedBy = users[l.CreatedByID]
	}

	return labels, err
}
