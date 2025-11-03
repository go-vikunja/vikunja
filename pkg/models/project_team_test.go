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
	"reflect"
	"runtime"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamProject_ReadAll(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		tl := TeamProject{
			TeamID:    1,
			ProjectID: 3,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		teams, _, _, err := tl.ReadAll(s, u, "", 1, 50)
		require.NoError(t, err)
		assert.Equal(t, reflect.Slice, reflect.TypeOf(teams).Kind())
		ts := reflect.ValueOf(teams)
		assert.Equal(t, 1, ts.Len())
		_ = s.Close()
	})
	t.Run("nonexistant project", func(t *testing.T) {
		tl := TeamProject{
			ProjectID: 99999,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		_, _, _, err := tl.ReadAll(s, u, "", 1, 50)
		require.Error(t, err)
		assert.True(t, IsErrProjectDoesNotExist(err))
		_ = s.Close()
	})
	t.Run("no access", func(t *testing.T) {
		tl := TeamProject{
			TeamID:     1,
			ProjectID:  5,
			Permission: PermissionAdmin,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		_, _, _, err := tl.ReadAll(s, u, "", 1, 50)
		require.Error(t, err)
		assert.True(t, IsErrNeedToHaveProjectReadAccess(err))
		_ = s.Close()
	})
	t.Run("search", func(t *testing.T) {
		tl := TeamProject{
			ProjectID: 19,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		teams, _, _, err := tl.ReadAll(s, u, "TEAM9", 1, 50)
		require.NoError(t, err)
		assert.Equal(t, reflect.Slice, reflect.TypeOf(teams).Kind())
		ts := teams.([]*TeamWithPermission)
		assert.Len(t, ts, 1)
		assert.Equal(t, int64(9), ts[0].ID)
		_ = s.Close()
	})
}

func TestTeamProject_Create(t *testing.T) {
	u := &user.User{ID: 1}
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamProject{
			TeamID:     1,
			ProjectID:  1,
			Permission: PermissionAdmin,
		}
		allowed, _ := tl.CanCreate(s, u)
		assert.True(t, allowed)
		err := tl.Create(s, u)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		db.AssertExists(t, "team_projects", map[string]interface{}{
			"team_id":    1,
			"project_id": 1,
			"permission": PermissionAdmin,
		}, false)
	})
	t.Run("team already has access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamProject{
			TeamID:     1,
			ProjectID:  3,
			Permission: PermissionAdmin,
		}
		err := tl.Create(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTeamAlreadyHasAccess(err))
		_ = s.Close()
	})
	t.Run("wrong permissions", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamProject{
			TeamID:     1,
			ProjectID:  1,
			Permission: PermissionUnknown,
		}
		err := tl.Create(s, u)
		require.Error(t, err)
		assert.True(t, IsErrInvalidPermission(err))
		_ = s.Close()
	})
	t.Run("nonexistant team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamProject{
			TeamID:    9999,
			ProjectID: 1,
		}
		err := tl.Create(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTeamDoesNotExist(err))
		_ = s.Close()
	})
	t.Run("nonexistant project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamProject{
			TeamID:    1,
			ProjectID: 9999,
		}
		err := tl.Create(s, u)
		require.Error(t, err)
		assert.True(t, IsErrProjectDoesNotExist(err))
		_ = s.Close()
	})
}

func TestTeamProject_Delete(t *testing.T) {
	user := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamProject{
			TeamID:    1,
			ProjectID: 3,
		}
		err := tl.Delete(s, user)
		require.NoError(t, err)
		err = s.Commit()
		require.NoError(t, err)
		db.AssertMissing(t, "team_projects", map[string]interface{}{
			"team_id":    1,
			"project_id": 3,
		})
	})
	t.Run("nonexistant team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamProject{
			TeamID:    9999,
			ProjectID: 1,
		}
		err := tl.Delete(s, user)
		require.Error(t, err)
		assert.True(t, IsErrTeamDoesNotExist(err))
		_ = s.Close()
	})
	t.Run("nonexistant project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamProject{
			TeamID:    1,
			ProjectID: 9999,
		}
		err := tl.Delete(s, user)
		require.Error(t, err)
		assert.True(t, IsErrTeamDoesNotHaveAccessToProject(err))
		_ = s.Close()
	})
}

func TestTeamProject_Update(t *testing.T) {
	type fields struct {
		ID          int64
		TeamID      int64
		ProjectID   int64
		Permission  Permission
		Created     time.Time
		Updated     time.Time
		CRUDable    web.CRUDable
		Permissions web.Permissions
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		errType func(err error) bool
	}{
		{
			name: "Test Update Normally",
			fields: fields{
				ProjectID:  3,
				TeamID:     1,
				Permission: PermissionAdmin,
			},
		},
		{
			name: "Test Update to write",
			fields: fields{
				ProjectID:  3,
				TeamID:     1,
				Permission: PermissionWrite,
			},
		},
		{
			name: "Test Update to Read",
			fields: fields{
				ProjectID:  3,
				TeamID:     1,
				Permission: PermissionRead,
			},
		},
		{
			name: "Test Update with invalid permission",
			fields: fields{
				ProjectID:  3,
				TeamID:     1,
				Permission: 500,
			},
			wantErr: true,
			errType: IsErrInvalidPermission,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()

			tl := &TeamProject{
				ID:          tt.fields.ID,
				TeamID:      tt.fields.TeamID,
				ProjectID:   tt.fields.ProjectID,
				Permission:  tt.fields.Permission,
				Created:     tt.fields.Created,
				Updated:     tt.fields.Updated,
				CRUDable:    tt.fields.CRUDable,
				Permissions: tt.fields.Permissions,
			}
			err := tl.Update(s, &user.User{ID: 1})
			if (err != nil) != tt.wantErr {
				t.Errorf("TeamProject.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("TeamProject.Update() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}
			err = s.Commit()
			require.NoError(t, err)
			if !tt.wantErr {
				db.AssertExists(t, "team_projects", map[string]interface{}{
					"project_id": tt.fields.ProjectID,
					"team_id":    tt.fields.TeamID,
					"permission": tt.fields.Permission,
				}, false)
			}
		})
	}
}
