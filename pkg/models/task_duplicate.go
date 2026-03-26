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
	"bytes"
	"io"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// TaskDuplicate holds everything needed to duplicate a task
type TaskDuplicate struct {
	// The task id of the task to duplicate
	TaskID int64 `json:"-" param:"projecttask"`

	// The duplicated task
	Task *Task `json:"duplicated_task,omitempty"`

	web.Permissions `json:"-"`
	web.CRUDable    `json:"-"`
}

// CanCreate checks if a user has the permission to duplicate a task
func (td *TaskDuplicate) CanCreate(s *xorm.Session, a web.Auth) (canCreate bool, err error) {
	// Need read access on the original task
	originalTask := &Task{ID: td.TaskID}
	canRead, _, err := originalTask.CanRead(s, a)
	if err != nil || !canRead {
		return canRead, err
	}

	// Need write access on the project to create tasks in it
	p := &Project{ID: originalTask.ProjectID}
	return p.CanUpdate(s, a)
}

// Create duplicates a task
// @Summary Duplicate a task
// @Description Copies a task with all its properties (labels, assignees, attachments, reminders) into the same project. Creates a "copied from" relation between the new and original task.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "The task ID to duplicate"
// @Success 201 {object} models.TaskDuplicate "The duplicated task."
// @Failure 403 {object} web.HTTPError "The user does not have access to the task."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/duplicate [put]
func (td *TaskDuplicate) Create(s *xorm.Session, doer web.Auth) (err error) {
	// Get the original task with all details
	originalTask := &Task{ID: td.TaskID}
	err = originalTask.ReadOne(s, doer)
	if err != nil {
		return err
	}

	// Create the new task
	newTask := &Task{
		Title:       originalTask.Title,
		Description: originalTask.Description,
		Done:        false,
		DueDate:     originalTask.DueDate,
		ProjectID:   originalTask.ProjectID,
		RepeatAfter: originalTask.RepeatAfter,
		RepeatMode:  originalTask.RepeatMode,
		Priority:    originalTask.Priority,
		StartDate:   originalTask.StartDate,
		EndDate:     originalTask.EndDate,
		HexColor:    originalTask.HexColor,
		PercentDone: originalTask.PercentDone,
		Assignees:   originalTask.Assignees,
		Reminders:   originalTask.Reminders,
	}

	err = createTask(s, newTask, doer, true, true)
	if err != nil {
		return err
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
	}
	if len(labelTasks) > 0 {
		if _, err := s.Insert(&labelTasks); err != nil {
			return err
		}
	}

	// Duplicate attachments (copy underlying files)
	attachments, err := getTaskAttachmentsByTaskIDs(s, []int64{td.TaskID})
	if err != nil {
		return err
	}
	oldToNewAttachmentIDs := make(map[int64]int64)
	for _, attachment := range attachments {
		oldAttachmentID := attachment.ID
		attachment.ID = 0
		attachment.TaskID = newTask.ID
		attachment.File = &files.File{ID: attachment.FileID}
		if err := attachment.File.LoadFileMetaByID(); err != nil {
			if files.IsErrFileDoesNotExist(err) {
				continue
			}
			return err
		}
		if err := attachment.File.LoadFileByID(); err != nil {
			return err
		}

		sourceFile := attachment.File.File
		defer sourceFile.Close()
		buf, err := io.ReadAll(sourceFile)
		if err != nil {
			return err
		}
		err = attachment.NewAttachment(s, bytes.NewReader(buf), attachment.File.Name, attachment.File.Size, doer)
		if err != nil {
			return err
		}
		oldToNewAttachmentIDs[oldAttachmentID] = attachment.ID
	}

	// Re-set the cover image if the original task had one
	if originalTask.CoverImageAttachmentID != 0 {
		if newAttachmentID, ok := oldToNewAttachmentIDs[originalTask.CoverImageAttachmentID]; ok {
			newTask.CoverImageAttachmentID = newAttachmentID
			if _, err := s.Where("id = ?", newTask.ID).
				Cols("cover_image_attachment_id").
				Update(newTask); err != nil {
				return err
			}
		}
	}

	// Create "copied from/to" relation
	rel := &TaskRelation{
		TaskID:       newTask.ID,
		OtherTaskID:  td.TaskID,
		RelationKind: RelationKindCopiedFrom,
	}
	if err := rel.Create(s, doer); err != nil {
		return err
	}

	// Re-read the task to populate all fields for the response
	td.Task = newTask
	return td.Task.ReadOne(s, doer)
}
