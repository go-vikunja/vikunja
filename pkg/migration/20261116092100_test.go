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

package migration

import (
	"strconv"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm/schemas"
)

func TestResetInvalidFrontendSettings20261116092100(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, x.Sync2(user.GetTables()...))

	atLimit := `{"a":"` + strings.Repeat("a", maxFrontendSettingsSize20261116092100-8) + `"}`
	overLimit := `{"a":"` + strings.Repeat("a", maxFrontendSettingsSize20261116092100-7) + `"}`

	cases := []struct {
		name  string
		value string
		keep  bool
	}{
		{"object", `{"color_schema":"dark"}`, true},
		{"empty object", `{}`, true},
		{"object at the limit", atLimit, true},
		{"object over the limit", overLimit, false},
		{"string", `"hello world"`, false},
		{"array", `[1,2]`, false},
		{"number", `42`, false},
		{"json null", `null`, false},
	}
	// Postgres stores frontend_settings in a json column, which rejects invalid JSON.
	if x.Dialect().URI().DBType != schemas.POSTGRES {
		cases = append(cases, struct {
			name  string
			value string
			keep  bool
		}{"invalid json", `{"color_schema":`, false})
	}

	for _, c := range cases {
		_, err := x.Insert(&user.User{Username: "fs20261116092100 " + c.name})
		require.NoError(t, err)
		_, err = x.Exec("UPDATE users SET frontend_settings = ? WHERE username = ?", c.value, "fs20261116092100 "+c.name)
		require.NoError(t, err)
	}
	_, err = x.Insert(&user.User{Username: "fs20261116092100 no value"})
	require.NoError(t, err)

	require.NoError(t, resetInvalidFrontendSettings20261116092100(x))

	read := func(username string) []map[string]string {
		rows, err := x.SQL("SELECT frontend_settings FROM users WHERE username = ? AND frontend_settings IS NOT NULL", username).QueryString()
		require.NoError(t, err)
		return rows
	}
	for _, c := range cases {
		rows := read("fs20261116092100 " + c.name)
		if c.keep {
			require.Len(t, rows, 1, c.name)
			assert.Equal(t, c.value, rows[0]["frontend_settings"], c.name)
		} else {
			assert.Empty(t, rows, c.name)
		}
	}
	assert.Empty(t, read("fs20261116092100 no value"))

	// Idempotent on clean data.
	require.NoError(t, resetInvalidFrontendSettings20261116092100(x))
	assert.Len(t, read("fs20261116092100 object"), 1)
}

func TestResetInvalidFrontendSettingsBatches20261116092100(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	require.NoError(t, x.Sync2(user.GetTables()...))

	// Every second user has an invalid value, across more than two batches.
	total := frontendSettingsBatch20261116092100*2 + 3
	for i := 0; i < total; i++ {
		username := "fsbatch20261116092100 " + strconv.Itoa(i)
		_, err := x.Insert(&user.User{Username: username})
		require.NoError(t, err)
		value := `{"color_schema":"dark"}`
		if i%2 == 1 {
			value = `[1,2]`
		}
		_, err = x.Exec("UPDATE users SET frontend_settings = ? WHERE username = ?", value, username)
		require.NoError(t, err)
	}

	require.NoError(t, resetInvalidFrontendSettings20261116092100(x))

	for i := 0; i < total; i++ {
		rows, err := x.SQL("SELECT frontend_settings FROM users WHERE username = ? AND frontend_settings IS NOT NULL",
			"fsbatch20261116092100 "+strconv.Itoa(i)).QueryString()
		require.NoError(t, err)
		if i%2 == 1 {
			assert.Empty(t, rows, i)
		} else {
			assert.Len(t, rows, 1, i)
		}
	}
}
