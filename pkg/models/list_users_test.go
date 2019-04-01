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
	"reflect"
	"runtime"
	"testing"

	"code.vikunja.io/web"
)

func TestListUser_Create(t *testing.T) {
	type fields struct {
		ID       int64
		UserID   int64
		ListID   int64
		Right    Right
		Created  int64
		Updated  int64
		CRUDable web.CRUDable
		Rights   web.Rights
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
				UserID: 1,
				ListID: 2,
			},
		},
		{
			name: "ListUsers Create for duplicate",
			fields: fields{
				UserID: 1,
				ListID: 2,
			},
			wantErr: true,
			errType: IsErrUserAlreadyHasAccess,
		},
		{
			name: "ListUsers Create with invalid right",
			fields: fields{
				UserID: 1,
				ListID: 2,
				Right:  500,
			},
			wantErr: true,
			errType: IsErrInvalidRight,
		},
		{
			name: "ListUsers Create with inexisting list",
			fields: fields{
				UserID: 1,
				ListID: 2000,
			},
			wantErr: true,
			errType: IsErrListDoesNotExist,
		},
		{
			name: "ListUsers Create with inexisting user",
			fields: fields{
				UserID: 500,
				ListID: 2,
			},
			wantErr: true,
			errType: IsErrUserDoesNotExist,
		},
		{
			name: "ListUsers Create with the owner as shared user",
			fields: fields{
				UserID: 1,
				ListID: 1,
			},
			wantErr: true,
			errType: IsErrUserAlreadyHasAccess,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ul := &ListUser{
				ID:       tt.fields.ID,
				UserID:   tt.fields.UserID,
				ListID:   tt.fields.ListID,
				Right:    tt.fields.Right,
				Created:  tt.fields.Created,
				Updated:  tt.fields.Updated,
				CRUDable: tt.fields.CRUDable,
				Rights:   tt.fields.Rights,
			}
			err := ul.Create(tt.args.a)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListUser.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("ListUser.Create() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}
		})
	}
}

func TestListUser_ReadAll(t *testing.T) {
	type fields struct {
		ID       int64
		UserID   int64
		ListID   int64
		Right    Right
		Created  int64
		Updated  int64
		CRUDable web.CRUDable
		Rights   web.Rights
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
				ListID: 3,
			},
			args: args{
				a: &User{ID: 3},
			},
			want: []*UserWithRight{
				{
					User: User{
						ID:       1,
						Username: "user1",
						Password: "1234",
					},
					Right: RightRead,
				},
				{
					User: User{
						ID:       2,
						Username: "user2",
						Password: "1234",
					},
					Right: RightRead,
				},
			},
		},
		{
			name: "Test ReadAll by a user who does not have access to the list",
			fields: fields{
				ListID: 3,
			},
			args: args{
				a: &User{ID: 4},
			},
			wantErr: true,
			errType: IsErrNeedToHaveListReadAccess,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ul := &ListUser{
				ID:       tt.fields.ID,
				UserID:   tt.fields.UserID,
				ListID:   tt.fields.ListID,
				Right:    tt.fields.Right,
				Created:  tt.fields.Created,
				Updated:  tt.fields.Updated,
				CRUDable: tt.fields.CRUDable,
				Rights:   tt.fields.Rights,
			}
			got, err := ul.ReadAll(tt.args.search, tt.args.a, tt.args.page)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListUser.ReadAll() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("ListUser.ReadAll() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListUser.ReadAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListUser_Update(t *testing.T) {
	type fields struct {
		ID       int64
		UserID   int64
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
				UserID: 1,
				Right:  RightAdmin,
			},
		},
		{
			name: "Test Update to write",
			fields: fields{
				ListID: 3,
				UserID: 1,
				Right:  RightWrite,
			},
		},
		{
			name: "Test Update to Read",
			fields: fields{
				ListID: 3,
				UserID: 1,
				Right:  RightRead,
			},
		},
		{
			name: "Test Update with invalid right",
			fields: fields{
				ListID: 3,
				UserID: 1,
				Right:  500,
			},
			wantErr: true,
			errType: IsErrInvalidRight,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lu := &ListUser{
				ID:       tt.fields.ID,
				UserID:   tt.fields.UserID,
				ListID:   tt.fields.ListID,
				Right:    tt.fields.Right,
				Created:  tt.fields.Created,
				Updated:  tt.fields.Updated,
				CRUDable: tt.fields.CRUDable,
				Rights:   tt.fields.Rights,
			}
			err := lu.Update()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListUser.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("ListUser.Update() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}
		})
	}
}

func TestListUser_Delete(t *testing.T) {
	type fields struct {
		ID       int64
		UserID   int64
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
			name: "Try deleting some unexistant user",
			fields: fields{
				UserID: 1000,
				ListID: 2,
			},
			wantErr: true,
			errType: IsErrUserDoesNotExist,
		},
		{
			name: "Try deleting a user which does not has access but exists",
			fields: fields{
				UserID: 1,
				ListID: 4,
			},
			wantErr: true,
			errType: IsErrUserDoesNotHaveAccessToList,
		},
		{
			name: "Try deleting normally",
			fields: fields{
				UserID: 1,
				ListID: 3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lu := &ListUser{
				ID:       tt.fields.ID,
				UserID:   tt.fields.UserID,
				ListID:   tt.fields.ListID,
				Right:    tt.fields.Right,
				Created:  tt.fields.Created,
				Updated:  tt.fields.Updated,
				CRUDable: tt.fields.CRUDable,
				Rights:   tt.fields.Rights,
			}
			err := lu.Delete()
			if (err != nil) != tt.wantErr {
				t.Errorf("ListUser.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("ListUser.Delete() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}
		})
	}
}
