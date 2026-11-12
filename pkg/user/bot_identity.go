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

	// Separate indexed lookups avoid estimating the OR as most of the users table.
	return builder.In(column, builder.Expr(`
  SELECT id FROM users WHERE id = ?
  UNION SELECT id FROM users WHERE bot_owner_id = ?
  UNION SELECT id FROM users WHERE id IN (
   SELECT bot_owner_id FROM users WHERE id = ? AND bot_owner_id > 0
  )
  UNION SELECT id FROM users WHERE bot_owner_id > 0 AND bot_owner_id IN (
   SELECT bot_owner_id FROM users WHERE id = ? AND bot_owner_id > 0
  )`, u.ID, u.ID, u.ID, u.ID))
}
