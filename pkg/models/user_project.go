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
	"xorm.io/builder"
	"xorm.io/xorm"
)

// ProjectUIDs hold all kinds of user IDs from accounts who have access to a project
type ProjectUIDs struct {
	ProjectOwnerID    int64 `xorm:"projectOwner"`
	ProjectUserID     int64 `xorm:"ulID"`
	TeamProjectUserID int64 `xorm:"tlUID"`
}

// ListUsersFromProject returns a list with all users who have access to a project, regardless of the method which gave them access
func ListUsersFromProject(s *xorm.Session, l *Project, currentUser *user.User, search string) (users []*user.User, err error) {

	userids := []*ProjectUIDs{}

	var currentProject *Project
	currentProject, err = GetProjectSimpleByID(s, l.ID)
	if err != nil {
		return nil, err
	}

	for {
		currentUserIDs := []*ProjectUIDs{}
		err = s.
			Select(`l.owner_id as projectOwner,
			ul.user_id as ulID,
			tm2.user_id as tlUID`).
			Table("projects").
			Alias("l").
			// User stuff
			Join("LEFT", []string{"users_projects", "ul"}, "ul.project_id = l.id").
			// Team stuff
			Join("LEFT", []string{"team_projects", "tl"}, "l.id = tl.project_id").
			Join("LEFT", []string{"team_members", "tm2"}, "tm2.team_id = tl.team_id").
			// The actual condition
			Where(
				builder.Or(
					builder.Or(builder.Eq{"ul.permission": PermissionRead}),
					builder.Or(builder.Eq{"tl.permission": PermissionRead}),

					builder.Or(builder.Eq{"ul.permission": PermissionWrite}),
					builder.Or(builder.Eq{"tl.permission": PermissionWrite}),

					builder.Or(builder.Eq{"ul.permission": PermissionAdmin}),
					builder.Or(builder.Eq{"tl.permission": PermissionAdmin}),
				),
				builder.Eq{"l.id": currentProject.ID},
			).
			Find(&currentUserIDs)
		if err != nil {
			return
		}
		userids = append(userids, currentUserIDs...)

		if currentProject.ParentProjectID == 0 {
			break
		}

		parent, err := GetProjectSimpleByID(s, currentProject.ParentProjectID)
		if err != nil && !IsErrProjectDoesNotExist(err) {
			return nil, err
		}
		if err != nil && IsErrProjectDoesNotExist(err) {
			break
		}

		currentProject = parent
	}

	// Remove duplicates from the project of ids and make it a slice
	uidmap := make(map[int64]bool)
	uidmap[l.OwnerID] = true
	for _, u := range userids {
		uidmap[u.ProjectUserID] = true
		uidmap[u.TeamProjectUserID] = true
	}

	uids := make([]int64, 0, len(uidmap))
	for id := range uidmap {
		uids = append(uids, id)
	}

	var cond builder.Cond

	if len(uids) > 0 {
		cond = builder.In("id", uids)
	}

	users, err = user.ListUsers(s, search, currentUser, &user.ProjectUserOpts{
		AdditionalCond:              cond,
		ReturnAllIfNoSearchProvided: true,
		MatchFuzzily:                true,
	})
	return
}
