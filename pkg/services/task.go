// Vikunja is a to-do list application to facilitate your life.
// Adding a comment to force a recompile and check the line number of the error.
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
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"strconv"
	"xorm.io/xorm"
)

// TaskService represents a service for managing tasks.
type TaskService struct {
	DB *xorm.Engine
}

// NewTaskService creates a new TaskService.
func NewTaskService(db *xorm.Engine) *TaskService {
	return &TaskService{DB: db}
}

// Update updates a task.
func (ts *TaskService) Update(s *xorm.Session, task *models.Task, u *user.User) (*models.Task, error) {
	can, err := ts.Can(s, task, u).Write()
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrAccessDenied
	}

	// The old logic used task.Update which did a lot of things.
	// We need to replicate that logic here.
	// For now, we'll just do a simple update.
	if _, err := s.ID(task.ID).AllCols().Update(task); err != nil {
		return nil, err
	}
	return task, nil
}


// Delete deletes a task.
func (ts *TaskService) Delete(s *xorm.Session, task *models.Task, a web.Auth) error {
	t, err := models.GetTaskByIDSimple(s, task.ID)
	if err != nil {
		return err
	}

	can, err := ts.canWriteTask(s, task.ID, a)
	if err != nil {
		return err
	}
	if !can {
		return ErrAccessDenied
	}

	// duplicate the task for the event
	fullTask := &models.Task{ID: task.ID}
	err = fullTask.ReadOne(s, a)
	if err != nil {
		return err
	}

	// Delete assignees
	if _, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskAssginee{}); err != nil {
		return err
	}

	// Delete Favorites
	err = models.RemoveFromFavorite(s, task.ID, a, models.FavoriteKindTask)
	if err != nil {
		return err
	}

	// Delete label associations
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.LabelTask{})
	if err != nil {
		return err
	}

	// Delete task attachments
	attachments, err := ts.getTaskAttachmentsByTaskIDs(s, []int64{task.ID})
	if err != nil {
		return err
	}
	for _, attachment := range attachments {
		// Using the attachment delete method here because that takes care of removing all files properly
		err = attachment.Delete(s, a)
		if err != nil && !models.IsErrTaskAttachmentDoesNotExist(err) {
			return err
		}
	}

	// Delete all comments
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskComment{})
	if err != nil {
		return err
	}

	// Delete all relations
	_, err = s.Where("task_id = ? OR other_task_id = ?", task.ID, task.ID).Delete(&models.TaskRelation{})
	if err != nil {
		return err
	}

	// Delete all reminders
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskReminder{})
	if err != nil {
		return err
	}

	// Delete all positions
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskPosition{})
	if err != nil {
		return err
	}

	// Delete all bucket relations
	_, err = s.Where("task_id = ?", task.ID).Delete(&models.TaskBucket{})
	if err != nil {
		return err
	}

	// Actually delete the task
	_, err = s.ID(task.ID).Delete(&models.Task{})
	if err != nil {
		return err
	}

	doer, _ := user.GetFromAuth(a)
	err = events.Dispatch(&models.TaskDeletedEvent{
		Task: fullTask,
		Doer: doer,
	})
	if err != nil {
		return err
	}

	err = ts.updateProjectLastUpdated(s, &models.Project{ID: t.ProjectID})
	return err
}

// TaskPermissions represents the permissions for a task.
type TaskPermissions struct {
	s    *xorm.Session
	task *models.Task
	user *user.User
}

// Can returns a new TaskPermissions struct.
func (ts *TaskService) Can(s *xorm.Session, task *models.Task, u *user.User) *TaskPermissions {
	return &TaskPermissions{s: s, task: task, user: u}
}

// Read checks if the user can read the task.
func (tp *TaskPermissions) Read() (bool, error) {
	if tp.user == nil {
		return false, nil
	}
	can, _, err := tp.task.CanRead(tp.s, tp.user)
	return can, err
}

// Write checks if the user can write to the task.
func (tp *TaskPermissions) Write() (bool, error) {
	if tp.user == nil {
		return false, nil
	}
	can, err := tp.task.CanWrite(tp.s, tp.user)
	return can, err
}

func (ts *TaskService) updateProjectLastUpdated(s *xorm.Session, project *models.Project) error {
	_, err := s.ID(project.ID).Cols("updated").Update(project)
	return err
}

func (ts *TaskService) getUsersOrLinkSharesFromIDs(s *xorm.Session, ids []int64) (users map[int64]*user.User, err error) {
	users = make(map[int64]*user.User)
	var userIDs []int64
	var linkShareIDs []int64
	for _, id := range ids {
		if id < 0 {
			linkShareIDs = append(linkShareIDs, id*-1)
			continue
		}

		userIDs = append(userIDs, id)
	}

	if len(userIDs) > 0 {
		users, err = user.GetUsersByIDs(s, userIDs)
		if err != nil {
			return
		}
	}

	if len(linkShareIDs) == 0 {
		return
	}

	shares, err := models.GetLinkSharesByIDs(s, linkShareIDs)
	if err != nil {
		return nil, err
	}

	for _, share := range shares {
		users[share.ID*-1] = ts.toUser(share)
	}

	return
}

func (ts *TaskService) toUser(share *models.LinkSharing) *user.User {
	suffix := "Link Share"
	if share.Name != "" {
		suffix = " (" + suffix + ")"
	}

	username := "link-share-" + strconv.FormatInt(share.ID, 10)

	return &user.User{
		ID:       ts.getUserID(share),
		Name:     share.Name + suffix,
		Username: username,
		Created:  share.Created,
		Updated:  share.Updated,
	}
}

func (ts *TaskService) getUserID(share *models.LinkSharing) int64 {
	return share.ID * -1
}

func (ts *TaskService) canWriteTask(s *xorm.Session, taskID int64, a web.Auth) (bool, error) {
	project, err := models.GetProjectSimpleByTaskID(s, taskID)
	if err != nil {
		if models.IsErrProjectDoesNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return project.CanWrite(s, a)
}

func (ts *TaskService) getTaskAttachmentsByTaskIDs(s *xorm.Session, taskIDs []int64) (attachments []*models.TaskAttachment, err error) {
	attachments = []*models.TaskAttachment{}
	err = s.
		In("task_id", taskIDs).
		Find(&attachments)
	if err != nil {
		return
	}

	if len(attachments) == 0 {
		return
	}

	fileIDs := []int64{}
	userIDs := []int64{}
	for _, a := range attachments {
		userIDs = append(userIDs, a.CreatedByID)
		fileIDs = append(fileIDs, a.FileID)
	}

	// Get all files
	fs := make(map[int64]*files.File)
	err = s.In("id", fileIDs).Find(&fs)
	if err != nil {
		return
	}

	users, err := ts.getUsersOrLinkSharesFromIDs(s, userIDs)
	if err != nil {
		return nil, err
	}

	// Obfuscate all user emails
	for _, u := range users {
		u.Email = ""
	}

	for _, a := range attachments {
		if createdBy, has := users[a.CreatedByID]; has {
			a.CreatedBy = createdBy
		}
		a.File = fs[a.FileID]
	}

	return
}
