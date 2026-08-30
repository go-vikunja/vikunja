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
	"code.vikunja.io/api/pkg/user"
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

	// Link shares legitimately reach here through hasAccessToLabel, and are not
	// users: a plain denial, not an error.
	caller, err := user.GetFromAuth(a)
	if user.IsErrMustNotBeLinkShare(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	lorig, err := getLabelByIDSimple(s, l.ID)
	if err != nil {
		return false, err
	}
	if lorig.CreatedByID == caller.ID {
		return true, nil
	}

	creator, err := user.GetUserByID(s, lorig.CreatedByID)
	if err != nil {
		if user.IsErrUserDoesNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return creator.IsBotOwnedBy(caller), nil
}

// labelVisibleCond matches every label the caller may see: used on a task in a
// project they can access, or created by their own bot identity. It is the one
// definition of label visibility - single-label reads and the list must not drift.
//
// The cond is assembled from builder.And / builder.Or values and must be handed
// to Where in one shot. Chaining xorm's session .Where/.Or/.And instead flattens
// the SQL to `A OR B OR C AND D`, which leaked any label with any label_tasks row
// to any authenticated user (GHSA-hj5c-mhh2-g7jq).
func labelVisibleCond(s *xorm.Session, a web.Auth) (builder.Cond, error) {

	// Must include projects inherited via a shared parent, otherwise users can
	// remove but not re-add labels on tasks in child projects.
	accessibleProjects, err := accessibleProjectIDsCond(s, a, "tasks.project_id")
	if err != nil {
		return nil, err
	}

	usedOnAccessibleTask := builder.In("labels.id",
		builder.
			Select("label_tasks.label_id").
			From("label_tasks").
			InnerJoin("tasks", "tasks.id = label_tasks.task_id").
			Where(builder.And(accessibleProjects, taskNotDeletedCond("tasks"))),
	)

	accessBranches := []builder.Cond{usedOnAccessibleTask}
	if _, isLinkShare := a.(*LinkSharing); !isLinkShare {
		caller, err := user.GetFromAuth(a)
		if err != nil {
			return nil, err
		}
		accessBranches = append(accessBranches, user.SameBotIdentityCond(caller, "labels.created_by_id"))
	}

	return builder.Or(accessBranches...), nil
}

// hasAccessToLabel reports whether the caller can read a label and, if so,
// the caller's maximum permission on it.
func (l *Label) hasAccessToLabel(s *xorm.Session, a web.Auth) (has bool, maxPermission int, err error) {

	visible, err := labelVisibleCond(s, a)
	if err != nil {
		return false, 0, err
	}

	has, err = s.Table("labels").
		Where(builder.And(builder.Eq{"labels.id": l.ID}, visible)).
		Exist(&Label{})
	if err != nil || !has {
		return
	}

	// Writes and deletes are owner-only (CanUpdate/CanDelete), so the caller's
	// max permission is admin for the owner and read for anyone else who can see it.
	owner, err := l.isLabelOwner(s, a)
	if err != nil {
		return
	}
	if owner {
		maxPermission = int(PermissionAdmin)
	} else {
		maxPermission = int(PermissionRead)
	}

	return
}
