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
	"xorm.io/builder"
)

// IsBotOwnedBy reports whether u is a bot owned by owner.
func (u *User) IsBotOwnedBy(owner *User) bool {
	return u.IsBot() && u.BotOwnerID == owner.ID
}

// SameBotIdentityCond matches a user ID column against users sharing u's root:
// a bot's owner or a human themselves. It is symmetric across an owner's fleet.
//
// The root is resolved in SQL so the result never depends on how populated the
// passed struct happens to be; callers routinely hold a User carrying only an ID.
//
// Resolving only one owner hop keeps accidental bot chains out of a human's identity set.
//
// column is interpolated as a raw SQL identifier and must be a trusted literal.
func SameBotIdentityCond(u *User, column string) builder.Cond {
	if u == nil || u.ID <= 0 {
		// Humans are stored with bot_owner_id = 0, so an unguarded zero id would
		// match every one of them.
		return builder.In(column, []int64{})
	}

	root := builder.Select("bot_owner_id").From("users").
		Where(builder.And(builder.Eq{"id": u.ID}, builder.Gt{"bot_owner_id": 0}))

	return builder.In(column,
		builder.Select("id").From("users").Where(builder.Or(
			builder.Eq{"id": u.ID},
			builder.Eq{"bot_owner_id": u.ID},
			builder.In("id", root),
			builder.In("bot_owner_id", root),
		)),
	)
}
