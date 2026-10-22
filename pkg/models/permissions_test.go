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

package models

import (
	"encoding/json"
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionJSON(t *testing.T) {
	t.Run("marshal", func(t *testing.T) {
		for permission, want := range map[Permission]string{
			PermissionUnknown: `null`,
			PermissionRead:    `0`,
			PermissionWrite:   `1`,
			PermissionAdmin:   `2`,
		} {
			b, err := json.Marshal(permission)
			require.NoError(t, err)
			assert.Equal(t, want, string(b), "marshalling permission %d", permission)
		}
	})

	t.Run("unmarshal", func(t *testing.T) {
		for raw, want := range map[string]Permission{
			`null`: PermissionUnknown,
			`-1`:   PermissionUnknown,
			`0`:    PermissionRead,
			`1`:    PermissionWrite,
			`2`:    PermissionAdmin,
		} {
			var got Permission
			require.NoError(t, json.Unmarshal([]byte(raw), &got), "unmarshalling %s", raw)
			assert.Equal(t, want, got, "unmarshalling %s", raw)
		}
	})

	t.Run("unmarshal out of range", func(t *testing.T) {
		var got Permission
		require.Error(t, json.Unmarshal([]byte(`3`), &got))
	})

	t.Run("round trip", func(t *testing.T) {
		for _, permission := range []Permission{PermissionUnknown, PermissionRead, PermissionWrite, PermissionAdmin} {
			b, err := json.Marshal(permission)
			require.NoError(t, err)

			var got Permission
			require.NoError(t, json.Unmarshal(b, &got))
			assert.Equal(t, permission, got, "round tripping permission %d", permission)
		}
	})

	t.Run("omitted key stays read", func(t *testing.T) {
		pu := &ProjectUser{}
		require.NoError(t, json.Unmarshal([]byte(`{"username":"user1"}`), pu))
		assert.Equal(t, PermissionRead, pu.Permission)
	})
}

func TestProjectUserCreateWithNullPermission(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	pu := &ProjectUser{ProjectID: 2}
	require.NoError(t, json.Unmarshal([]byte(`{"username":"user1","permission":null}`), pu))
	require.Equal(t, Permission(PermissionUnknown), pu.Permission)

	err := pu.Create(s, &user.User{ID: 3})
	require.Error(t, err)
	assert.True(t, IsErrInvalidPermission(err), "expected an invalid permission error, got %v", err)
}
