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

	"code.vikunja.io/api/pkg/events"

	"xorm.io/builder"
	"xorm.io/xorm"

	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
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
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"-"`
	// The ID of the "base" task, the task which has a relation to another.
	TaskID int64 `xorm:"bigint not null" json:"task_id" param:"task"`
	// The ID of the other task, the task which is being related.
	OtherTaskID int64 `xorm:"bigint not null" json:"other_task_id" param:"otherTask"`
	// The kind of the relation.
	RelationKind RelationKind `xorm:"varchar(50) not null" json:"relation_kind" param:"relationKind"`

	CreatedByID int64 `xorm:"bigint not null" json:"-"`
	// The user who created this relation
	CreatedBy *user.User `xorm:"-" json:"created_by"`

	// A timestamp when this label was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`

	web.CRUDable    `xorm:"-" json:"-"`
	web.Permissions `xorm:"-" json:"-"`
}

// TableName holds the table name for the task relation table
func (*TaskRelation) TableName() string {
	return "task_relations"
}

// RelatedTaskMap holds all relations of a single task, grouped by relation kind.
// This avoids the need for an extra type TaskWithRelation (or similar).
type RelatedTaskMap map[RelationKind][]*Task

func getInverseRelation(kind RelationKind) RelationKind {
	switch kind {
	case RelationKindSubtask:
		return RelationKindParenttask
	case RelationKindParenttask:
		return RelationKindSubtask
	case RelationKindRelated:
		return RelationKindRelated
	case RelationKindDuplicateOf:
		return RelationKindDuplicates
	case RelationKindDuplicates:
		return RelationKindDuplicateOf
	case RelationKindBlocking:
		return RelationKindBlocked
	case RelationKindBlocked:
		return RelationKindBlocking
	case RelationKindPreceeds:
		return RelationKindFollows
	case RelationKindFollows:
		return RelationKindPreceeds
	case RelationKindCopiedFrom:
		return RelationKindCopiedTo
	case RelationKindCopiedTo:
		return RelationKindCopiedFrom
	case RelationKindUnknown:
		// Nothing to do
	}
	return RelationKindUnknown
}

func checkTaskRelationCycle(s *xorm.Session, relation *TaskRelation, otherTaskIDToCheck int64, visited map[int64]bool, currentPath map[int64]bool) (err error) {
	if visited == nil {
		visited = make(map[int64]bool)
	}

	if currentPath == nil {
		currentPath = make(map[int64]bool)
	}

	if visited[relation.TaskID] {
		return nil // Node already visited, no cycle detected
	}

	if relation.TaskID == otherTaskIDToCheck || // This checks for cycles between leaf nodes
		currentPath[relation.TaskID] ||
		currentPath[otherTaskIDToCheck] {
		// Cycle detected
		return ErrTaskRelationCycle{
			TaskID:      relation.TaskID,
			OtherTaskID: relation.OtherTaskID,
			Kind:        relation.RelationKind,
		}
	}

	visited[relation.TaskID] = true
	currentPath[relation.TaskID] = true

	parenttasks := []*TaskRelation{}
	// where child = relation.id
	err = s.Where("other_task_id = ? AND relation_kind = ?", relation.TaskID, relation.RelationKind).
		Find(&parenttasks)
	if err != nil {
		return
	}

	for _, parent := range parenttasks {
		err = checkTaskRelationCycle(s, parent, otherTaskIDToCheck, visited, currentPath)
		if err != nil {
			return err
		}
	}

	// Remove the current node from the currentPath to avoid false positives
	delete(currentPath, relation.TaskID)

	return nil
}

