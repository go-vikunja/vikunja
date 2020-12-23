// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"reflect"
	"runtime"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
	"github.com/stretchr/testify/assert"
)

func TestTeamList_ReadAll(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		tl := TeamList{
			TeamID: 1,
			ListID: 3,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		teams, _, _, err := tl.ReadAll(s, u, "", 1, 50)
		assert.NoError(t, err)
		assert.Equal(t, reflect.TypeOf(teams).Kind(), reflect.Slice)
		ts := reflect.ValueOf(teams)
		assert.Equal(t, ts.Len(), 1)
		_ = s.Close()
	})
	t.Run("nonexistant list", func(t *testing.T) {
		tl := TeamList{
			ListID: 99999,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		_, _, _, err := tl.ReadAll(s, u, "", 1, 50)
		assert.Error(t, err)
		assert.True(t, IsErrListDoesNotExist(err))
		_ = s.Close()
	})
	t.Run("namespace owner", func(t *testing.T) {
		tl := TeamList{
			TeamID: 1,
			ListID: 2,
			Right:  RightAdmin,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		_, _, _, err := tl.ReadAll(s, u, "", 1, 50)
		assert.NoError(t, err)
		_ = s.Close()
	})
	t.Run("no access", func(t *testing.T) {
		tl := TeamList{
			TeamID: 1,
			ListID: 5,
			Right:  RightAdmin,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		_, _, _, err := tl.ReadAll(s, u, "", 1, 50)
		assert.Error(t, err)
		assert.True(t, IsErrNeedToHaveListReadAccess(err))
		_ = s.Close()
	})
}

func TestTeamList_Create(t *testing.T) {
	u := &user.User{ID: 1}
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamList{
			TeamID: 1,
			ListID: 1,
			Right:  RightAdmin,
		}
		allowed, _ := tl.CanCreate(s, u)
		assert.True(t, allowed)
		err := tl.Create(s, u)
		assert.NoError(t, err)
		err = s.Commit()
		assert.NoError(t, err)
		db.AssertExists(t, "team_list", map[string]interface{}{
			"team_id": 1,
			"list_id": 1,
			"right":   RightAdmin,
		}, false)
	})
	t.Run("team already has access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamList{
			TeamID: 1,
			ListID: 3,
			Right:  RightAdmin,
		}
		err := tl.Create(s, u)
		assert.Error(t, err)
		assert.True(t, IsErrTeamAlreadyHasAccess(err))
		_ = s.Close()
	})
	t.Run("wrong rights", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamList{
			TeamID: 1,
			ListID: 1,
			Right:  RightUnknown,
		}
		err := tl.Create(s, u)
		assert.Error(t, err)
		assert.True(t, IsErrInvalidRight(err))
		_ = s.Close()
	})
	t.Run("nonexistant team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamList{
			TeamID: 9999,
			ListID: 1,
		}
		err := tl.Create(s, u)
		assert.Error(t, err)
		assert.True(t, IsErrTeamDoesNotExist(err))
		_ = s.Close()
	})
	t.Run("nonexistant list", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamList{
			TeamID: 1,
			ListID: 9999,
		}
		err := tl.Create(s, u)
		assert.Error(t, err)
		assert.True(t, IsErrListDoesNotExist(err))
		_ = s.Close()
	})
}

func TestTeamList_Delete(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamList{
			TeamID: 1,
			ListID: 3,
		}
		err := tl.Delete(s)
		assert.NoError(t, err)
		err = s.Commit()
		assert.NoError(t, err)
		db.AssertMissing(t, "team_list", map[string]interface{}{
			"team_id": 1,
			"list_id": 3,
		})
	})
	t.Run("nonexistant team", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamList{
			TeamID: 9999,
			ListID: 1,
		}
		err := tl.Delete(s)
		assert.Error(t, err)
		assert.True(t, IsErrTeamDoesNotExist(err))
		_ = s.Close()
	})
	t.Run("nonexistant list", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		tl := TeamList{
			TeamID: 1,
			ListID: 9999,
		}
		err := tl.Delete(s)
		assert.Error(t, err)
		assert.True(t, IsErrTeamDoesNotHaveAccessToList(err))
		_ = s.Close()
	})
}

func TestTeamList_Update(t *testing.T) {
	type fields struct {
		ID       int64
		TeamID   int64
		ListID   int64
		Right    Right
		Created  time.Time
		Updated  time.Time
		CRUDable web.CRUDable
		Rights   web.Rights
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
				ListID: 3,
				TeamID: 1,
				Right:  RightAdmin,
			},
		},
		{
			name: "Test Update to write",
			fields: fields{
				ListID: 3,
				TeamID: 1,
				Right:  RightWrite,
			},
		},
		{
			name: "Test Update to Read",
			fields: fields{
				ListID: 3,
				TeamID: 1,
				Right:  RightRead,
			},
		},
		{
			name: "Test Update with invalid right",
			fields: fields{
				ListID: 3,
				TeamID: 1,
				Right:  500,
			},
			wantErr: true,
			errType: IsErrInvalidRight,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()

			tl := &TeamList{
				ID:       tt.fields.ID,
				TeamID:   tt.fields.TeamID,
				ListID:   tt.fields.ListID,
				Right:    tt.fields.Right,
				Created:  tt.fields.Created,
				Updated:  tt.fields.Updated,
				CRUDable: tt.fields.CRUDable,
				Rights:   tt.fields.Rights,
			}
			err := tl.Update(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("TeamList.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("TeamList.Update() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}
			err = s.Commit()
			assert.NoError(t, err)
			if !tt.wantErr {
				db.AssertExists(t, "team_list", map[string]interface{}{
					"list_id": tt.fields.ListID,
					"team_id": tt.fields.TeamID,
					"right":   tt.fields.Right,
				}, false)
			}
		})
	}
}
