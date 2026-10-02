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
	"context"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

// ImportAssignedRow is one task of an import that names its project, assignee and assignor.
type ImportAssignedRow struct {
	// Number is the row number in the file, for the report.
	Number int
	// Task carries title, description, dates, priority, done, labels and reminders.
	Task Task
	// AssigneeEmail is optional. Without it the task has no assignee, except in the importer's own
	// inbox, where the importer becomes the assignee.
	AssigneeEmail string
	// AssignorEmail is optional, see importAssignedRow for the defaults.
	AssignorEmail string
	// ProjectID is the id of an existing project. Without it the task goes to the assignee's
	// inbox, or the importer's own inbox when there is no other assignee.
	ProjectID int64
}

// ImportRowResult is the outcome of one row.
type ImportRowResult struct {
	Number int
	TaskID int64
	Err    error
}

// ImportAssignedRows imports the rows one by one, each in its own transaction, so a rejected row is
// reported and the others still go in. The importer is read from the database, never trusted from
// the caller.
//
// Instance admins and bots an admin owns may put tasks into other people's inboxes and name an
// assignor other than themselves. For everybody else a row that needs either is rejected, a defaulted
// assignor falls back to the importer (which is not a claim about anybody else), and people who are
// not discoverable by email are reported as not found.
func ImportAssignedRows(ctx context.Context, actor *user.User, rows []*ImportAssignedRow) []ImportRowResult {
	results := make([]ImportRowResult, 0, len(rows))
	for _, row := range rows {
		res := ImportRowResult{Number: row.Number}
		res.TaskID, res.Err = importAssignedRowInOwnTransaction(ctx, actor, row)
		results = append(results, res)
	}
	return results
}

func importAssignedRowInOwnTransaction(ctx context.Context, actor *user.User, row *ImportAssignedRow) (int64, error) {
	s := db.NewSession()
	defer s.Close()
	defer events.CleanupPending(s)

	taskID, err := importAssignedRow(s, actor.ID, row)
	if err != nil {
		_ = s.Rollback()
		return 0, err
	}
	err = s.Commit()
	if err != nil {
		return 0, err
	}
	events.DispatchPending(ctx, s)
	return taskID, nil
}

func importAssignedRow(s *xorm.Session, actorID int64, row *ImportAssignedRow) (int64, error) {
	actor, err := user.GetUserByID(s, actorID)
	if err != nil {
		return 0, err
	}
	privileged, err := isPrivilegedActor(s, actor)
	if err != nil {
		return 0, err
	}

	var assignee *user.User
	if strings.TrimSpace(row.AssigneeEmail) != "" {
		assignee, err = resolveAssigneeByEmail(s, row.AssigneeEmail)
		if err != nil {
			return 0, err
		}
		// Anybody else may only find the people who allow to be found by email, like in the
		// sharing dialogs. Otherwise an import would tell them who has an account.
		if hiddenFromImporter(actor, assignee, privileged) {
			return 0, ErrAssigneeEmailNotFound{}
		}
	}

	var project *Project
	switch {
	case row.ProjectID != 0:
		project, err = GetProjectSimpleByID(s, row.ProjectID)
		if err != nil {
			return 0, err
		}
		var canWrite bool
		canWrite, err = project.CanWrite(s, actor)
		if err != nil {
			return 0, err
		}
		if !canWrite {
			return 0, ErrGenericForbidden{}
		}
	case assignee != nil && assignee.ID != actor.ID:
		// Somebody else's inbox is not the importer's to write into.
		if !privileged {
			return 0, ErrGenericForbidden{}
		}
		project, err = assigneeInbox(s, assignee)
	default:
		project, err = assigneeInbox(s, actor)
	}
	if err != nil {
		return 0, err
	}

	assignor, err := importAssignor(s, actor, privileged, row, assignee, project)
	if err != nil {
		return 0, err
	}

	task := row.Task
	labels := task.Labels
	task.Labels = nil
	task.ProjectID = project.ID
	task.Assignees = nil
	if assignee != nil {
		task.Assignees = []*user.User{assignee}
	}
	if assignor.ID != actor.ID {
		task.assignorOverride = assignor
	}

	err = createTask(s, &task, actor, true, true)
	if err != nil {
		return 0, err
	}

	err = attachImportLabels(s, actor, task.ID, labels)
	if err != nil {
		return 0, err
	}
	return task.ID, nil
}

// hiddenFromImporter reports whether an importer without special rights must not learn that the
// user exists: somebody else who did not make themselves discoverable by email.
func hiddenFromImporter(actor, found *user.User, privileged bool) bool {
	return !privileged && found.ID != actor.ID && !found.DiscoverableByEmail
}

// importAssignor decides who the task is stored as created by. An explicit email wins. Otherwise a
// privileged importer gets the project owner (with a project) or the assignee (without one), and
// anybody else gets themselves.
func importAssignor(s *xorm.Session, actor *user.User, privileged bool, row *ImportAssignedRow, assignee *user.User, project *Project) (*user.User, error) {
	if strings.TrimSpace(row.AssignorEmail) != "" {
		in := &AssignTaskByEmailInput{AssignorEmail: row.AssignorEmail}
		explicit, err := resolveAssignor(s, in, nil, nil)
		if err != nil {
			return nil, err
		}
		if hiddenFromImporter(actor, explicit, privileged) {
			return nil, ErrAssignorNotFound{}
		}
		if !privileged && explicit.ID != actor.ID {
			return nil, ErrGenericForbidden{}
		}
		return explicit, nil
	}

	if !privileged {
		return actor, nil
	}
	if row.ProjectID != 0 {
		return resolveAssignor(s, &AssignTaskByEmailInput{ProjectID: row.ProjectID}, assignee, project)
	}
	if assignee != nil {
		return assignee, nil
	}
	return actor, nil
}

// attachImportLabels gives the task its labels, creating missing ones for the importer. The
// label task rows are written directly: LabelTask.Create asks for write access to the task, which an
// importer filling somebody else's inbox does not have.
func attachImportLabels(s *xorm.Session, actor *user.User, taskID int64, labels []*Label) error {
	for _, wanted := range labels {
		if wanted == nil || strings.TrimSpace(wanted.Title) == "" {
			continue
		}
		title := strings.TrimSpace(wanted.Title)

		label := &Label{}
		has, err := s.Where("created_by_id = ? AND title = ?", actor.ID, title).Get(label)
		if err != nil {
			return err
		}
		if !has {
			label = &Label{Title: title, HexColor: wanted.HexColor}
			err = label.Create(s, actor)
			if err != nil {
				return err
			}
		}

		exists, err := s.Exist(&LabelTask{LabelID: label.ID, TaskID: taskID})
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		_, err = s.Insert(&LabelTask{LabelID: label.ID, TaskID: taskID})
		if err != nil {
			return err
		}
	}
	return nil
}
