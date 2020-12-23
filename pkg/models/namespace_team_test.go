// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
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

func TestTeamNamespace_ReadAll(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		tn := TeamNamespace{
			NamespaceID: 3,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		teams, _, _, err := tn.ReadAll(s, u, "", 1, 50)
		assert.NoError(t, err)
		assert.Equal(t, reflect.TypeOf(teams).Kind(), reflect.Slice)
		ts := reflect.ValueOf(teams)
		assert.Equal(t, ts.Len(), 2)
		_ = s.Close()
	})
	t.Run("nonexistant namespace", func(t *testing.T) {
		tn := TeamNamespace{
			NamespaceID: 9999,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		_, _, _, err := tn.ReadAll(s, u, "", 1, 50)
		assert.Error(t, err)
		assert.True(t, IsErrNamespaceDoesNotExist(err))
		_ = s.Close()
	})
	t.Run("no right for namespace", func(t *testing.T) {
		tn := TeamNamespace{
			NamespaceID: 17,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		_, _, _, err := tn.ReadAll(s, u, "", 1, 50)
		assert.Error(t, err)
		assert.True(t, IsErrNeedToHaveNamespaceReadAccess(err))
		_ = s.Close()
	})
}

func TestTeamNamespace_Create(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		tn := TeamNamespace{
			TeamID:      1,
			NamespaceID: 1,
			Right:       RightAdmin,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		allowed, _ := tn.CanCreate(s, u)
		assert.True(t, allowed)
		err := tn.Create(s, u)
		assert.NoError(t, err)

		err = s.Commit()
		assert.NoError(t, err)

		db.AssertExists(t, "team_namespaces", map[string]interface{}{
			"team_id":      1,
			"namespace_id": 1,
			"right":        RightAdmin,
		}, false)
	})
	t.Run("team already has access", func(t *testing.T) {
		tn := TeamNamespace{
			TeamID:      1,
			NamespaceID: 3,
			Right:       RightRead,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		err := tn.Create(s, u)
		assert.Error(t, err)
		assert.True(t, IsErrTeamAlreadyHasAccess(err))
		_ = s.Close()
	})
	t.Run("invalid team right", func(t *testing.T) {
		tn := TeamNamespace{
			TeamID:      1,
			NamespaceID: 3,
			Right:       RightUnknown,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		err := tn.Create(s, u)
		assert.Error(t, err)
		assert.True(t, IsErrInvalidRight(err))
		_ = s.Close()
	})
	t.Run("nonexistant team", func(t *testing.T) {
		tn := TeamNamespace{
			TeamID:      9999,
			NamespaceID: 1,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		err := tn.Create(s, u)
		assert.Error(t, err)
		assert.True(t, IsErrTeamDoesNotExist(err))
		_ = s.Close()
	})
	t.Run("nonexistant namespace", func(t *testing.T) {
		tn := TeamNamespace{
			TeamID:      1,
			NamespaceID: 9999,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		err := tn.Create(s, u)
		assert.Error(t, err)
		assert.True(t, IsErrNamespaceDoesNotExist(err))
		_ = s.Close()
	})
}

func TestTeamNamespace_Delete(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		tn := TeamNamespace{
			TeamID:      7,
			NamespaceID: 9,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		allowed, _ := tn.CanDelete(s, u)
		assert.True(t, allowed)
		err := tn.Delete(s)
		assert.NoError(t, err)
		err = s.Commit()
		assert.NoError(t, err)

		db.AssertMissing(t, "team_namespaces", map[string]interface{}{
			"team_id":      7,
			"namespace_id": 9,
		})
	})
	t.Run("nonexistant team", func(t *testing.T) {
		tn := TeamNamespace{
			TeamID:      9999,
			NamespaceID: 3,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		err := tn.Delete(s)
		assert.Error(t, err)
		assert.True(t, IsErrTeamDoesNotExist(err))
		_ = s.Close()
	})
	t.Run("nonexistant namespace", func(t *testing.T) {
		tn := TeamNamespace{
			TeamID:      1,
			NamespaceID: 9999,
		}
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		err := tn.Delete(s)
		assert.Error(t, err)
		assert.True(t, IsErrTeamDoesNotHaveAccessToNamespace(err))
		_ = s.Close()
	})
}

func TestTeamNamespace_Update(t *testing.T) {
	type fields struct {
		ID          int64
		TeamID      int64
		NamespaceID int64
		Right       Right
		Created     time.Time
		Updated     time.Time
		CRUDable    web.CRUDable
		Rights      web.Rights
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
				NamespaceID: 3,
				TeamID:      1,
				Right:       RightAdmin,
			},
		},
		{
			name: "Test Update to write",
			fields: fields{
				NamespaceID: 3,
				TeamID:      1,
				Right:       RightWrite,
			},
		},
		{
			name: "Test Update to Read",
			fields: fields{
				NamespaceID: 3,
				TeamID:      1,
				Right:       RightRead,
			},
		},
		{
			name: "Test Update with invalid right",
			fields: fields{
				NamespaceID: 3,
				TeamID:      1,
				Right:       500,
			},
			wantErr: true,
			errType: IsErrInvalidRight,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()

			tl := &TeamNamespace{
				ID:          tt.fields.ID,
				TeamID:      tt.fields.TeamID,
				NamespaceID: tt.fields.NamespaceID,
				Right:       tt.fields.Right,
				Created:     tt.fields.Created,
				Updated:     tt.fields.Updated,
				CRUDable:    tt.fields.CRUDable,
				Rights:      tt.fields.Rights,
			}
			err := tl.Update(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("TeamNamespace.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("TeamNamespace.Update() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}

			err = s.Commit()
			assert.NoError(t, err)

			if !tt.wantErr {
				db.AssertExists(t, "team_namespaces", map[string]interface{}{
					"team_id":      tt.fields.TeamID,
					"namespace_id": tt.fields.NamespaceID,
					"right":        tt.fields.Right,
				}, false)
			}
		})
	}
}
