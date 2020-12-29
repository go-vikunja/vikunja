// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
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
	"code.vikunja.io/api/pkg/user"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// ListUIDs hold all kinds of user IDs from accounts who have somehow access to a list
type ListUIDs struct {
	ListOwnerID          int64 `xorm:"listOwner"`
	NamespaceUserID      int64 `xorm:"unID"`
	ListUserID           int64 `xorm:"ulID"`
	NamespaceOwnerUserID int64 `xorm:"nOwner"`
	TeamNamespaceUserID  int64 `xorm:"tnUID"`
	TeamListUserID       int64 `xorm:"tlUID"`
}

// ListUsersFromList returns a list with all users who have access to a list, regardless of the method which gave them access
func ListUsersFromList(s *xorm.Session, l *List, search string) (users []*user.User, err error) {

	userids := []*ListUIDs{}

	err = s.
		Select(`l.owner_id as listOwner,
			un.user_id as unID,
			ul.user_id as ulID,
			n.owner_id as nOwner,
			tm.user_id as tnUID,
			tm2.user_id as tlUID`).
		Table("list").
		Alias("l").
		// User stuff
		Join("LEFT", []string{"users_namespace", "un"}, "un.namespace_id = l.namespace_id").
		Join("LEFT", []string{"users_list", "ul"}, "ul.list_id = l.id").
		Join("LEFT", []string{"namespaces", "n"}, "n.id = l.namespace_id").
		// Team stuff
		Join("LEFT", []string{"team_namespaces", "tn"}, " l.namespace_id = tn.namespace_id").
		Join("LEFT", []string{"team_members", "tm"}, "tm.team_id = tn.team_id").
		Join("LEFT", []string{"team_list", "tl"}, "l.id = tl.list_id").
		Join("LEFT", []string{"team_members", "tm2"}, "tm2.team_id = tl.team_id").
		// The actual condition
		Where(
			builder.Or(
				builder.Or(builder.Eq{"ul.right": RightRead}),
				builder.Or(builder.Eq{"un.right": RightRead}),
				builder.Or(builder.Eq{"tl.right": RightRead}),
				builder.Or(builder.Eq{"tn.right": RightRead}),

				builder.Or(builder.Eq{"ul.right": RightWrite}),
				builder.Or(builder.Eq{"un.right": RightWrite}),
				builder.Or(builder.Eq{"tl.right": RightWrite}),
				builder.Or(builder.Eq{"tn.right": RightWrite}),

				builder.Or(builder.Eq{"ul.right": RightAdmin}),
				builder.Or(builder.Eq{"un.right": RightAdmin}),
				builder.Or(builder.Eq{"tl.right": RightAdmin}),
				builder.Or(builder.Eq{"tn.right": RightAdmin}),
			),
			builder.Eq{"l.id": l.ID},
		).
		Find(&userids)
	if err != nil {
		return
	}

	// Remove duplicates from the list of ids and make it a slice
	uidmap := make(map[int64]bool)
	uidmap[l.OwnerID] = true
	for _, u := range userids {
		uidmap[u.ListUserID] = true
		uidmap[u.NamespaceOwnerUserID] = true
		uidmap[u.NamespaceUserID] = true
		uidmap[u.TeamListUserID] = true
		uidmap[u.TeamNamespaceUserID] = true
	}

	uids := make([]int64, len(uidmap))
	for id := range uidmap {
		uids = append(uids, id)
	}

	// Get all users
	err = s.
		Table("users").
		Select("*").
		In("id", uids).
		And("username LIKE ?", "%"+search+"%").
		GroupBy("id").
		OrderBy("id").
		Find(&users)

	// Obfuscate all user emails
	for _, u := range users {
		u.Email = ""
	}

	return
}
