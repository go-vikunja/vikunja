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

// hasAccessToLabel reports whether the caller can read a label and, if so,
// the caller's maximum permission on it.
//
// The access cond is assembled with explicit builder.And / builder.Or.
// Chaining xorm's session .Where/.Or/.And instead flattens the SQL to
// `A OR B OR C AND D`, which leaked any label with any label_tasks row
// to any authenticated user (GHSA-hj5c-mhh2-g7jq).
func (l *Label) hasAccessToLabel(s *xorm.Session, a web.Auth) (has bool, maxPermission int, err error) {

	linkShare, isLinkShare := a.(*LinkSharing)

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

	accessBranches := []builder.Cond{labelAttachedToAccessibleTask}
	if !isLinkShare {
		accessBranches = append(accessBranches, builder.Eq{"labels.created_by_id": a.GetID()})
	}

	cond := builder.And(
		builder.Eq{"labels.id": l.ID},
		builder.Or(accessBranches...),
	)

	has, err = s.Table("labels").
		Join("LEFT", "label_tasks", "label_tasks.label_id = labels.id").
		Where(cond).
		Exist(&Label{})
	if err != nil || !has {
		return
	}

	// maxPermission is derived only from label_tasks rows whose task is
	// actually accessible. The pre-fix code used Get(ll) against the
	// unrestricted LEFT JOIN, so it could return an inaccessible row and
	// yield a wrong (or errored) permission.
	accessibleTaskIDs := []int64{}
	err = s.Table("label_tasks").
		Join("INNER", "tasks", "tasks.id = label_tasks.task_id").
		Where(builder.And(
			builder.Eq{"label_tasks.label_id": l.ID},
			accessibleProjects,
		)).
		Cols("label_tasks.task_id").
		Find(&accessibleTaskIDs)
	if err != nil {
		return
	}

	for _, taskID := range accessibleTaskIDs {
		t := &Task{ID: taskID}
		_, taskPermission, tErr := t.CanRead(s, a)
		if tErr != nil {
			err = tErr
			return
		}
		if taskPermission > maxPermission {
			maxPermission = taskPermission
		}
	}

	// Creator-branch fallback: access came from created_by_id with no
	// accessible task to derive a permission from.
	if len(accessibleTaskIDs) == 0 {
		maxPermission = int(PermissionRead)
	}

	return
}
