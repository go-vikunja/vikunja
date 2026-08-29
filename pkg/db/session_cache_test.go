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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsWriteStatement(t *testing.T) {
	tests := map[string]bool{
		"SELECT * FROM projects": false,
		"\n  select 1":           false,
		"BEGIN TRANSACTION":      false,
		"COMMIT":                 false,
		"ROLLBACK":               false,
		"PREPARE":                false,
		"WITH RECURSIVE tree AS (SELECT 1) SELECT * FROM tree":                        false,
		"WITH t AS (SELECT created, updated, deleted FROM tasks) SELECT * FROM t":     false,
		"INSERT INTO projects (id) VALUES (1)":                                        true,
		"  update projects set title = ?":                                             true,
		"DELETE FROM users_projects":                                                  true,
		"CREATE INDEX foo ON bar (baz)":                                               true,
		"EXPLAIN ANALYZE DELETE FROM tasks":                                           true,
		"WITH moved AS (DELETE FROM a RETURNING *) INSERT INTO b SELECT * FROM moved": true,
		"": true,
	}
	for sqlStr, want := range tests {
		assert.Equal(t, want, isWriteStatement(sqlStr), sqlStr)
	}
}
