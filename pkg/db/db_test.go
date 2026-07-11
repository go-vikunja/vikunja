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

func TestGetPostgreSQLConnectionString(t *testing.T) {
	t.Run("with schema", func(t *testing.T) {
		connStr := getPostgreSQLConnectionString("localhost:5432", "vikunja", "secret", "vikunja", "vikunja", "disable", "", "", "")
		assert.Equal(t, "postgres://vikunja:secret@localhost:5432/vikunja?sslmode=disable&sslcert=&sslkey=&sslrootcert=&search_path=%22vikunja%22%2Cpublic", connStr)
	})
	t.Run("without schema", func(t *testing.T) {
		connStr := getPostgreSQLConnectionString("localhost:5432", "vikunja", "secret", "vikunja", "", "disable", "", "", "")
		assert.Equal(t, "postgres://vikunja:secret@localhost:5432/vikunja?sslmode=disable&sslcert=&sslkey=&sslrootcert=", connStr)
	})
	t.Run("schema needing quoting", func(t *testing.T) {
		connStr := getPostgreSQLConnectionString("localhost:5432", "vikunja", "secret", "vikunja", "MySchema", "disable", "", "", "")
		assert.Equal(t, "postgres://vikunja:secret@localhost:5432/vikunja?sslmode=disable&sslcert=&sslkey=&sslrootcert=&search_path=%22MySchema%22%2Cpublic", connStr)
	})
	t.Run("unix socket", func(t *testing.T) {
		connStr := getPostgreSQLConnectionString("/var/run/postgresql", "vikunja", "secret", "vikunja", "public", "disable", "", "", "")
		assert.Equal(t, "postgres://vikunja:secret@:5432/vikunja?sslmode=disable&sslcert=&sslkey=&sslrootcert=&host=/var/run/postgresql&search_path=%22public%22", connStr)
	})
}
