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

package user

import (
	"strings"

	"code.vikunja.io/api/pkg/config"

	"code.vikunja.io/api/pkg/db"

	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

type ProjectUserOpts struct {
	AdditionalCond              builder.Cond
	ReturnAllIfNoSearchProvided bool
	MatchFuzzily                bool
}

// ListUsers returns a list with all users, filtered by an optional search string
func ListUsers(s *xorm.Session, search string, currentUser *User, opts *ProjectUserOpts) (users []*User, err error) {
	if opts == nil {
		opts = &ProjectUserOpts{}
	}

	// Prevent searching for placeholders
	search = strings.ReplaceAll(search, "%", "")

	if (search == "" || strings.ReplaceAll(search, " ", "") == "") && !opts.ReturnAllIfNoSearchProvided {
		return
	}

	conds := []builder.Cond{}

	queryParts := strings.Split(search, ",")

	if search != "" {
		for _, queryPart := range queryParts {

			if opts.MatchFuzzily {
				conds = append(conds,
					db.ILIKE("name", queryPart),
					db.ILIKE("username", queryPart),
					db.ILIKE("email", queryPart),
				)
				continue
			}

			var usernameCond builder.Cond = builder.Eq{"username": queryPart}
			if db.Type() == schemas.POSTGRES {
				usernameCond = builder.Expr("username ILIKE ?", queryPart)
			}
			if db.Type() == schemas.SQLITE {
				usernameCond = builder.Expr("username = ? COLLATE NOCASE", queryPart)
			}

			conds = append(conds,
				usernameCond,
				builder.And(
					db.ILIKE("name", queryPart),
					builder.Eq{"discoverable_by_name": true},
				),
			)
		}
	}

	if !opts.MatchFuzzily {
		conds = append(conds,
			builder.And(
				builder.In("email", queryParts),
				builder.Eq{"discoverable_by_email": true},
			),
		)
	}

	cond := builder.Or(conds...)

	if opts.AdditionalCond != nil {
		cond = builder.And(
			cond,
			opts.AdditionalCond,
		)
	}

	if config.ServiceEnableOpenIDTeamUserOnlySearch.GetBool() {
		teamMemberCond := builder.In("id", builder.Select("user_id").
			From("team_members").
			Where(builder.In("team_id",
				builder.Select("team_id").
					From("team_members").
					Where(builder.Eq{"team_members.user_id": currentUser.ID}),
			)),
		)

		if !opts.MatchFuzzily {
			cond = builder.And(
				cond,
				builder.Or(
					teamMemberCond,
					builder.And(
						builder.In("email", queryParts),
						builder.Eq{"discoverable_by_email": true},
					),
				),
			)
		} else {
			cond = builder.And(
				cond,
				teamMemberCond,
			)
		}
	}

	err = s.
		Where(cond).
		Find(&users)

outer:
	for _, u := range users {
		for _, part := range strings.Split(search, ",") {
			if u.Email == part {
				continue outer
			}
		}
		u.Email = ""
	}
	return
}

// ListAllUsers returns all users
func ListAllUsers(s *xorm.Session) (users []*User, err error) {
	err = s.Find(&users)
	return
}
