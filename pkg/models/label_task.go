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

// LabelTask represents a relation between a label and a task
type LabelTask struct {
	// The unique, numeric id of this label.
	ID     int64 `xorm:"int(11) autoincr not null unique pk" json:"-"`
	TaskID int64 `xorm:"int(11) INDEX not null" json:"-" param:"listtask"`
	// The label id you want to associate with a task.
	LabelID int64 `xorm:"int(11) INDEX not null" json:"label_id" param:"label"`
	// A unix timestamp when this task was created. You cannot change this value.
	Created int64 `xorm:"created not null" json:"created"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName makes a pretty table name
func (LabelTask) TableName() string {
	return "label_task"
}

// Delete deletes a label on a task
// @Summary Remove a label from a task
// @Description Remove a label from a task. The user needs to have write-access to the list to be able do this.
// @tags labels
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param task path int true "Task ID"
// @Param label path int true "Label ID"
// @Success 200 {object} models.Message "The label was successfully removed."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "Not allowed to remove the label."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "Label not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/labels/{label} [delete]
func (lt *LabelTask) Delete() (err error) {
	_, err = x.Delete(&LabelTask{LabelID: lt.LabelID, TaskID: lt.TaskID})
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
// @Param label body models.LabelTask true "The label object"
// @Success 200 {object} models.LabelTask "The created label relation object."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid label object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "Not allowed to add the label."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "The label does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/labels [put]
func (lt *LabelTask) Create(a web.Auth) (err error) {
	// Check if the label is already added
	exists, err := x.Exist(&LabelTask{LabelID: lt.LabelID, TaskID: lt.TaskID})
	if err != nil {
		return err
	}
	if exists {
		return ErrLabelIsAlreadyOnTask{lt.LabelID, lt.TaskID}
	}

	// Insert it
	_, err = x.Insert(lt)
	return
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
func (lt *LabelTask) ReadAll(search string, a web.Auth, page int) (labels interface{}, err error) {
	u, err := getUserWithError(a)
	if err != nil {
		return nil, err
	}

	// Check if the user has the right to see the task
	task := ListTask{ID: lt.TaskID}
	canRead, err := task.CanRead(a)
	if err != nil {
		return nil, err
	}
	if !canRead {
		return nil, ErrNoRightToSeeTask{lt.TaskID, u.ID}
	}

	return getLabelsByTaskIDs(&LabelByTaskIDsOptions{
		User:    u,
		Search:  search,
		Page:    page,
		TaskIDs: []int64{lt.TaskID},
	})
}

// Helper struct, contains the label + its task ID
type labelWithTaskID struct {
	TaskID int64
	Label  `xorm:"extends"`
}

// LabelByTaskIDsOptions is a struct to not clutter the function with too many optional parameters.
type LabelByTaskIDsOptions struct {
	User            *User
	Search          string
	Page            int
	TaskIDs         []int64
	GetUnusedLabels bool
}

// Helper function to get all labels for a set of tasks
// Used when getting all labels for one task as well when getting all lables
func getLabelsByTaskIDs(opts *LabelByTaskIDsOptions) (ls []*labelWithTaskID, err error) {
	// Include unused labels. Needed to be able to show a list of all unused labels a user
	// has access to.
	var uidOrNil interface{}
	var requestOrNil interface{}
	if opts.GetUnusedLabels {
		uidOrNil = opts.User.ID
		requestOrNil = "label_task.label_id != null OR labels.created_by_id = ?"
	}

	// Get all labels associated with these tasks
	var labels []*labelWithTaskID
	err = x.Table("labels").
		Select("labels.*, label_task.task_id").
		Join("LEFT", "label_task", "label_task.label_id = labels.id").
		Where(requestOrNil, uidOrNil).
		Or(builder.In("label_task.task_id", opts.TaskIDs)).
		And("labels.title LIKE ?", "%"+opts.Search+"%").
		GroupBy("labels.id,label_task.task_id"). // This filters out doubles
		Limit(getLimitFromPageIndex(opts.Page)).
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

// Create or update a bunch of task labels
func (t *ListTask) updateTaskLabels(creator web.Auth, labels []*Label) (err error) {

	// If we don't have any new labels, delete everything right away. Saves us some hassle.
	if len(labels) == 0 && len(t.Labels) > 0 {
		_, err = x.Where("task_id = ?", t.ID).
			Delete(LabelTask{})
		return err
	}

	// If we didn't change anything (from 0 to zero) don't do anything.
	if len(labels) == 0 && len(t.Labels) == 0 {
		return nil
	}

	// Make a hashmap of the new labels for easier comparison
	newLabels := make(map[int64]*Label, len(labels))
	for _, newLabel := range labels {
		newLabels[newLabel.ID] = newLabel
	}

	// Get old labels to delete
	var found bool
	var labelsToDelete []int64
	oldLabels := make(map[int64]*Label, len(t.Labels))
	allLabels := t.Labels
	t.Labels = []*Label{} // We re-empty our labels struct here because we want it to be fully empty so we can put in all the actual labels.
	for _, oldLabel := range allLabels {
		found = false
		if newLabels[oldLabel.ID] != nil {
			found = true // If a new label is already in the list with old labels
		}

		// Put all labels which are only on the old list to the trash
		if !found {
			labelsToDelete = append(labelsToDelete, oldLabel.ID)
		} else {
			t.Labels = append(t.Labels, oldLabel)
		}

		// Put it in a list with all old labels, just using the loop here
		oldLabels[oldLabel.ID] = oldLabel
	}

	// Delete all labels not passed
	if len(labelsToDelete) > 0 {
		_, err = x.In("label_id", labelsToDelete).
			And("task_id = ?", t.ID).
			Delete(LabelTask{})
		if err != nil {
			return err
		}
	}

	// Loop through our labels and add them
	for _, l := range labels {
		// Check if the label is already added on the task and only add it if not
		if oldLabels[l.ID] != nil {
			// continue outer loop
			continue
		}

		// Add the new label
		label, err := getLabelByIDSimple(l.ID)
		if err != nil {
			return err
		}

		// Check if the user has the rights to see the label he is about to add
		hasAccessToLabel, err := label.hasAccessToLabel(creator)
		if err != nil {
			return err
		}
		if !hasAccessToLabel {
			user, _ := creator.(*User)
			return ErrUserHasNoAccessToLabel{LabelID: l.ID, UserID: user.ID}
		}

		// Insert it
		_, err = x.Insert(&LabelTask{LabelID: l.ID, TaskID: t.ID})
		if err != nil {
			return err
		}
		t.Labels = append(t.Labels, label)
	}
	return
}

// LabelTaskBulk is a helper struct to update a bunch of labels at once
type LabelTaskBulk struct {
	// All labels you want to update at once.
	Labels []*Label `json:"labels"`
	TaskID int64    `json:"-" param:"listtask"`

	web.CRUDable `json:"-"`
	web.Rights   `json:"-"`
}

// Create updates a bunch of labels on a task at once
// @Summary Update all labels on a task.
// @Description Updates all labels on a task. Every label which is not passed but exists on the task will be deleted. Every label which does not exist on the task will be added. All labels which are passed and already exist on the task won't be touched.
// @tags labels
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param label body models.LabelTaskBulk true "The array of labels"
// @Param taskID path int true "Task ID"
// @Success 200 {object} models.LabelTaskBulk "The updated labels object."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid label object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/labels/bulk [post]
func (ltb *LabelTaskBulk) Create(a web.Auth) (err error) {
	task, err := GetListTaskByID(ltb.TaskID)
	if err != nil {
		return
	}
	return task.updateTaskLabels(a, ltb.Labels)
}
