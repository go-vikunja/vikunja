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
	"code.vikunja.io/api/pkg/web"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// CanUpdate checks if a user can update a label
func (l *Label) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return l.isLabelOwner(s, a) // Only owners should be allowed to update a label
}

// CanDelete checks if a user can delete a label
func (l *Label) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return l.isLabelOwner(s, a) // Only owners should be allowed to delete a label
}

// CanRead checks if a user can read a label
func (l *Label) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	return l.hasAccessToLabel(s, a)
}

// CanCreate checks if the user can create a label
// Currently a dummy.
func (l *Label) CanCreate(_ *xorm.Session, a web.Auth) (bool, error) {
	if _, is := a.(*LinkSharing); is {
		return false, nil
	}

	return true, nil
}

func (l *Label) isLabelOwner(s *xorm.Session, a web.Auth) (bool, error) {

	if _, is := a.(*LinkSharing); is {
		return false, nil
	}

	lorig, err := getLabelByIDSimple(s, l.ID)
	if err != nil {
		return false, err
	}
	return lorig.CreatedByID == a.GetID(), nil
}

// Helper method to check if a user can see a specific label.
//
// A user can read a label when at least one of the following is true:
//
//  1. The auth is a real user and the user created the label, OR
//  2. The label is attached to a task in a project the auth can access.
//
// The implementation uses explicit builder.And / builder.Or grouping so
// the boolean precedence is unambiguous. A previous implementation chained
// xorm session .Where / .Or / .And calls which SQL flattened to
// `WHERE A OR B OR C AND D`, leaking any label with any label_tasks row
// to any authenticated user (GHSA-hj5c-mhh2-g7jq).
func (l *Label) hasAccessToLabel(s *xorm.Session, a web.Auth) (has bool, maxPermission int, err error) {

	linkShare, isLinkShare := a.(*LinkSharing)

	// Build the "task is in a project the caller can access" subquery.
	var accessibleProjects builder.Cond
	if isLinkShare {
		accessibleProjects = builder.Eq{"project_id": linkShare.ProjectID}
	} else {
		accessibleProjects = builder.In(
			"project_id",
			getUserProjectsStatement(a.GetID(), "").Select("l.id"),
		)
	}

	labelAttachedToAccessibleTask := builder.In(
		"label_tasks.task_id",
		builder.
			Select("id").
			From("tasks").
			Where(accessibleProjects),
	)

	// A user can see a label if:
	//   - they created it (only when the auth is an actual user), OR
	//   - it is attached to a task in a project they have access to.
	//
	// The outer AND enforces that the result is scoped to the requested label ID.
	accessBranches := []builder.Cond{labelAttachedToAccessibleTask}
	if !isLinkShare {
		accessBranches = append(accessBranches, builder.Eq{"labels.created_by_id": a.GetID()})
	}

	cond := builder.And(
		builder.Eq{"labels.id": l.ID},
		builder.Or(accessBranches...),
	)

	ll := &LabelTask{}
	has, err = s.Table("labels").
		Select("label_tasks.*").
		Join("LEFT", "label_tasks", "label_tasks.label_id = labels.id").
		Where(cond).
		Get(ll)
	if err != nil || !has {
		return
	}

	// If the label was matched via an attached task, compute the caller's
	// permission level from that task. Otherwise (creator-only branch with
	// no attachment) default to read permission.
	if ll.TaskID > 0 {
		t := &Task{ID: ll.TaskID}
		_, maxPermission, err = t.CanRead(s, a)
		if err != nil {
			return
		}
		return
	}

	maxPermission = int(PermissionRead)
	return
}
