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

package services

import (
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// Label is a service for labels.
type Label struct {
	DB *xorm.Engine
}

// Create creates a new label
func (l *Label) Create(s *xorm.Session, label *models.Label, a interface{}) (*models.Label, error) {
	// Permission check - users can create labels
	u, ok := a.(*user.User)
	if !ok || u == nil {
		return nil, &models.ErrGenericForbidden{}
	}

	// Normalize hex color
	label.HexColor = utils.NormalizeHex(label.HexColor)

	// Set creator information
	label.ID = 0
	label.CreatedBy = u
	label.CreatedByID = u.ID

	_, err := s.Insert(label)
	if err != nil {
		return nil, err
	}

	// Dispatch event
	err = events.Dispatch(&models.LabelCreatedEvent{
		Label: label,
		Doer:  u,
	})
	if err != nil {
		return nil, err
	}

	// Reload the label
	u, ok := a.(*user.User)
	if !ok {
		return nil, &models.ErrGenericForbidden{}
	}
	updatedLabel, err := l.GetByID(s, label.ID, u)
	if err != nil {
		return nil, err
	}

	return updatedLabel, nil
}

// Update updates an existing label
func (l *Label) Update(s *xorm.Session, label *models.Label, a interface{}) (*models.Label, error) {
	u, ok := a.(*user.User)
	if !ok || u == nil {
		return nil, &models.ErrGenericForbidden{}
	}

	// Permission check - only owners can update labels
	can, err := l.canUpdate(s, label, u)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, &models.ErrGenericForbidden{}
	}

	// Normalize hex color
	label.HexColor = utils.NormalizeHex(label.HexColor)

	// Update the label
	_, err = s.
		ID(label.ID).
		Cols(
			"title",
			"description",
			"hex_color",
		).
		Update(label)
	if err != nil {
		return nil, err
	}

	// Dispatch event
	err = events.Dispatch(&models.LabelUpdatedEvent{
		Label: label,
		Doer:  u,
	})
	if err != nil {
		return nil, err
	}

	// Reload the label
	updatedLabel, err := l.GetByID(s, label.ID, u)
	if err != nil {
		return nil, err
	}

	return updatedLabel, nil
}

// Delete deletes a label
func (l *Label) Delete(s *xorm.Session, label *models.Label, a interface{}) error {
	u, ok := a.(*user.User)
	if !ok || u == nil {
		return &models.ErrGenericForbidden{}
	}

	// Permission check - only owners can delete labels
	can, err := l.canDelete(s, label, u)
	if err != nil {
		return err
	}
	if !can {
		return &models.ErrGenericForbidden{}
	}

	// Delete the label
	_, err = s.ID(label.ID).Delete(&models.Label{})
	if err != nil {
		return err
	}

	// Dispatch event
	return events.Dispatch(&models.LabelDeletedEvent{
		Label: label,
		Doer:  u,
	})
}

// GetByID gets a label by its ID
func (l *Label) GetByID(s *xorm.Session, labelID int64, u *user.User) (*models.Label, error) {
	label, err := l.getLabelByIDSimple(s, labelID)
	if err != nil {
		return nil, err
	}

	// Permission check - check if user can read the label
	can, _, err := l.canRead(s, label, u)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, &models.ErrGenericForbidden{}
	}

	// Load creator information
	createdBy, err := user.GetUserByID(s, label.CreatedByID)
	if err != nil {
		return nil, err
	}
	label.CreatedBy = createdBy

	return label, nil
}

// GetAllForUser gets all labels a user can access
func (l *Label) GetAllForUser(s *xorm.Session, a interface{}, search string, page, perPage int) ([]*models.LabelWithTaskID, int, int64, error) {
	return models.GetLabelsByTaskIDs(s, &models.LabelByTaskIDsOptions{
		Search:              []string{search},
		User:                a,
		Page:                page,
		PerPage:             perPage,
		GetUnusedLabels:     true,
		GroupByLabelIDsOnly: true,
		GetForUser:          true,
	})
}

// GetAllForUser gets all labels a user can access

// AddLabelToTask adds a label to a task
func (l *Label) AddLabelToTask(s *xorm.Session, labelTask *models.LabelTask, a interface{}) error {
	u, ok := a.(*user.User)
	if !ok || u == nil {
		return &models.ErrGenericForbidden{}
	}

	// Check if the label is already added
	exists, err := s.Exist(&models.LabelTask{LabelID: labelTask.LabelID, TaskID: labelTask.TaskID})
	if err != nil {
		return err
	}
	if exists {
		return models.ErrLabelIsAlreadyOnTask{LabelID: labelTask.LabelID, TaskID: labelTask.TaskID}
	}

	// Check if user has permission to modify the task
	task := &models.Task{ID: labelTask.TaskID}
	can, _, err := task.CanUpdate(s, a)
	if err != nil {
		return err
	}
	if !can {
		return &models.ErrGenericForbidden{}
	}

	// Check if user has access to the label
	label, err := l.GetLabelSimple(s, &models.Label{ID: labelTask.LabelID})
	if err != nil {
		return err
	}
	hasAccess, _, err := l.canRead(s, label, a)
	if err != nil {
		return err
	}
	if !hasAccess {
		return models.ErrUserHasNoAccessToLabel{LabelID: labelTask.LabelID, UserID: u.ID}
	}

	// Add the label to the task
	labelTask.ID = 0
	_, err = s.Insert(labelTask)
	if err != nil {
		return err
	}

	// Trigger task updated event
	auth, ok := a.(web.Auth)
	if !ok {
		return &models.ErrGenericForbidden{}
	}
	err = models.TriggerTaskUpdatedEventForTaskID(s, auth, labelTask.TaskID)
	if err != nil {
		return err
	}

	// Update project
	return models.UpdateProjectByTaskID(s, labelTask.TaskID)
}

