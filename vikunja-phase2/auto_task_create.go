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
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// CheckAndCreateAutoTasks scans all active auto-task templates for a user
// and creates task instances for any that are due.
// Rules:
//   - Only one open (not done) instance per template can exist at a time
//   - If the previous instance isn't done, no new one is created (it just goes overdue)
//   - next_due_at is recalculated from the last completion time, not creation time
func CheckAndCreateAutoTasks(s *xorm.Session, u *user.User) ([]*Task, error) {
	now := time.Now()

	// Find all active templates where next_due_at <= now
	templates := []*AutoTaskTemplate{}
	err := s.Where("owner_id = ? AND active = ? AND next_due_at <= ?", u.ID, true, now).Find(&templates)
	if err != nil {
		return nil, err
	}

	created := make([]*Task, 0)

	for _, tmpl := range templates {
		// Check if end_date has passed
		if tmpl.EndDate != nil && now.After(*tmpl.EndDate) {
			tmpl.Active = false
			_, _ = s.ID(tmpl.ID).Cols("active").Update(tmpl)
			continue
		}

		// Check: does an open (not done) task already exist for this template?
		// If so, skip — the user needs to complete it first.
		openCount, err := s.Where("auto_template_id = ? AND done = ?", tmpl.ID, false).Count(&Task{})
		if err != nil {
			return nil, err
		}
		if openCount > 0 {
			continue
		}

		// No open task exists. Before creating a new one, check if a task was
		// recently completed that we haven't processed yet (OnAutoTaskCompleted
		// isn't hooked into the task update path). Find the most recently
		// completed task for this template and advance next_due_at from its
		// completion time.
		type doneInfo struct {
			ID     int64     `xorm:"'id'"`
			DoneAt time.Time `xorm:"'done_at'"`
		}
		lastDone := &doneInfo{}
		hasDone, err := s.SQL(
			"SELECT id, done_at FROM tasks WHERE auto_template_id = ? AND done = ? AND done_at IS NOT NULL ORDER BY done_at DESC LIMIT 1",
			tmpl.ID, true,
		).Get(lastDone)
		if err != nil {
			return nil, err
		}

		if hasDone && !lastDone.DoneAt.IsZero() {
			// Compare: was this task completed AFTER the template's last_completed_at?
			// If so, we need to advance next_due_at first.
			needsAdvance := false
			if tmpl.LastCompletedAt == nil {
				needsAdvance = true
			} else if lastDone.DoneAt.After(*tmpl.LastCompletedAt) {
				needsAdvance = true
			}

			if needsAdvance {
				completedAt := lastDone.DoneAt
				tmpl.LastCompletedAt = &completedAt
				nextDue := advanceFromTime(completedAt, tmpl.IntervalValue, tmpl.IntervalUnit)
				tmpl.NextDueAt = &nextDue
				_, _ = s.ID(tmpl.ID).Cols("last_completed_at", "next_due_at").Update(tmpl)

				// Log the completion event
				_, _ = s.Insert(&AutoTaskLog{
					TemplateID:  tmpl.ID,
					TaskID:      lastDone.ID,
					TriggerType: "completed",
				})

				// After advancing, check if the NEW next_due_at is still in the past.
				// If not, this template isn't due yet — skip it.
				if nextDue.After(now) {
					continue
				}
			}
		}

		task, err := createAutoTaskInstance(s, tmpl, u, "system")
		if err != nil {
			return nil, err
		}
		if task != nil {
			created = append(created, task)
		}
	}

	return created, nil
}

// TriggerAutoTask manually creates a task from an auto-task template immediately,
// regardless of its schedule. Respects the "one open instance" rule.
func TriggerAutoTask(s *xorm.Session, templateID int64, u *user.User) (*Task, error) {
	tmpl := &AutoTaskTemplate{}
	has, err := s.Where("id = ? AND owner_id = ?", templateID, u.ID).Get(tmpl)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrAutoTaskTemplateNotFound{ID: templateID}
	}

	// Check for existing open instance
	openCount, err := s.Where("auto_template_id = ? AND done = ?", tmpl.ID, false).Count(&Task{})
	if err != nil {
		return nil, err
	}
	if openCount > 0 {
		return nil, fmt.Errorf("an open task already exists for this template — complete it first")
	}

	return createAutoTaskInstance(s, tmpl, u, "manual")
}

// OnAutoTaskCompleted should be called when a task with auto_template_id is marked done.
// It recalculates the next_due_at based on the completion time.
func OnAutoTaskCompleted(s *xorm.Session, task *Task) error {
	// Check if this task has an auto_template_id via SQL
	var autoTemplateID int64
	has, err := s.SQL("SELECT auto_template_id FROM tasks WHERE id = ? AND auto_template_id IS NOT NULL AND auto_template_id > 0", task.ID).Get(&autoTemplateID)
	if err != nil || !has || autoTemplateID == 0 {
		return err
	}

	tmpl := &AutoTaskTemplate{}
	has, err = s.ID(autoTemplateID).Get(tmpl)
	if err != nil || !has {
		return err
	}

	now := time.Now()
	tmpl.LastCompletedAt = &now

	// Calculate next due from NOW (completion time), not from the original due date.
	// This prevents task pile-up if the user was late completing.
	nextDue := advanceFromTime(now, tmpl.IntervalValue, tmpl.IntervalUnit)
	tmpl.NextDueAt = &nextDue

	_, err = s.ID(tmpl.ID).Cols("last_completed_at", "next_due_at").Update(tmpl)
	return err
}

