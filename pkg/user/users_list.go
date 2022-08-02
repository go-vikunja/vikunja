// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
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

package user

import (
	"strings"

	"xorm.io/builder"
	"xorm.io/xorm"
)

type ListUserOpts struct {
	AdditionalCond              builder.Cond
	ReturnAllIfNoSearchProvided bool
}

// ListUsers returns a list with all users, filtered by an optional search string
func ListUsers(s *xorm.Session, search string, opts *ListUserOpts) (users []*User, err error) {
	if opts == nil {
		opts = &ListUserOpts{}
	}

	// Prevent searching for placeholders
	search = strings.ReplaceAll(search, "%", "")

	if (search == "" || strings.ReplaceAll(search, " ", "") == "") && !opts.ReturnAllIfNoSearchProvided {
		return
	}

	cond := builder.Or(
		builder.Like{"username", "%" + search + "%"},
		builder.And(
			builder.Eq{"email": search},
			builder.Eq{"discoverable_by_email": true},
		),
		builder.And(
			builder.Like{"name", "%" + search + "%"},
			builder.Eq{"discoverable_by_name": true},
		),
	)

	if opts.AdditionalCond != nil {
		cond = builder.And(
			cond,
			opts.AdditionalCond,
		)
	}

	err = s.
		Where(cond).
		Find(&users)
	return
}

// ListAllUsers returns all users
func ListAllUsers(s *xorm.Session) (users []*User, err error) {
	err = s.Find(&users)
	return
}
