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
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
	"xorm.io/xorm"
)

func (n *ReminderDueNotification) ProjectID() int64 { return notificationProjectID(n.Task, n.Project) }

func (n *TaskCommentNotification) ProjectID() int64 { return notificationProjectID(n.Task, n.Project) }

func (n *TaskAssignedNotification) ProjectID() int64 { return notificationProjectID(n.Task, n.Project) }

func (n *TaskDeletedNotification) ProjectID() int64 { return notificationProjectID(n.Task, nil) }

func (n *UserMentionedInTaskNotification) ProjectID() int64 {
	return notificationProjectID(n.Task, n.Project)
}

func (n *ProjectCreatedNotification) ProjectID() int64 { return notificationProjectID(nil, n.Project) }

func (n *TeamMemberAddedNotification) ProjectID() int64 { return 0 }

func (n *APITokenExpiringWeekNotification) ProjectID() int64 { return 0 }

func (n *APITokenExpiringDayNotification) ProjectID() int64 { return 0 }

func notificationProjectID(t *Task, p *Project) int64 {
	if p != nil && p.ID > 0 {
		return p.ID
	}
	if t != nil && t.ProjectID > 0 {
		return t.ProjectID
	}
	return notifications.ProjectIDUnresolved
}

// NotificationProjectFilter restricts stored notifications to the ones the
// caller may read. No isInstanceAdmin bypass on purpose: admin says nothing
// about whether they may still read a project they were removed from.
func NotificationProjectFilter(a web.Auth) builder.Cond {
	if isLinkShare(a) {
		// The empty slice is load-bearing: builder.In with no argument is not a
		// valid cond and gets dropped, matching every row.
		return builder.In("project_id", []int64{})
	}

	// The unresolved sentinel is negative, so it matches neither arm.
	return builder.Or(
		builder.Eq{"project_id": 0},
		accessibleProjectIDsSubquery(a, "project_id"),
	)
}

// CanReadNotification is NotificationProjectFilter for a single loaded row.
func CanReadNotification(s *xorm.Session, a web.Auth, dbn *notifications.DatabaseNotification) (bool, error) {
	if isLinkShare(a) {
		return false, nil
	}
	if dbn.ProjectID == 0 {
		return true, nil
	}
	if dbn.ProjectID < 0 {
		return false, nil
	}

	count, err := s.
		Where(builder.And(
			builder.Eq{"id": dbn.ProjectID},
			accessibleProjectIDsSubquery(a, "id"),
		)).
		Count(&Project{})
	return count > 0, err
}

// A link share owns no notifications; accessibleProjectIDsSubquery would hand it
// every row of the project it is shared on.
func isLinkShare(a web.Auth) bool {
	_, is := a.(*LinkSharing)
	return is
}