// createAutoTaskInstance handles the actual task creation, label/assignee assignment,
// and logging for both auto-check and manual triggers.
func createAutoTaskInstance(s *xorm.Session, tmpl *AutoTaskTemplate, u *user.User, triggerType string) (*Task, error) {
	// Determine the target project
	projectID := tmpl.ProjectID
	if projectID == 0 {
		projectID = u.DefaultProjectID
		if projectID == 0 {
			// Fallback: find first non-archived project
			inbox := &Project{}
			has, err := s.Where("owner_id = ? AND is_archived = ?", u.ID, false).
				OrderBy("id ASC").Limit(1).Get(inbox)
			if err != nil {
				return nil, err
			}
			if !has {
				return nil, fmt.Errorf("no project found for user %d", u.ID)
			}
			projectID = inbox.ID
		}
	}

	// Build and create the task
	dueDate := time.Now()
	if tmpl.NextDueAt != nil {
		dueDate = *tmpl.NextDueAt
	}

	task := &Task{
		Title:          tmpl.Title,
		Description:    tmpl.Description,
		Priority:       tmpl.Priority,
		HexColor:       tmpl.HexColor,
		ProjectID:      projectID,
		DueDate:        dueDate,
		AutoTemplateID: tmpl.ID,
	}

	err := createTask(s, task, u, false, false)
	if err != nil {
		return nil, fmt.Errorf("auto-task create failed for template %d: %w", tmpl.ID, err)
	}

	// Ensure auto_template_id is persisted (createTask may not include it in default cols)
	_, _ = s.Exec("UPDATE tasks SET auto_template_id = ? WHERE id = ?", tmpl.ID, task.ID)

	// Add labels
	for _, labelID := range tmpl.LabelIDs {
		lt := &LabelTask{LabelID: labelID, TaskID: task.ID}
		_, _ = s.Insert(lt)
	}

	// Add assignees
	for _, assigneeID := range tmpl.AssigneeIDs {
		ta := &TaskAssginee{TaskID: task.ID, UserID: assigneeID}
		_, _ = s.Insert(ta)
	}

	// Copy attachments from template to the new task
	copyAutoTaskTemplateAttachments(s, tmpl.ID, task.ID, u)

	// Update template tracking
	nowTime := time.Now()
	tmpl.LastCreatedAt = &nowTime
	_, _ = s.ID(tmpl.ID).Cols("last_created_at").Update(tmpl)

	// Log the generation event
	triggeredBy := int64(0)
	if triggerType == "manual" {
		triggeredBy = u.ID
	}
	logEntry := &AutoTaskLog{
		TemplateID:    tmpl.ID,
		TaskID:        task.ID,
		TriggerType:   triggerType,
		TriggeredByID: triggeredBy,
	}
	_, _ = s.Insert(logEntry)

	return task, nil
}

// advanceFromTime calculates the next due date by adding one interval to the given time.
func advanceFromTime(from time.Time, intervalValue int, intervalUnit string) time.Time {
	switch intervalUnit {
	case "hours":
		return from.Add(time.Duration(intervalValue) * time.Hour)
	case "weeks":
		return from.AddDate(0, 0, intervalValue*7)
	case "months":
		return from.AddDate(0, intervalValue, 0)
	default: // days
		return from.AddDate(0, 0, intervalValue)
	}
}

// TriggerAutoTaskFromAuth is a convenience wrapper that resolves the user from web.Auth.
func TriggerAutoTaskFromAuth(s *xorm.Session, templateID int64, auth web.Auth) (*Task, error) {
	u := auth.(*user.User)
	return TriggerAutoTask(s, templateID, u)
}

// CheckAutoTasksFromAuth is a convenience wrapper that resolves the user from web.Auth.
func CheckAutoTasksFromAuth(s *xorm.Session, auth web.Auth) ([]*Task, error) {
	u := auth.(*user.User)
	return CheckAndCreateAutoTasks(s, u)
}

// copyAutoTaskTemplateAttachments copies all file attachments from an auto-task
// template to a newly generated task. Each file is duplicated in storage so that
// the template and task attachments are independent.
func copyAutoTaskTemplateAttachments(s *xorm.Session, templateID, taskID int64, u *user.User) {
	templateAttachments := make([]*AutoTaskTemplateAttachment, 0)
	err := s.Where("template_id = ?", templateID).Find(&templateAttachments)
	if err != nil {
		log.Errorf("[Auto-Task] Could not load template attachments for template %d: %s", templateID, err)
		return
	}

	if len(templateAttachments) == 0 {
		return
	}

	for _, tmplAttach := range templateAttachments {
		// Load the source file metadata
		srcFile := &files.File{ID: tmplAttach.FileID}
		if err := srcFile.LoadFileMetaByID(); err != nil {
			log.Debugf("[Auto-Task] Skipping attachment %d: file %d metadata not found: %s", tmplAttach.ID, tmplAttach.FileID, err)
			continue
		}

		// Load the actual file content
		if err := srcFile.LoadFileByID(); err != nil {
			log.Debugf("[Auto-Task] Skipping attachment %d: could not load file %d: %s", tmplAttach.ID, tmplAttach.FileID, err)
			continue
		}

		// Create a new task attachment by duplicating the file
		newAttachment := &TaskAttachment{
			TaskID: taskID,
		}
		err := newAttachment.NewAttachment(s, srcFile.File, srcFile.Name, srcFile.Size, u)
		if err != nil {
			log.Errorf("[Auto-Task] Could not copy attachment %d to task %d: %s", tmplAttach.ID, taskID, err)
		}

		if srcFile.File != nil {
			_ = srcFile.File.Close()
		}

		log.Debugf("[Auto-Task] Copied attachment '%s' (file %d) to task %d as attachment %d",
			srcFile.Name, tmplAttach.FileID, taskID, newAttachment.ID)
	}
}
