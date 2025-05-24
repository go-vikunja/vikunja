// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

type TaskDuplicate struct {
	// The task id of the task to duplicate
	TaskID    int64 `json:"-" param:"taskid"`
	ProjectID int64 `json:"-" param:"projectid"`

	// The copied task
	Task *Task `json:"duplicated_task,omitempty"`

	web.Rights   `json:"-"`
	web.CRUDable `json:"-"`
}

// Create duplicates of a task
// @Summary Duplicate an existing task
// @Description Copies the task, assignees, lables, subtasks from one task to a new one. The user needs read and write access in the project.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param taskID path int true "The task ID to duplicate"
// @Param projectID path int true "The project ID of task to duplicate"
// @Success 201 {object} models.TaskDuplicate "The duplicated task."
// @Failure 400 {object} web.HTTPError "Invalid task duplicate object provided."
// @Failure 403 {object} web.HTTPError "The user does not have access to the project."
// @Failure 500 {object} models.Message "Internal error"
// @Router /projects/{projectID}tasks/{taskID}/duplicate [put]
func (td *TaskDuplicate) Create(s *xorm.Session, doer web.Auth) (err error) {
	idMap := map[int64]*Task{}
	// Start duplicating from the root task, do not copy parent relation in recursion
	root, err := duplicateTask(s, doer, td.TaskID, idMap)
	if err != nil {
		return err
	}
	td.Task = root

	// Copy follows/precedes relations for all duplicated tasks, only one type, corresponding will be created too
	for oldID, newTask := range idMap {
		rels := []*TaskRelation{}
		err := s.Where("task_id = ? AND relation_kind = ? ", oldID, RelationKindFollows).Find(&rels)
		if err != nil {
			return err
		}
		for _, rel := range rels {
			// Only copy if the other task is also duplicated
			if other, ok := idMap[rel.OtherTaskID]; ok {
				tr := &TaskRelation{
					TaskID:       newTask.ID,
					OtherTaskID:  other.ID,
					RelationKind: rel.RelationKind,
				}
				if err := tr.Create(s, doer); err != nil {
					return err
				}
			}
		}
	}

	// If the topmost task has a parent, copy that relation to the new task
	parentRel := &TaskRelation{}
	hasParent, err := s.Where("other_task_id = ? AND relation_kind = ?", td.TaskID, RelationKindSubtask).Get(parentRel)
	if err != nil {
		return err
	}
	if hasParent {
		tr := &TaskRelation{
			TaskID:       parentRel.TaskID, // original parent
			OtherTaskID:  root.ID,          // new duplicated task
			RelationKind: RelationKindSubtask,
		}
		if err := tr.Create(s, doer); err != nil {
			return err
		}
	}

	return nil
}

// duplicateTask duplicates a task and its subtasks recursively, copying reminders and relations.
func duplicateTask(s *xorm.Session, doer web.Auth, origID int64, idMap map[int64]*Task) (*Task, error) {

	// Get the original task with all info
	origTask, err := GetTaskByIDSimple(s, origID)
	if err != nil {
		return nil, err
	}
	err = addMoreInfoToTasks(s, map[int64]*Task{origTask.ID: &origTask}, doer, nil, nil)
	if err != nil {
		return nil, err
	}

	// Copy the task
	newTask := &Task{
		Title:       origTask.Title,
		Description: origTask.Description,
		Done:        origTask.Done,
		DoneAt:      origTask.DoneAt,
		DueDate:     origTask.DueDate,
		ProjectID:   origTask.ProjectID,
		RepeatAfter: origTask.RepeatAfter,
		RepeatMode:  origTask.RepeatMode,
		Priority:    origTask.Priority,
		StartDate:   origTask.StartDate,
		EndDate:     origTask.EndDate,
		Assignees:   origTask.Assignees,
		Labels:      origTask.Labels,
		HexColor:    origTask.HexColor,
		PercentDone: origTask.PercentDone,
	}

	// Copy reminders
	reminders, _ := getRemindersForTasks(s, []int64{origTask.ID})
	if len(reminders) > 0 {
		newReminders := make([]*TaskReminder, len(reminders))
		for i, r := range reminders {
			newReminders[i] = &TaskReminder{
				Reminder:       r.Reminder,
				RelativeTo:     r.RelativeTo,
				RelativePeriod: r.RelativePeriod,
			}
		}
		newTask.Reminders = newReminders
	}

	// Insert the new task
	err = newTask.Create(s, doer)
	if err != nil {
		return nil, err
	}
	idMap[origTask.ID] = newTask

	// Duplicate subtasks recursively
	relations := []*TaskRelation{}
	err = s.Where("task_id = ? AND relation_kind = ?", origTask.ID, RelationKindSubtask).Find(&relations)
	if err != nil {
		return nil, err
	}
	for _, rel := range relations {
		child, err := duplicateTask(s, doer, rel.OtherTaskID, idMap)
		if err != nil {
			return nil, err
		}

		// Create subtask relation (newTask.ID -> child.ID)
		tr := &TaskRelation{
			TaskID:       newTask.ID,
			OtherTaskID:  child.ID,
			RelationKind: RelationKindSubtask,
		}
		if err := tr.Create(s, doer); err != nil {
			return nil, err
		}
	}
	return newTask, nil
}

// CanCreate checks if a user has the right to duplicate a task
func (td *TaskDuplicate) CanCreate(s *xorm.Session, a web.Auth) (canCreate bool, err error) {
	project := &Project{ID: td.ProjectID}
	return project.CanWrite(s, a)
}
