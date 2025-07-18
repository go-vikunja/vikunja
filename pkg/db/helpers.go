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
	"strings"

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

func ParadeDBAvailable() bool {
	return Type() == schemas.POSTGRES && paradedbInstalled
}

// MultiFieldSearch performs an optimized search across multiple fields for ParadeDB
// using a single query rather than multiple OR conditions.
// Falls back to individual ILIKE queries for PGroonga and standard PostgreSQL.
func MultiFieldSearch(fields []string, search string) builder.Cond {
	return MultiFieldSearchWithTableAlias(fields, search, "")
}

// MultiFieldSearchWithTableAlias performs an optimized search across multiple fields for ParadeDB
// with support for table aliases. When tableAlias is provided, it will be used to prefix field names
// for non-ParadeDB queries and the id field for ParadeDB queries.
func MultiFieldSearchWithTableAlias(fields []string, search, tableAlias string) builder.Cond {
	if Type() == schemas.POSTGRES && paradedbInstalled {
		if len(fields) == 1 {
			// Single field search - use optimized match function
			return builder.Expr("id @@@ paradedb.match(?, ?)", fields[0], search)
		}
		// Multi-field search - use disjunction_max for optimal performance
		fieldMatches := make([]string, len(fields))
		args := make([]interface{}, len(fields)*2)
		for i, field := range fields {
			fieldMatches[i] = "paradedb.match(?, ?)"
			args[i*2] = field
			args[i*2+1] = search
		}

		idField := "`id`"
		if tableAlias != "" {
			idField = "`" + tableAlias + "`.`id`"
		}

		return builder.Expr(idField+" @@@ paradedb.disjunction_max(ARRAY["+strings.Join(fieldMatches, ", ")+"])", args...)
	}

	// For non-PostgreSQL databases, use ILIKE on all fields
	conditions := make([]builder.Cond, len(fields))
	for i, field := range fields {
		// Add table alias to field name if provided
		fieldName := field
		if tableAlias != "" {
			fieldName = tableAlias + "." + field
		}
		conditions[i] = ILIKE(fieldName, search)
	}
	return builder.Or(conditions...)
}
