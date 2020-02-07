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
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"
	"github.com/stretchr/testify/assert"
	"reflect"
	"runtime"
	"testing"
)

func TestTeamList(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	// Dummy relation
	tl := TeamList{
		TeamID: 1,
		ListID: 1,
		Right:  RightAdmin,
	}

	// Dummyuser
	u, err := user.GetUserByID(1)
	assert.NoError(t, err)

	// Check normal creation
	allowed, _ := tl.CanCreate(u)
	assert.True(t, allowed)
	err = tl.Create(u)
	assert.NoError(t, err)

	// Check again
	err = tl.Create(u)
	assert.Error(t, err)
	assert.True(t, IsErrTeamAlreadyHasAccess(err))

	// Check with wrong rights
	tl2 := tl
	tl2.Right = RightUnknown
	err = tl2.Create(u)
	assert.Error(t, err)
	assert.True(t, IsErrInvalidRight(err))

	// Check with inexistant team
	tl3 := tl
	tl3.TeamID = 3253
	err = tl3.Create(u)
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))

	// Check with inexistant list
	tl4 := tl
	tl4.ListID = 3252
	err = tl4.Create(u)
	assert.Error(t, err)
	assert.True(t, IsErrListDoesNotExist(err))

	// Test Read all
	teams, _, _, err := tl.ReadAll(u, "", 1, 50)
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(teams).Kind(), reflect.Slice)
	s := reflect.ValueOf(teams)
	assert.Equal(t, s.Len(), 1)

	// Test Read all for nonexistant list
	_, _, _, err = tl4.ReadAll(u, "", 1, 50)
	assert.Error(t, err)
	assert.True(t, IsErrListDoesNotExist(err))

	// Test Read all for a list where the user is owner of the namespace this list belongs to
	tl5 := tl
	tl5.ListID = 2
	_, _, _, err = tl5.ReadAll(u, "", 1, 50)
	assert.NoError(t, err)

	// Test read all for a list where the user not has access
	tl6 := tl
	tl6.ListID = 5
	_, _, _, err = tl6.ReadAll(u, "", 1, 50)
	assert.Error(t, err)
	assert.True(t, IsErrNeedToHaveListReadAccess(err))

	// Delete
	allowed, _ = tl.CanDelete(u)
	assert.True(t, allowed)
	err = tl.Delete()
	assert.NoError(t, err)

	// Delete a nonexistant team
	err = tl3.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotExist(err))

	// Delete with a nonexistant list
	err = tl4.Delete()
	assert.Error(t, err)
	assert.True(t, IsErrTeamDoesNotHaveAccessToList(err))
}

func TestTeamList_Update(t *testing.T) {
	type fields struct {
		ID       int64
		TeamID   int64
		ListID   int64
		Right    Right
		Created  int64
		Updated  int64
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
			err := tl.Update()
			if (err != nil) != tt.wantErr {
				t.Errorf("TeamList.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("TeamList.Update() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}
		})
	}
}
