// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"time"

	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// LabelTask represents a relation between a label and a task
type LabelTask struct {
	// The unique, numeric id of this label.
	ID     int64 `xorm:"bigint autoincr not null unique pk" json:"-"`
	TaskID int64 `xorm:"bigint INDEX not null" json:"-" param:"projecttask"`
	// The label id you want to associate with a task.
	LabelID int64 `xorm:"bigint INDEX not null" json:"label_id" param:"label"`
	// A timestamp when this task was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName makes a pretty table name
func (*LabelTask) TableName() string {
	return "label_tasks"
}

// LabelTaskServiceProvider is the interface for the label task service
// This interface allows models to call service layer methods without import cycles
type LabelTaskServiceProvider interface {
	AddLabelToTask(s *xorm.Session, labelID, taskID int64, a web.Auth) error
	RemoveLabelFromTask(s *xorm.Session, labelID, taskID int64, a web.Auth) error
	UpdateTaskLabels(s *xorm.Session, taskID int64, newLabels []*Label, a web.Auth) error
	GetLabelsByTaskIDs(s *xorm.Session, opts *LabelByTaskIDsOptions) ([]*LabelWithTaskID, int, int64, error)
}

var labelTaskService LabelTaskServiceProvider

// RegisterLabelTaskService registers the label task service
func RegisterLabelTaskService(service LabelTaskServiceProvider) {
	labelTaskService = service
}

// getLabelTaskService returns the registered label task service
func getLabelTaskService() LabelTaskServiceProvider {
	if labelTaskService == nil {
		panic("LabelTaskService not registered. Make sure to call RegisterLabelTaskService during initialization.")
	}
	return labelTaskService
}

// CanCreate checks if a user can add a label to a task
// Delegates to the service layer via function pointer
func (lt *LabelTask) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckLabelTaskCreateFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	return CheckLabelTaskCreateFunc(s, lt, a)
}

// CanDelete checks if a user can delete a label from a task
// Delegates to the service layer via function pointer
func (lt *LabelTask) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	if CheckLabelTaskDeleteFunc == nil {
		return false, ErrPermissionDelegationNotInitialized{}
	}
	return CheckLabelTaskDeleteFunc(s, lt, a)
}

// Delete deletes a label on a task
// @Summary Remove a label from a task
// @Description Remove a label from a task. The user needs to have write-access to the project to be able do this.
// @tags labels
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param task path int true "Task ID"
// @Param label path int true "Label ID"
// @Success 200 {object} models.Message "The label was successfully removed."
// @Failure 403 {object} web.HTTPError "Not allowed to remove the label."
// @Failure 404 {object} web.HTTPError "Label not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/labels/{label} [delete]
// @Deprecated Use LabelTaskService.RemoveLabelFromTask instead
func (lt *LabelTask) Delete(s *xorm.Session, auth web.Auth) (err error) {
	service := getLabelTaskService()
	return service.RemoveLabelFromTask(s, lt.LabelID, lt.TaskID, auth)
}

// Create adds a label to a task
// @Summary Add a label to a task
// @Description Add a label to a task. The user needs to have write-access to the project to be able do this.
// @tags labels
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param task path int true "Task ID"
// @Param label body models.LabelTask true "The label object"
// @Success 201 {object} models.LabelTask "The created label relation object."
// @Failure 400 {object} web.HTTPError "Invalid label object provided."
// @Failure 403 {object} web.HTTPError "Not allowed to add the label."
// @Failure 404 {object} web.HTTPError "The label does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/labels [put]
// @Deprecated Use LabelTaskService.AddLabelToTask instead
func (lt *LabelTask) Create(s *xorm.Session, auth web.Auth) (err error) {
	service := getLabelTaskService()
	return service.AddLabelToTask(s, lt.LabelID, lt.TaskID, auth)
}

// ReadAll gets all labels on a task
// @Summary Get all labels on a task
// @Description Returns all labels which are assicociated with a given task.
// @tags labels
// @Accept json
// @Produce json
// @Param task path int true "Task ID"
// @Param page query int false "The page number. Used for pagination. If not provided, the first page of results is returned."
// @Param per_page query int false "The maximum number of items per page. Note this parameter is limited by the configured maximum of items per page."
// @Param s query string false "Search labels by label text."
// @Security JWTKeyAuth
// @Success 200 {array} models.Label "The labels"
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{task}/labels [get]
// @Deprecated Use LabelTaskService.GetLabelsByTaskIDs instead
func (lt *LabelTask) ReadAll(s *xorm.Session, a web.Auth, search string, page int, _ int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	service := getLabelTaskService()
	labels, resultCount, totalItems, err := service.GetLabelsByTaskIDs(s, &LabelByTaskIDsOptions{
		User:    a,
		Search:  []string{search},
		Page:    page,
		TaskIDs: []int64{lt.TaskID},
	})
	return labels, resultCount, totalItems, err
}

// LabelWithTaskID is a helper struct, contains the label + its task ID
type LabelWithTaskID struct {
	TaskID int64 `json:"-"`
	Label  `xorm:"extends"`
}

// LabelByTaskIDsOptions is a struct to not clutter the function with too many optional parameters.
type LabelByTaskIDsOptions struct {
	User                web.Auth
	Search              []string
	Page                int
	PerPage             int
	TaskIDs             []int64
	GetUnusedLabels     bool
	GroupByLabelIDsOnly bool
	GetForUser          bool
}

// GetLabelsByTaskIDs is a helper function to get all labels for a set of tasks
// Used when getting all labels for one task as well when getting all labels
// @Deprecated Use LabelTaskService.GetLabelsByTaskIDs instead
func GetLabelsByTaskIDs(s *xorm.Session, opts *LabelByTaskIDsOptions) (ls []*LabelWithTaskID, resultCount int, totalEntries int64, err error) {
	service := getLabelTaskService()
	return service.GetLabelsByTaskIDs(s, opts)
}

// Create or update a bunch of task labels
// @Deprecated Use LabelTaskService.UpdateTaskLabels instead
func (t *Task) UpdateTaskLabels(s *xorm.Session, creator web.Auth, labels []*Label) (err error) {
	service := getLabelTaskService()
	return service.UpdateTaskLabels(s, t.ID, labels, creator)
}

// LabelTaskBulk is a helper struct to update a bunch of labels at once
type LabelTaskBulk struct {
	// All labels you want to update at once.
	Labels []*Label `json:"labels"`
	TaskID int64    `json:"-" param:"projecttask"`

	web.CRUDable    `json:"-"`
	web.Permissions `json:"-"`
}

// CanCreate checks if a user can update labels on a task
// This checks if the user can write to the task
func (ltb *LabelTaskBulk) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	task, err := GetTaskByIDSimple(s, ltb.TaskID)
	if err != nil {
		return false, err
	}
	return task.CanUpdate(s, a)
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
// @Success 201 {object} models.LabelTaskBulk "The updated labels object."
// @Failure 400 {object} web.HTTPError "Invalid label object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/labels/bulk [post]
// @Deprecated Use LabelTaskService.UpdateTaskLabels instead
func (ltb *LabelTaskBulk) Create(s *xorm.Session, a web.Auth) (err error) {
	service := getLabelTaskService()
	return service.UpdateTaskLabels(s, ltb.TaskID, ltb.Labels, a)
}
