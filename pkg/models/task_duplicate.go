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

	log.Debugf("Duplicating task %d", td.TaskID)

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

	log.Debugf("Duplicated task %d into new task %d", td.TaskID, newTask.ID)

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

	log.Debugf("Duplicated labels from task %d into %d", td.TaskID, newTask.ID)

	// Duplicate attachments (copy underlying files)
	attachments, err := getTaskAttachmentsByTaskIDs(s, []int64{td.TaskID})
	if err != nil {
		return err
	}
	for _, attachment := range attachments {
		attachment.ID = 0
		attachment.TaskID = newTask.ID
		attachment.File = &files.File{ID: attachment.FileID}
		if err := attachment.File.LoadFileMetaByID(); err != nil {
			if files.IsErrFileDoesNotExist(err) {
				log.Debugf("Not duplicating attachment (file %d) because it does not exist", attachment.FileID)
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
	}

	log.Debugf("Duplicated attachments from task %d into %d", td.TaskID, newTask.ID)

	// Create "copied from/to" relation
	rel := &TaskRelation{
		TaskID:       newTask.ID,
		OtherTaskID:  td.TaskID,
		RelationKind: RelationKindCopiedFrom,
		CreatedByID:  doer.GetID(),
	}
	if _, err := s.Insert(rel); err != nil {
		return err
	}
	reverseRel := &TaskRelation{
		TaskID:       td.TaskID,
		OtherTaskID:  newTask.ID,
		RelationKind: RelationKindCopiedTo,
		CreatedByID:  doer.GetID(),
	}
	if _, err := s.Insert(reverseRel); err != nil {
		return err
	}

	log.Debugf("Created copy relations between task %d and %d", td.TaskID, newTask.ID)

	// Re-read the task to populate all fields for the response
	td.Task = newTask
	return td.Task.ReadOne(s, doer)
}
