// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
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
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/modules"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"reflect"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/d4l3k/messagediff.v1"
)

func TestProjectUser_Create(t *testing.T) {
	type fields struct {
		ID        int64
		UserID    int64
		Username  string
		ProjectID int64
		Right     Right
		Created   *modules.Time
		Updated   *modules.Time
		CRUDable  web.CRUDable
		Rights    web.Rights
	}
	type args struct {
		a web.Auth
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		errType func(err error) bool
	}{
		{
			name: "ListUsers Create normally",
			fields: fields{
				Username:  "user1",
				ProjectID: 2,
			},
		},
		{
			name: "ListUsers Create for duplicate",
			fields: fields{
				Username:  "user1",
				ProjectID: 3,
			},
			wantErr: true,
			errType: IsErrUserAlreadyHasAccess,
		},
		{
			name: "ListUsers Create with invalid right",
			fields: fields{
				Username:  "user1",
				ProjectID: 2,
				Right:     500,
			},
			wantErr: true,
			errType: IsErrInvalidRight,
		},
		{
			name: "ListUsers Create with inexisting project",
			fields: fields{
				Username:  "user1",
				ProjectID: 2000,
			},
			wantErr: true,
			errType: IsErrProjectDoesNotExist,
		},
		{
			name: "ListUsers Create with inexisting user",
			fields: fields{
				Username:  "user500",
				ProjectID: 2,
			},
			wantErr: true,
			errType: user.IsErrUserDoesNotExist,
		},
		{
			name: "ListUsers Create with the owner as shared user",
			fields: fields{
				Username:  "user1",
				ProjectID: 1,
			},
			wantErr: true,
			errType: IsErrUserAlreadyHasAccess,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()

			ul := &ProjectUser{
				ID:        tt.fields.ID,
				UserID:    tt.fields.UserID,
				Username:  tt.fields.Username,
				ProjectID: tt.fields.ProjectID,
				Right:     tt.fields.Right,
				Created:   tt.fields.Created,
				Updated:   tt.fields.Updated,
				CRUDable:  tt.fields.CRUDable,
				Rights:    tt.fields.Rights,
			}
			err := ul.Create(s, tt.args.a)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectUser.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("ProjectUser.Create() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}

			err = s.Commit()
			require.NoError(t, err)

			if !tt.wantErr {
				db.AssertExists(t, "users_projects", map[string]interface{}{
					"user_id":    ul.UserID,
					"project_id": tt.fields.ProjectID,
				}, false)
			}
		})
	}
}

