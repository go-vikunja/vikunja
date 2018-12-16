//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/web"
	"github.com/stretchr/testify/assert"
	"reflect"
	"runtime"
	"testing"
)

func TestTeamNamespace(t *testing.T) {
	// Dummy team <-> namespace relation
	tn := TeamNamespace{
		TeamID:      1,
		NamespaceID: 1,
		Right:       TeamRightAdmin,
	}

	dummyuser, err := GetUserByID(1)
	assert.NoError(t, err)

	// Test normal creation
	assert.True(t, tn.CanCreate(&dummyuser))
	err = tn.Create(&dummyuser)
	assert.NoError(t, err)

	// Test again (should fail)
	err = tn.Create(&dummyuser)
	assert.Error(t, err)
	assert.True(t, IsErrTeamAlreadyHasAccess(err))

	// Test with invalid team right
	tn2 := tn
	tn2.Right = TeamRightUnknown
	err = tn2.Create(&dummyuser)
	assert.Error(t, err)
	assert.True(t, IsErrInvalidTeamRight(err))

	// Check with inexistant team
	tn3 := tn
	tn3.TeamID = 324
	err = tn3.Create(&dummyuser)
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))

	// Check with a namespace which does not exist
	tn4 := tn
	tn4.NamespaceID = 423
	err = tn4.Create(&dummyuser)
	assert.Error(t, err)
	assert.True(t, IsErrNamespaceDoesNotExist(err))

	// Check readall
	teams, err := tn.ReadAll("", &dummyuser, 1)
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(teams).Kind(), reflect.Slice)
	s := reflect.ValueOf(teams)
	assert.Equal(t, s.Len(), 1)

	// Check readall for a nonexistant namespace
	_, err = tn4.ReadAll("", &dummyuser, 1)
	assert.Error(t, err)
	assert.True(t, IsErrNamespaceDoesNotExist(err))

	// Check with no right to read the namespace
	nouser := &User{ID: 393}
	_, err = tn.ReadAll("", nouser, 1)
	assert.Error(t, err)
	assert.True(t, IsErrNeedToHaveNamespaceReadAccess(err))

	// Delete it
	assert.True(t, tn.CanDelete(&dummyuser))
	err = tn.Delete()
	assert.NoError(t, err)

	// Try deleting with a nonexisting team
	err = tn3.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))

	// Try deleting with a nonexistant namespace
	err = tn4.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotHaveAccessToNamespace(err))
}

func TestTeamNamespace_Update(t *testing.T) {
	type fields struct {
		ID          int64
		TeamID      int64
		NamespaceID int64
		Right       TeamRight
		Created     int64
		Updated     int64
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
				Right:       TeamRightAdmin,
			},
		},
		{
			name: "Test Update to write",
			fields: fields{
				NamespaceID: 3,
				TeamID:      1,
				Right:       TeamRightWrite,
			},
		},
		{
			name: "Test Update to Read",
			fields: fields{
				NamespaceID: 3,
				TeamID:      1,
				Right:       TeamRightRead,
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
			errType: IsErrInvalidTeamRight,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			err := tl.Update()
			if (err != nil) != tt.wantErr {
				t.Errorf("TeamNamespace.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("TeamNamespace.Update() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}
		})
	}
}
