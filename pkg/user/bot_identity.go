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
	"code.vikunja.io/api/pkg/web"

	"xorm.io/builder"
)

// IsBotOwnedBy reports whether u is a bot owned by a.
//
// Directional on purpose: it grants the owner access to what their bots own,
// never the reverse, so a bot cannot act as the human who owns it.
func (u *User) IsBotOwnedBy(a web.Auth) bool {
	return u.IsBot() && u.BotOwnerID == a.GetID()
}

// SameBotIdentityCond matches a user id column against every user sharing the
// caller's identity root - a bot's owner, or a human themselves. That covers
// the caller, its owner, and all bots under that owner, in every direction.
//
// The root is resolved in SQL because callers run this once per row, and
// because auth values built from a JWT or a CalDAV login carry no BotOwnerID.
func SameBotIdentityCond(a web.Auth, column string) builder.Cond {
	id := a.GetID()

	// Bots cannot own bots (CreateBotUser), so the root is at most one hop away.
	root := builder.Select("bot_owner_id").From("users").
		Where(builder.And(builder.Eq{"id": id}, builder.Gt{"bot_owner_id": 0}))

	return builder.In(column,
		builder.Select("id").From("users").Where(builder.Or(
			builder.Eq{"id": id},
			builder.Eq{"bot_owner_id": id},
			builder.In("id", root),
			builder.In("bot_owner_id", root),
		)),
	)
}
