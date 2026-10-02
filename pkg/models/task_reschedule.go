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
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// maxRescheduleSteps bounds the walk over successors. A dependency graph without cycles needs far
// fewer; reaching the bound means there is one, and the walk stops instead of looping.
const maxRescheduleSteps = 20000

// RescheduleResult is what moving a task did.
type RescheduleResult struct {
	// Task is the task with its new dates.
	Task *Task
	// Changed are the successors that were pushed later, with their new dates.
	Changed []*Task
	// Skipped are the ids of successors that had to move but the caller may not write.
	Skipped []int64
}

// writeTaskDates updates the dates and nothing else. updateSingleTask is not used on purpose: it
// replaces the assignees with the ones of the task it is given.
func writeTaskDates(s *xorm.Session, taskID int64, start, end time.Time) error {
	_, err := s.ID(taskID).Cols("start_date", "end_date").Update(&Task{StartDate: start, EndDate: end})
	return err
}

// RescheduleTask moves a task and pushes the tasks that depend on it.
//
// Dependencies are finish-to-start: when A precedes B, B cannot start before A ends. If the new end
// of A is later than the start of B, B moves so that it starts when A ends and keeps its duration,
// and so on down the chain. It never pulls a successor earlier: moving a task earlier leaves its
// successors where they are. Tasks without dates, done tasks and successors the caller cannot write
// are not moved (the latter are reported). A milestone has no duration: it sits at its end date.
//
// It does not commit; the caller owns the transaction and dispatches the queued events.
func RescheduleTask(s *xorm.Session, a web.Auth, taskID int64, start, end time.Time) (*RescheduleResult, error) {
	task, err := GetTaskByIDSimple(s, taskID)
	if err != nil {
		return nil, err
	}

	can, err := (&Task{ID: taskID}).CanUpdate(s, a)
	if err != nil {
		return nil, err
	}
	if !can {
		return nil, ErrGenericForbidden{}
	}

	if !start.IsZero() && !end.IsZero() && end.Before(start) {
		return nil, ErrInvalidData{Message: "the end date must not be before the start date"}
	}
	if task.IsMilestone {
		if end.IsZero() {
			return nil, ErrInvalidData{Message: "a milestone needs an end date"}
		}
		start = end
	}

	err = writeTaskDates(s, task.ID, start, end)
	if err != nil {
		return nil, err
	}
	task.StartDate, task.EndDate = start, end
	doer := doerFromAuth(s, a)
	events.DispatchOnCommit(s, &TaskUpdatedEvent{Task: &task, Doer: doer})

	result := &RescheduleResult{Task: &task}
	if end.IsZero() {
		// Without a finish there is nothing for the successors to wait for.
		return result, nil
	}

	type step struct {
		id  int64
		end time.Time
	}
	queue := []step{{id: task.ID, end: end}}
	loaded := map[int64]*Task{task.ID: &task}
	writable := map[int64]bool{}
	moved := map[int64]*Task{}
	skipped := map[int64]bool{}

	canWriteProject := func(projectID int64) (bool, error) {
		if ok, seen := writable[projectID]; seen {
			return ok, nil
		}
		// A successor that cannot be written, whatever the reason, is skipped and reported. It must not
		// fail the move of everything else.
		project, projectErr := GetProjectSimpleByID(s, projectID)
		if projectErr != nil {
			if IsErrProjectDoesNotExist(projectErr) {
				writable[projectID] = false
				return false, nil
			}
			return false, projectErr
		}
		// CanWrite answers (true, ErrProjectIsArchived) for the owner of an archived project: it can be
		// written to in principle, but not while it is archived.
		ok, permErr := project.CanWrite(s, a)
		if permErr != nil {
			if IsErrProjectIsArchived(permErr) {
				writable[projectID] = false
				return false, nil
			}
			return false, permErr
		}
		writable[projectID] = ok
		return ok, nil
	}

	for steps := 0; len(queue) > 0; steps++ {
		if steps > maxRescheduleSteps {
			return nil, ErrTaskRelationCycle{TaskID: task.ID, OtherTaskID: task.ID, Kind: RelationKindPreceeds}
		}
		cur := queue[0]
		queue = queue[1:]

		var relations []*TaskRelation
		err = s.Where("task_id = ? AND relation_kind = ?", cur.id, RelationKindPreceeds).Find(&relations)
		if err != nil {
			return nil, err
		}

		for _, rel := range relations {
			succ, ok := loaded[rel.OtherTaskID]
			if !ok {
				var t Task
				t, err = GetTaskByIDSimple(s, rel.OtherTaskID)
				if err != nil {
					if IsErrTaskDoesNotExist(err) {
						continue
					}
					return nil, err
				}
				succ = &t
				loaded[succ.ID] = succ
			}

			newStart, newEnd, shift := successorDates(succ, cur.end)
			if !shift || succ.Done {
				continue
			}

			var canWrite bool
			canWrite, err = canWriteProject(succ.ProjectID)
			if err != nil {
				return nil, err
			}
			if !canWrite {
				skipped[succ.ID] = true
				continue
			}

			err = writeTaskDates(s, succ.ID, newStart, newEnd)
			if err != nil {
				return nil, err
			}
			succ.StartDate, succ.EndDate = newStart, newEnd
			moved[succ.ID] = succ
			events.DispatchOnCommit(s, &TaskUpdatedEvent{Task: succ, Doer: doer})

			if !newEnd.IsZero() {
				queue = append(queue, step{id: succ.ID, end: newEnd})
			}
		}
	}

	for _, t := range moved {
		result.Changed = append(result.Changed, t)
	}
	for id := range skipped {
		result.Skipped = append(result.Skipped, id)
	}
	return result, nil
}

// successorDates decides whether a successor has to move for a predecessor that ends at predEnd,
// and to where. Tasks without a start (a milestone: without an end) are left alone, and so is a
// successor that already starts at or after the predecessor's end.
func successorDates(succ *Task, predEnd time.Time) (start, end time.Time, shift bool) {
	if succ.IsMilestone {
		if succ.EndDate.IsZero() || !succ.EndDate.Before(predEnd) {
			return time.Time{}, time.Time{}, false
		}
		return predEnd, predEnd, true
	}

	if succ.StartDate.IsZero() || !succ.StartDate.Before(predEnd) {
		return time.Time{}, time.Time{}, false
	}
	if succ.EndDate.IsZero() {
		return predEnd, time.Time{}, true
	}
	return predEnd, predEnd.Add(succ.EndDate.Sub(succ.StartDate)), true
}
