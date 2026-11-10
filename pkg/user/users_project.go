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
	"unicode/utf8"

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

// SearchUsersPageSize is the page size of the global user search.
const SearchUsersPageSize = 25

// globalSearchMinLength keeps a single letter from listing half the instance.
const globalSearchMinLength = 2

// SearchUsers performs the global user search shared by both API versions.
// It returns one ranked page of matches plus the total number of matches.
// Email addresses are only returned when service.enableglobalusersearch is on.
func SearchUsers(s *xorm.Session, search string, currentUser *User, page int) (users []*User, total int64, err error) {
	cond := userSearchCond(search, currentUser, &ProjectUserOpts{})
	if cond == nil {
		return nil, 0, nil
	}
	if page < 1 {
		page = 1
	}

	order, args := db.SearchRelevanceOrder(strings.TrimSpace(sanitizeUserSearch(search)), "username", "name")
	total, err = s.
		Where(cond).
		OrderBy(order, args...).
		OrderBy("username ASC").
		Limit(SearchUsersPageSize, (page-1)*SearchUsersPageSize).
		FindAndCount(&users)
	if err != nil {
		return nil, 0, err
	}

	if !config.ServiceEnableGlobalUserSearch.GetBool() {
		for _, u := range users {
			u.Email = ""
		}
	}
	return users, total, nil
}

// ListUsers returns a list with all users, filtered by an optional search string
func ListUsers(s *xorm.Session, search string, currentUser *User, opts *ProjectUserOpts) (users []*User, err error) {
	if opts == nil {
		opts = &ProjectUserOpts{}
	}

	cond := userSearchCond(search, currentUser, opts)
	if cond == nil {
		return
	}

	err = s.
		Where(cond).
		OrderBy("id").
		Find(&users)

	search = sanitizeUserSearch(search)
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

// Prevent searching for placeholders
func sanitizeUserSearch(search string) string {
	return strings.ReplaceAll(search, "%", "")
}

func globalSearchParts(parts []string) []string {
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if utf8.RuneCountInString(part) >= globalSearchMinLength {
			kept = append(kept, part)
		}
	}
	return kept
}

// userSearchCond returns nil when the search must not return anything.
func userSearchCond(search string, currentUser *User, opts *ProjectUserOpts) builder.Cond {
	search = sanitizeUserSearch(search)

	if (search == "" || strings.ReplaceAll(search, " ", "") == "") && !opts.ReturnAllIfNoSearchProvided {
		return nil
	}

	// Global mode ignores the per-user discoverability settings.
	global := !opts.MatchFuzzily && config.ServiceEnableGlobalUserSearch.GetBool()

	conds := []builder.Cond{}

	// Subquery: find user IDs that share an external team with the current user
	externalTeamMemberIDs := builder.Select("tm2.user_id").
		From("team_members tm1").
		Join("INNER", "team_members tm2", "tm1.team_id = tm2.team_id").
		Join("INNER", "teams t", "t.id = tm1.team_id").
		Where(builder.And(
			builder.Eq{"tm1.user_id": currentUser.ID},
			builder.Neq{"t.external_id": ""},
			builder.Neq{"tm2.user_id": currentUser.ID},
		))

	queryParts := strings.Split(search, ",")
	if global {
		queryParts = globalSearchParts(queryParts)
		if len(queryParts) == 0 {
			return nil
		}
	}

	if search != "" {
		for _, queryPart := range queryParts {

			if opts.MatchFuzzily || global {
				conds = append(conds,
					db.ILIKE("name", queryPart),
					db.ILIKE("username", queryPart),
					db.ILIKE("email", queryPart),
				)
				continue
			}

			var usernameCond builder.Cond = builder.Eq{"username": queryPart}
			if db.Type() == schemas.POSTGRES {
				usernameCond = builder.Expr("LOWER(username) = LOWER(?)", queryPart)
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
				// External team bypass: match by name or email without discoverability check
				builder.And(
					builder.In("id", externalTeamMemberIDs),
					builder.Or(
						db.ILIKE("name", queryPart),
						builder.Eq{"email": queryPart},
					),
				),
			)
		}
	}

	if !opts.MatchFuzzily && !global {
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

		if !opts.MatchFuzzily && !global {
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

	notSomeoneElsesBot := builder.Or(
		builder.IsNull{"bot_owner_id"},
		builder.Eq{"bot_owner_id": 0},
		builder.Eq{"bot_owner_id": currentUser.ID},
	)

	// Own bots: filtered by the search string when one is provided (matched
	// against username and name), but bypass the discoverable_by_name flag
	// regular users have. Without a search, ListUsers only returns rows when
	// ReturnAllIfNoSearchProvided is set, so we surface own bots there too.
	var ownBotCond builder.Cond = builder.Eq{"bot_owner_id": currentUser.ID}
	if search != "" {
		botSearchConds := []builder.Cond{}
		for _, queryPart := range queryParts {
			if opts.MatchFuzzily {
				botSearchConds = append(botSearchConds,
					db.ILIKE("name", queryPart),
					db.ILIKE("username", queryPart),
				)
				continue
			}

			botSearchConds = append(botSearchConds,
				db.ILIKE("username", queryPart),
				db.ILIKE("name", queryPart),
			)
		}
		ownBotCond = builder.And(
			builder.Eq{"bot_owner_id": currentUser.ID},
			builder.Or(botSearchConds...),
		)
	}

	return builder.Or(
		builder.And(cond, notSomeoneElsesBot),
		ownBotCond,
	)
}

// ListAllUsers returns all users
func ListAllUsers(s *xorm.Session) (users []*User, err error) {
	err = s.Find(&users)
	return
}
