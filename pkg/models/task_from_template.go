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
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// TaskFromTemplate holds everything needed to create a task from a template
type TaskFromTemplate struct {
	// The template id to create the task from
	TemplateID int64 `json:"-" param:"template"`
	// The target project id
	TargetProjectID int64 `json:"target_project_id"`
	// Optional title override. If empty, uses the template title.
	Title string `json:"title"`

	// The created task (returned after creation)
	Task *Task `json:"created_task,omitempty"`

	web.Permissions `json:"-"`
	web.CRUDable    `json:"-"`
}

// CanCreate checks if the user can create a task from this template
func (tft *TaskFromTemplate) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	// User must own the template
	tt := &TaskTemplate{ID: tft.TemplateID}
	exists, err := s.Where("id = ?", tft.TemplateID).Get(tt)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, &ErrTaskTemplateDoesNotExist{ID: tft.TemplateID}
	}
	if tt.OwnerID != a.GetID() {
		return false, nil
	}

	// User must have write access to the target project
	targetProject := &Project{ID: tft.TargetProjectID}
	return targetProject.CanUpdate(s, a)
}

// Create creates a new task from a template
// @Summary Create a task from a template
// @Description Creates a new task in the target project using the template's predefined values.
// @tags task_templates
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param templateID path int true "The template ID"
// @Param task body models.TaskFromTemplate true "The target project and optional title override"
// @Success 201 {object} models.TaskFromTemplate "The created task."
// @Failure 400 {object} web.HTTPError "Invalid request."
// @Failure 403 {object} web.HTTPError "Not authorized."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasktemplates/{templateID}/tasks [put]
func (tft *TaskFromTemplate) Create(s *xorm.Session, doer web.Auth) (err error) {

	// Load the template
	tt := &TaskTemplate{ID: tft.TemplateID}
	exists, err := s.Where("id = ?", tft.TemplateID).Get(tt)
	if err != nil {
		return err
	}
	if !exists {
		return &ErrTaskTemplateDoesNotExist{ID: tft.TemplateID}
	}

	log.Debugf("Creating task from template %d in project %d", tft.TemplateID, tft.TargetProjectID)

	// Determine the title
	title := tt.Title
	if tft.Title != "" {
		title = tft.Title
	}

	// Create the task with template values
	description := tt.Description
	// Append a small note indicating this task was created from a template
	templateNote := `<p><small style="color:#888">📋 Created from template: ` + tt.Title + `</small></p>`
	if description != "" {
		description = description + templateNote
	} else {
		description = templateNote
	}

	newTask := &Task{
		Title:       title,
		Description: description,
		Priority:    tt.Priority,
		HexColor:    tt.HexColor,
		PercentDone: tt.PercentDone,
		RepeatAfter: tt.RepeatAfter,
		RepeatMode:  tt.RepeatMode,
		ProjectID:   tft.TargetProjectID,
	}

	err = createTask(s, newTask, doer, false, true)
	if err != nil {
		return err
	}

	log.Debugf("Created task %d from template %d", newTask.ID, tft.TemplateID)

	// Apply labels from template
	if len(tt.LabelIDs) > 0 {
		for _, labelID := range tt.LabelIDs {
			lt := &LabelTask{
				TaskID:  newTask.ID,
				LabelID: labelID,
			}
			if _, err := s.Insert(lt); err != nil {
				log.Debugf("Could not add label %d to task %d: %v", labelID, newTask.ID, err)
				continue
			}
		}
		log.Debugf("Applied %d labels to task %d from template %d", len(tt.LabelIDs), newTask.ID, tft.TemplateID)
	}

	// Read back the full task to return it
	err = newTask.ReadOne(s, doer)
	if err != nil {
		return err
	}

	tft.Task = newTask

	return nil
}

// Stub methods to satisfy CRUDable interface
func (tft *TaskFromTemplate) ReadOne(_ *xorm.Session, _ web.Auth) error   { return nil }
func (tft *TaskFromTemplate) ReadAll(_ *xorm.Session, _ web.Auth, _ string, _ int, _ int) (interface{}, int, int64, error) {
	return nil, 0, 0, nil
}
func (tft *TaskFromTemplate) Update(_ *xorm.Session, _ web.Auth) error { return nil }
func (tft *TaskFromTemplate) Delete(_ *xorm.Session, _ web.Auth) error { return nil }