// Create creates a new task relation
// @Summary Create a new relation between two tasks
// @Description Creates a new relation between two tasks. The user needs to have update permissions on the base task and at least read permissions on the other task. Both tasks do not need to be on the same project. Take a look at the docs for available task relation kinds.
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param relation body models.TaskRelation true "The relation object"
// @Param taskID path int true "Task ID"
// @Success 201 {object} models.TaskRelation "The created task relation object."
// @Failure 400 {object} web.HTTPError "Invalid task relation object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/relations [put]
func (rel *TaskRelation) Create(s *xorm.Session, a web.Auth) error {

	// Check if both tasks are the same
	if rel.TaskID == rel.OtherTaskID {
		return ErrRelationTasksCannotBeTheSame{
			TaskID:      rel.TaskID,
			OtherTaskID: rel.OtherTaskID,
		}
	}

	// Check if the relation already exists, in one form or the other.
	exists, err := s.
		Where("(task_id = ? AND other_task_id = ? AND relation_kind = ?) OR (task_id = ? AND other_task_id = ? AND relation_kind = ?)",
			rel.TaskID, rel.OtherTaskID, rel.RelationKind, rel.TaskID, rel.OtherTaskID, rel.RelationKind).
		Exist(&TaskRelation{})
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

	rel.CreatedBy, err = GetUserOrLinkShareUser(s, a)
	if err != nil {
		return err
	}
	rel.CreatedByID = rel.CreatedBy.ID
	rel.ID = 0

	// Build up the other relation (see the comment above for explanation)
	otherRelation := &TaskRelation{
		TaskID:       rel.OtherTaskID,
		OtherTaskID:  rel.TaskID,
		CreatedByID:  rel.CreatedByID,
		RelationKind: getInverseRelation(rel.RelationKind),
	}

	// If we're creating a subtask relation, check if we're about to create a cycle
	if rel.RelationKind == RelationKindSubtask || rel.RelationKind == RelationKindParenttask {
		err = checkTaskRelationCycle(s, rel, rel.OtherTaskID, nil, nil)
		if err != nil {
			return err
		}
	}

	// Finally insert everything
	_, err = s.Insert(&[]*TaskRelation{
		rel,
		otherRelation,
	})
	if err != nil {
		return err
	}

	doer, _ := user.GetFromAuth(a)
	task, err := GetTaskByIDSimple(s, rel.TaskID)
	if err != nil {
		return err
	}

	return events.Dispatch(&TaskRelationCreatedEvent{
		Task:     &task,
		Relation: rel,
		Doer:     doer,
	})
}

// Delete removes a task relation
// @Summary Remove a task relation
// @tags task
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param relation body models.TaskRelation true "The relation object"
// @Param taskID path int true "Task ID"
// @Param relationKind path string true "The kind of the relation. See the TaskRelation type for more info."
// @Param otherTaskID path int true "The id of the other task."
// @Success 200 {object} models.Message "The task relation was successfully deleted."
// @Failure 400 {object} web.HTTPError "Invalid task relation object provided."
// @Failure 404 {object} web.HTTPError "The task relation was not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /tasks/{taskID}/relations/{relationKind}/{otherTaskID} [delete]
func (rel *TaskRelation) Delete(s *xorm.Session, a web.Auth) error {

	cond := builder.Or(
		builder.And(
			builder.Eq{"task_id": rel.TaskID},
			builder.Eq{"other_task_id": rel.OtherTaskID},
			builder.Eq{"relation_kind": rel.RelationKind},
		),
		builder.And(
			builder.Eq{"task_id": rel.OtherTaskID},
			builder.Eq{"other_task_id": rel.TaskID},
			builder.Eq{"relation_kind": getInverseRelation(rel.RelationKind)},
		),
	)

	// Check if the relation exists
	exists, err := s.
		Where(cond).
		Exist(&TaskRelation{})
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

	_, err = s.
		Where(cond).
		Delete(&TaskRelation{})
	if err != nil {
		return err
	}

	doer, _ := user.GetFromAuth(a)
	task, err := GetTaskByIDSimple(s, rel.TaskID)
	if err != nil {
		return err
	}

	return events.Dispatch(&TaskRelationDeletedEvent{
		Task:     &task,
		Relation: rel,
		Doer:     doer,
	})
}