// RemoveLabelFromTask removes a label from a task
func (l *Label) RemoveLabelFromTask(s *xorm.Session, labelTask *models.LabelTask, a interface{}) error {
	u, ok := a.(*user.User)
	if !ok || u == nil {
		return &models.ErrGenericForbidden{}
	}

	// Check if user has permission to modify the task
	task := &models.Task{ID: labelTask.TaskID}
	can, _, err := task.CanUpdate(s, a)
	if err != nil {
		return err
	}
	if !can {
		return &models.ErrGenericForbidden{}
	}

	// Remove the label from the task
	_, err = s.Delete(&models.LabelTask{LabelID: labelTask.LabelID, TaskID: labelTask.TaskID})
	if err != nil {
		return err
	}

	// Trigger task updated event
	return models.TriggerTaskUpdatedEventForTaskID(s, a, labelTask.TaskID)
}

// GetLabelsForTask gets all labels associated with a task
func (l *Label) GetLabelsForTask(s *xorm.Session, taskID int64, a interface{}, search string, page int) ([]*models.LabelWithTaskID, int, int64, error) {
	// Check if the user has the permission to see the task
	task := &models.Task{ID: taskID}
	auth, ok := a.(web.Auth)
	if !ok {
		return nil, 0, 0, models.ErrNoPermissionToSeeTask{TaskID: taskID, UserID: 0}
	}
	canRead, _, err := task.CanRead(s, auth)
	if err != nil {
		return nil, 0, 0, err
	}
	if !canRead {
		userID := int64(0)
		if au, ok := auth.User().(*user.User); ok && au != nil {
			userID = au.ID
		}
		return nil, 0, 0, models.ErrNoPermissionToSeeTask{TaskID: taskID, UserID: userID}
	}

	return models.GetLabelsByTaskIDs(s, &models.LabelByTaskIDsOptions{
		User:    a,
		Search:  []string{search},
		Page:    page,
		TaskIDs: []int64{taskID},
	})
}

// UpdateTaskLabels updates all labels on a task at once
func (l *Label) UpdateTaskLabels(s *xorm.Session, taskID int64, labels []*models.Label, a interface{}) error {
	u, ok := a.(*user.User)
	if !ok || u == nil {
		return &models.ErrGenericForbidden{}
	}

	// Get the task
	task, err := models.GetTaskByIDSimple(s, taskID)
	if err != nil {
		return err
	}

	// Check if user has permission to modify the task
	can, _, err := task.CanUpdate(s, a)
	if err != nil {
		return err
	}
	if !can {
		return &models.ErrGenericForbidden{}
	}

	// Get existing labels for the task
	existingLabels, _, _, err := models.GetLabelsByTaskIDs(s, &models.LabelByTaskIDsOptions{
		TaskIDs: []int64{taskID},
	})
	if err != nil {
		return err
	}

	for i := range existingLabels {
		task.Labels = append(task.Labels, &existingLabels[i].Label)
	}

	return task.UpdateTaskLabels(s, a, labels)
}

func (l *Label) canUpdate(s *xorm.Session, label *models.Label, auth interface{}) (bool, error) {
	return l.isLabelOwner(s, label, auth)
}

func (l *Label) canDelete(s *xorm.Session, label *models.Label, auth interface{}) (bool, error) {
	return l.isLabelOwner(s, label, auth)
}

// Helper method to check if a user can read a label (moved from model)
func (l *Label) canRead(s *xorm.Session, label *models.Label, a interface{}) (bool, int, error) {
	return l.hasAccessToLabel(s, label, a)
}

// Helper method to check if a user is the owner of a label (moved from model)
func (l *Label) isLabelOwner(s *xorm.Session, label *models.Label, a interface{}) (bool, error) {
	// Link sharing users cannot be owners
	u, ok := a.(*user.User)
	if !ok || u == nil {
		return false, nil
	}

	labelOrig, err := l.getLabelByIDSimple(s, label.ID)
	if err != nil {
		return false, err
	}
	return labelOrig.CreatedByID == u.ID, nil
}

