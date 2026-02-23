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
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// TaskDuplicate holds everything needed to duplicate a task
type TaskDuplicate struct {
	// The task id of the task to duplicate
	TaskID int64 `json:"-" param:"task"`
	// The target project id to duplicate the task into
	TargetProjectID int64 `json:"target_project_id"`

	// The duplicated task
	Task *Task `json:"duplicated_task,omitempty"`

	web.Permissions `json:"-"`
	web.CRUDable    `json:"-"`
}

// CanCreate checks if the user has the right to duplicate a task
func (td *TaskDuplicate) CanCreate(s *xorm.Session, a web.Auth) (canCreate bool, err error) {
	// The user needs read access to the original task
	originalTask := &Task{ID: td.TaskID}
	canRead, _, err := originalTask.CanRead(s, a)
	if err != nil || !canRead {
		return canRead, err
	}

	// The user needs write access to the target project (to create tasks in it)
	targetProject := &Project{ID: td.TargetProjectID}
	return targetProject.CanUpdate(s, a)
}

// Create duplicates a task into the target project
// @Summary Duplicate a task
// @Description Copies a task with all its metadata (description, labels, assignees, attachments, comments, reminders, relations, subtasks) into a target project.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "The task ID to duplicate"
// @Param task body models.TaskDuplicate true "The target project into which the task should be duplicated."
// @Success 201 {object} models.TaskDuplicate "The duplicated task."
// @Failure 400 {object} web.HTTPError "Invalid task duplicate object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task or the target project."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/duplicate [put]
func (td *TaskDuplicate) Create(s *xorm.Session, doer web.Auth) (err error) {

	// Get the original task with all details
	originalTask := &Task{ID: td.TaskID}
	err = originalTask.ReadOne(s, doer)
	if err != nil {
		return err
	}

	log.Debugf("Duplicating task %d into project %d", td.TaskID, td.TargetProjectID)

	// Create the new task
	newTask := &Task{}
	*newTask = *originalTask
	newTask.ID = 0
	newTask.ProjectID = td.TargetProjectID
	newTask.UID = ""
	newTask.Title = originalTask.Title
	newTask.Position = 0 // Let the system assign a position

	err = createTask(s, newTask, doer, false, true)
	if err != nil {
		return err
	}

	log.Debugf("Duplicated task %d into new task %d in project %d", td.TaskID, newTask.ID, td.TargetProjectID)

	// Duplicate attachments
	attachments, err := getTaskAttachmentsByTaskIDs(s, []int64{td.TaskID})
	if err != nil {
		return err
	}

	attachmentMap := make(map[int64]int64) // old attachment ID -> new attachment ID
	for _, attachment := range attachments {
		oldAttachmentID := attachment.ID
		attachment.ID = 0
		attachment.TaskID = newTask.ID
		attachment.File = &files.File{ID: attachment.FileID}
		if err := attachment.File.LoadFileMetaByID(); err != nil {
			if files.IsErrFileDoesNotExist(err) {
				log.Debugf("Not duplicating attachment %d (file %d) because it does not exist", oldAttachmentID, attachment.FileID)
				continue
			}
			return err
		}
		if err := attachment.File.LoadFileByID(); err != nil {
			return err
		}

		err := attachment.NewAttachment(s, attachment.File.File, attachment.File.Name, attachment.File.Size, doer)
		if err != nil {
			return err
		}

		if attachment.File.File != nil {
			_ = attachment.File.File.Close()
		}

		attachmentMap[oldAttachmentID] = attachment.ID
		log.Debugf("Duplicated attachment %d into %d for task %d", oldAttachmentID, attachment.ID, newTask.ID)
	}

	// Update cover image if the original task had one
	if originalTask.CoverImageAttachmentID != 0 {
		if newCoverID, exists := attachmentMap[originalTask.CoverImageAttachmentID]; exists {
			newTask.CoverImageAttachmentID = newCoverID
			_, err = s.Where("id = ?", newTask.ID).Cols("cover_image_attachment_id").Update(newTask)
			if err != nil {
				return err
			}
		}
	}

	// Duplicate labels
	labelTasks := []*LabelTask{}
	err = s.Where("task_id = ?", td.TaskID).Find(&labelTasks)
	if err != nil {
		return err
	}
	for _, lt := range labelTasks {
		lt.ID = 0
		lt.TaskID = newTask.ID
		if _, err := s.Insert(lt); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated %d labels for task %d", len(labelTasks), newTask.ID)

	// Duplicate assignees (only those who have access to the target project)
	assignees := []*TaskAssginee{}
	err = s.Where("task_id = ?", td.TaskID).Find(&assignees)
	if err != nil {
		return err
	}
	for _, a := range assignees {
		t := &Task{
			ID:        newTask.ID,
			ProjectID: td.TargetProjectID,
		}
		targetProject := &Project{ID: td.TargetProjectID}
		if err := t.addNewAssigneeByID(s, a.UserID, targetProject, doer); err != nil {
			if IsErrUserDoesNotHaveAccessToProject(err) {
				continue
			}
			return err
		}
	}

	log.Debugf("Duplicated assignees for task %d", newTask.ID)

	// Duplicate comments
	comments := []*TaskComment{}
	err = s.Where("task_id = ?", td.TaskID).Find(&comments)
	if err != nil {
		return err
	}
	for _, c := range comments {
		c.ID = 0
		c.TaskID = newTask.ID
		if _, err := s.Insert(c); err != nil {
			return err
		}
	}

	log.Debugf("Duplicated %d comments for task %d", len(comments), newTask.ID)

	// Read the full duplicated task to return it
	newTask.ID = newTask.ID
	err = newTask.ReadOne(s, doer)
	if err != nil {
		return err
	}

	td.Task = newTask

	return nil
}
