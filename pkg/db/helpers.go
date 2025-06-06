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

package db

import (
	"xorm.io/builder"
	"xorm.io/xorm/schemas"
)

// ILIKE returns an ILIKE query on postgres and a LIKE query on all other platforms.
// Postgres' is case-sensitive by default.
// To work around this, we're using ILIKE as opposed to normal LIKE statements.
// ILIKE is preferred over LOWER(text) LIKE for performance reasons.
// See https://stackoverflow.com/q/7005302/10924593
func ILIKE(column, search string) builder.Cond {
	if Type() == schemas.POSTGRES {
		return builder.Expr(column+" ILIKE ?", "%"+search+"%")
	}

	return &builder.Like{column, "%" + search + "%"}
}