// Helper method to check if a user has access to a specific label (moved from model)
func (l *Label) hasAccessToLabel(s *xorm.Session, label *models.Label, a interface{}) (has bool, maxPermission int, err error) {
	// This logic is moved from the model's hasAccessToLabel method
	linkShare, isLinkShare := a.(*models.LinkSharing)

	var where builder.Cond
	var createdByID int64
	if isLinkShare {
		where = builder.Eq{"project_id": linkShare.ProjectID}
	} else {
		u, ok := a.(*user.User)
		if !ok {
			return false, 0, nil
		}
		where = builder.In("project_id", models.GetUserProjectsStatement(u.ID, "", false).Select("l.id"))
		createdByID = u.ID
	}

	cond := builder.In("label_tasks.task_id",
		builder.
			Select("id").
			From("tasks").
			Where(where),
	)

	ll := &models.LabelTask{}
	has, err = s.Table("labels").
		Select("label_tasks.*").
		Join("LEFT", "label_tasks", "label_tasks.label_id = labels.id").
		Where(builder.Or(
			builder.And(builder.NotNull{"label_tasks.label_id"}, builder.Eq{"labels.created_by_id": createdByID}),
			cond,
		)).
		And("labels.id = ?", label.ID).
		Exist(ll)
	if err != nil {
		return
	}

	// Since the permission depends on the task the label is associated with, we need to check that too.
	if ll.TaskID > 0 {
		t := &models.Task{ID: ll.TaskID}
		_, maxPermission, err = t.CanRead(s, a)
		if err != nil {
			return
		}
	}

	return
}

// Helper method to get a label by ID (moved from model)
func (l *Label) getLabelByIDSimple(s *xorm.Session, labelID int64) (*models.Label, error) {
	return l.GetLabelSimple(s, &models.Label{ID: labelID})
}

// Helper method to get a label by its properties (moved from model)
func (l *Label) GetLabelSimple(s *xorm.Session, label *models.Label) (*models.Label, error) {
	exists, err := s.Get(label)
	if err != nil {
		return label, err
	}
	if !exists {
		return &models.Label{}, models.ErrLabelDoesNotExist{LabelID: label.ID}
	}
	return label, err
}

// Adapter methods for the model function variables
func (l *Label) CreateFromModel(s *xorm.Session, label *models.Label, a web.Auth) error {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	createdLabel, err := l.Create(s, label, u)
	if err != nil {
		return err
	}

	// Copy the created label back to the original instance
	*label = *createdLabel
	return nil
}

func (l *Label) UpdateFromModel(s *xorm.Session, label *models.Label, a web.Auth) error {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	updatedLabel, err := l.Update(s, label, u)
	if err != nil {
		return err
	}

	// Copy the updated label back to the original instance
	*label = *updatedLabel
	return nil
}

func (l *Label) DeleteFromModel(s *xorm.Session, label *models.Label, a web.Auth) error {
	u, err := user.GetFromAuth(a)
	if err != nil {
		return err
	}

	return l.Delete(s, label, u)
}

func (l *Label) ReadOneFromModel(s *xorm.Session, label *models.Label, a web.Auth) error {
	updatedLabel, err := l.GetByID(s, label.ID, a)
	if err != nil {
		return err
	}

	// Copy the label back to the original instance
	*label = *updatedLabel
	return nil
}

func init() {
	// Set the model function variables to delegate to the service
	models.LabelCreateFunc = NewLabelService().CreateFromModel
	models.LabelUpdateFunc = NewLabelService().UpdateFromModel
	models.LabelDeleteFunc = NewLabelService().DeleteFromModel
	models.LabelReadAllFunc = func(s *xorm.Session, a web.Auth, search string, page int, perPage int) (ls interface{}, resultCount int, numberOfEntries int64, err error) {
		service := NewLabelService()
		labels, count, total, err := service.GetAllForUser(s, a, search, page, perPage)
		return labels, count, total, err
	}
	models.LabelReadOneFunc = NewLabelService().ReadOneFromModel

	// Set the label task function variables to delegate to the service
	models.LabelTaskCreateFunc = func(s *xorm.Session, lt *models.LabelTask, auth web.Auth) error {
		service := NewLabelService()
		return service.AddLabelToTask(s, lt, auth)
	}

	// models.LabelTaskUpdateFunc = func(s *xorm.Session, lt *models.LabelTask, auth web.Auth) error {
	// 	service := NewLabelService()
	// 	return service.UpdateLabelTask(s, lt, auth)
	// }

	models.LabelTaskDeleteFunc = func(s *xorm.Session, lt *models.LabelTask, auth web.Auth) error {
		service := NewLabelService()
		return service.RemoveLabelFromTask(s, lt, auth)
	}

	models.LabelTaskReadAllFunc = func(s *xorm.Session, taskID int64, a web.Auth, search string, page int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
		service := NewLabelService()
		return service.GetLabelsForTask(s, taskID, a, search, page)
	}

	models.LabelTaskBulkCreateFunc = func(s *xorm.Session, taskID int64, labels []*models.Label, auth web.Auth) error {
		service := NewLabelService()
		return service.UpdateTaskLabels(s, taskID, labels, auth)
	}
}

// NewLabelService creates a new label service instance
func NewLabelService() *Label {
	return &Label{
		DB: db.GetEngine(),
	}
}

// Helper method to get a label by ID (moved from model)
/* Removed duplicated getLabelByIDSimple method */

// Helper method to get a label by its properties (moved from model)