func TestProjectUser_ReadAll(t *testing.T) {
	user1Read := &UserWithRight{
		User: user.User{
			ID:                           1,
			Username:                     "user1",
			Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
			Issuer:                       "local",
			EmailRemindersEnabled:        true,
			OverdueTasksRemindersEnabled: true,
			OverdueTasksRemindersTime:    "09:00",
			Created:                      testCreatedTime,
			Updated:                      testUpdatedTime,
		},
		Right: RightRead,
	}
	user2Read := &UserWithRight{
		User: user.User{
			ID:                           2,
			Username:                     "user2",
			Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
			Issuer:                       "local",
			EmailRemindersEnabled:        true,
			OverdueTasksRemindersEnabled: true,
			OverdueTasksRemindersTime:    "09:00",
			DefaultProjectID:             4,
			Created:                      testCreatedTime,
			Updated:                      testUpdatedTime,
		},
		Right: RightRead,
	}

	type fields struct {
		ID        int64
		UserID    int64
		ProjectID int64
		Right     Right
		Created   *modules.Time
		Updated   *modules.Time
		CRUDable  web.CRUDable
		Rights    web.Rights
	}
	type args struct {
		search string
		a      web.Auth
		page   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
		errType func(err error) bool
	}{
		{
			name: "Test readall normal",
			fields: fields{
				ProjectID: 3,
			},
			args: args{
				a: &user.User{ID: 3},
			},
			want: []*UserWithRight{
				user1Read,
				user2Read,
			},
		},
		{
			name: "Test ReadAll by a user who does not have access to the project",
			fields: fields{
				ProjectID: 3,
			},
			args: args{
				a: &user.User{ID: 4},
			},
			wantErr: true,
			errType: IsErrNeedToHaveProjectReadAccess,
		},
		{
			name: "Search",
			fields: fields{
				ProjectID: 3,
			},
			args: args{
				a:      &user.User{ID: 3},
				search: "USER2",
			},
			want: []*UserWithRight{
				user2Read,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()

			ul := &ProjectUser{
				ID:        tt.fields.ID,
				UserID:    tt.fields.UserID,
				ProjectID: tt.fields.ProjectID,
				Right:     tt.fields.Right,
				Created:   tt.fields.Created,
				Updated:   tt.fields.Updated,
				CRUDable:  tt.fields.CRUDable,
				Rights:    tt.fields.Rights,
			}
			got, _, _, err := ul.ReadAll(s, tt.args.a, tt.args.search, tt.args.page, 50)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectUser.ReadAll() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("ProjectUser.ReadAll() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}
			if diff, equal := messagediff.PrettyDiff(got, tt.want); !equal {
				t.Errorf("ProjectUser.ReadAll() = %v, want %v, diff: %v", got, tt.want, diff)
			}
			_ = s.Close()
		})
	}
}

func TestProjectUser_Update(t *testing.T) {
	type fields struct {
		ID        int64
		Username  string
		ProjectID int64
		Right     Right
		Created   *modules.Time
		Updated   *modules.Time
		CRUDable  web.CRUDable
		Rights    web.Rights
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
				ProjectID: 3,
				Username:  "user1",
				Right:     RightAdmin,
			},
		},
		{
			name: "Test Update to write",
			fields: fields{
				ProjectID: 3,
				Username:  "user1",
				Right:     RightWrite,
			},
		},
		{
			name: "Test Update to Read",
			fields: fields{
				ProjectID: 3,
				Username:  "user1",
				Right:     RightRead,
			},
		},
		{
			name: "Test Update with invalid right",
			fields: fields{
				ProjectID: 3,
				Username:  "user1",
				Right:     500,
			},
			wantErr: true,
			errType: IsErrInvalidRight,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()

			lu := &ProjectUser{
				ID:        tt.fields.ID,
				Username:  tt.fields.Username,
				ProjectID: tt.fields.ProjectID,
				Right:     tt.fields.Right,
				Created:   tt.fields.Created,
				Updated:   tt.fields.Updated,
				CRUDable:  tt.fields.CRUDable,
				Rights:    tt.fields.Rights,
			}
			err := lu.Update(s, &user.User{ID: 1})
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectUser.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("ProjectUser.Update() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}

			err = s.Commit()
			require.NoError(t, err)

			if !tt.wantErr {
				db.AssertExists(t, "users_projects", map[string]interface{}{
					"project_id": tt.fields.ProjectID,
					"user_id":    lu.UserID,
					"right":      tt.fields.Right,
				}, false)
			}
		})
	}
}

func TestProjectUser_Delete(t *testing.T) {
	type fields struct {
		ID        int64
		Username  string
		UserID    int64
		ProjectID int64
		Right     Right
		Created   *modules.Time
		Updated   *modules.Time
		CRUDable  web.CRUDable
		Rights    web.Rights
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
		errType func(err error) bool
	}{
		{
			name: "Try deleting some unexistant user",
			fields: fields{
				Username:  "user1000",
				ProjectID: 2,
			},
			wantErr: true,
			errType: user.IsErrUserDoesNotExist,
		},
		{
			name: "Try deleting a user which does not has access but exists",
			fields: fields{
				Username:  "user1",
				ProjectID: 4,
			},
			wantErr: true,
			errType: IsErrUserDoesNotHaveAccessToProject,
		},
		{
			name: "Try deleting normally",
			fields: fields{
				Username:  "user1",
				UserID:    1,
				ProjectID: 3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()

			lu := &ProjectUser{
				ID:        tt.fields.ID,
				Username:  tt.fields.Username,
				ProjectID: tt.fields.ProjectID,
				Right:     tt.fields.Right,
				Created:   tt.fields.Created,
				Updated:   tt.fields.Updated,
				CRUDable:  tt.fields.CRUDable,
				Rights:    tt.fields.Rights,
			}
			err := lu.Delete(s, &user.User{ID: 1})
			if (err != nil) != tt.wantErr {
				t.Errorf("ProjectUser.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("ProjectUser.Delete() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}

			err = s.Commit()
			require.NoError(t, err)

			if !tt.wantErr {
				db.AssertMissing(t, "users_projects", map[string]interface{}{
					"user_id":    tt.fields.UserID,
					"project_id": tt.fields.ProjectID,
				})
			}
		})
	}
}
