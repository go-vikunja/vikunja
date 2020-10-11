// Copyright 2018-2020 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"time"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
)

// RelationKind represents a kind of relation between to tasks
type RelationKind string

// All valid relation kinds
const (
	RelationKindUnknown     RelationKind = `unknown`
	RelationKindSubtask     RelationKind = `subtask`
	RelationKindParenttask  RelationKind = `parenttask`
	RelationKindRelated     RelationKind = `related`
	RelationKindDuplicateOf RelationKind = `duplicateof`
	RelationKindDuplicates  RelationKind = `duplicates`
	RelationKindBlocking    RelationKind = `blocking`
	RelationKindBlocked     RelationKind = `blocked`
	RelationKindPreceeds    RelationKind = `precedes`
	RelationKindFollows     RelationKind = `follows`
	RelationKindCopiedFrom  RelationKind = `copiedfrom`
	RelationKindCopiedTo    RelationKind = `copiedto`
)

/*
 * The direction of the relation goes _from_ task_id -> other_task_id.
 * The relation kind only tells us something about the relation in that direction, and NOT
 * the other way around. This means each relation exists two times in the db, one for each
 * relevant direction.
 * This design allows to easily do things like "Give me every relation for this task" whithout having
 * to deal with each possible case of relation. Instead, it would just give me every relation record
 * which has task_id set to the task ID I care about.
 *
 * For example, when I create a relation where I define task 2 as a subtask of task 1, it would actually
 * create two relations. One from Task 2 -> Task 1 with relation kind subtask and one from Task 1 -> Task 2
 * with relation kind parent task.
 * When I now want to have all relations task 1 is a part of, I just ask "Give me all relations where
 * task_id = 1".
 */

func (rk RelationKind) isValid() bool {
	return rk == RelationKindSubtask ||
		rk == RelationKindParenttask ||
		rk == RelationKindRelated ||
		rk == RelationKindDuplicateOf ||
		rk == RelationKindDuplicates ||
		rk == RelationKindBlocked ||
		rk == RelationKindBlocking ||
		rk == RelationKindPreceeds ||
		rk == RelationKindFollows ||
		rk == RelationKindCopiedFrom ||
		rk == RelationKindCopiedTo
}

// TaskRelation represents a kind of relation between two tasks
type TaskRelation struct {
	// The unique, numeric id of this relation.
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"-"`
	// The ID of the "base" task, the task which has a relation to another.
	TaskID int64 `xorm:"int(11) not null" json:"task_id" param:"task"`
	// The ID of the other task, the task which is being related.
	OtherTaskID int64 `xorm:"int(11) not null" json:"other_task_id"`
	// The kind of the relation.
	RelationKind RelationKind `xorm:"varchar(50) not null" json:"relation_kind"`

	CreatedByID int64 `xorm:"int(11) not null" json:"-"`
	// The user who created this relation
	CreatedBy *user.User `xorm:"-" json:"created_by"`

	// A timestamp when this label was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`

	web.CRUDable `xorm:"-" json:"-"`
	web.Rights   `xorm:"-" json:"-"`
}

// TableName holds the table name for the task relation table
func (TaskRelation) TableName() string {
	return "task_relations"
}

// RelatedTaskMap holds all relations of a single task, grouped by relation kind.
// This avoids the need for an extra type TaskWithRelation (or similar).
type RelatedTaskMap map[RelationKind][]*Task

// Create creates a new task relation
// @Summary Create a new relation between two tasks
// @Description Creates a new relation between two tasks. The user needs to have update rights on the base task and at least read rights on the other task. Both tasks do not need to be on the same list. Take a look at the docs for available task relation kinds.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param relation body models.TaskRelation true "The relation object"
// @Param taskID path int true "Task ID"
// @Success 200 {object} models.TaskRelation "The created task relation object."
// @Failure 400 {object} web.HTTPError "Invalid task relation object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/relations [put]
func (rel *TaskRelation) Create(a web.Auth) error {

	// Check if both tasks are the same
	if rel.TaskID == rel.OtherTaskID {
		return ErrRelationTasksCannotBeTheSame{
			TaskID:      rel.TaskID,
			OtherTaskID: rel.OtherTaskID,
		}
	}

	// Check if the relation already exists, in one form or the other.
	exists, err := x.
		Where("(task_id = ? AND other_task_id = ? AND relation_kind = ?) OR (task_id = ? AND other_task_id = ? AND relation_kind = ?)",
			rel.TaskID, rel.OtherTaskID, rel.RelationKind, rel.TaskID, rel.OtherTaskID, rel.RelationKind).
		Exist(rel)
	if err != nil {
		return err
	}
	if exists {
		return ErrRelationAlreadyExists{
			TaskID:      rel.TaskID,
			OtherTaskID: rel.OtherTaskID,
			Kind:        rel.RelationKind,
		}
	}

	rel.CreatedByID = a.GetID()

	// Build up the other relation (see the comment above for explanation)
	otherRelation := &TaskRelation{
		TaskID:      rel.OtherTaskID,
		OtherTaskID: rel.TaskID,
		CreatedByID: a.GetID(),
	}

	switch rel.RelationKind {
	case RelationKindSubtask:
		otherRelation.RelationKind = RelationKindParenttask
	case RelationKindParenttask:
		otherRelation.RelationKind = RelationKindSubtask
	case RelationKindRelated:
		otherRelation.RelationKind = RelationKindRelated
	case RelationKindDuplicateOf:
		otherRelation.RelationKind = RelationKindDuplicates
	case RelationKindDuplicates:
		otherRelation.RelationKind = RelationKindDuplicateOf
	case RelationKindBlocking:
		otherRelation.RelationKind = RelationKindBlocked
	case RelationKindBlocked:
		otherRelation.RelationKind = RelationKindBlocking
	case RelationKindPreceeds:
		otherRelation.RelationKind = RelationKindFollows
	case RelationKindFollows:
		otherRelation.RelationKind = RelationKindPreceeds
	case RelationKindCopiedFrom:
		otherRelation.RelationKind = RelationKindCopiedTo
	case RelationKindCopiedTo:
		otherRelation.RelationKind = RelationKindCopiedFrom
	case RelationKindUnknown:
		// Nothing to do
	}

	// Finally insert everything
	_, err = x.Insert(&[]*TaskRelation{
		rel,
		otherRelation,
	})
	return err
}

// Delete removes a task relation
// @Summary Remove a task relation
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param relation body models.TaskRelation true "The relation object"
// @Param taskID path int true "Task ID"
// @Success 200 {object} models.Message "The task relation was successfully deleted."
// @Failure 400 {object} web.HTTPError "Invalid task relation object provided."
// @Failure 404 {object} web.HTTPError "The task relation was not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/relations [delete]
func (rel *TaskRelation) Delete() error {
	// Check if the relation exists
	exists, err := x.
		Cols("task_id", "other_task_id", "relation_kind").
		Get(rel)
	if err != nil {
		return err
	}
	if !exists {
		return ErrRelationDoesNotExist{
			TaskID:      rel.TaskID,
			OtherTaskID: rel.OtherTaskID,
			Kind:        rel.RelationKind,
		}
	}

	_, err = x.Delete(rel)
	return err
}
