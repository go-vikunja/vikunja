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
	"gopkg.in/d4l3k/messagediff.v1"
	"reflect"
	"runtime"
	"testing"
	"time"

	"code.vikunja.io/web"
)

func TestLabel_ReadAll(t *testing.T) {
	type fields struct {
		ID          int64
		Title       string
		Description string
		HexColor    string
		CreatedByID int64
		CreatedBy   *user.User
		Created     time.Time
		Updated     time.Time
		CRUDable    web.CRUDable
		Rights      web.Rights
	}
	type args struct {
		search string
		a      web.Auth
		page   int
	}
	user1 := &user.User{
		ID:       1,
		Username: "user1",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive: true,
		Created:  testCreatedTime,
		Updated:  testUpdatedTime,
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantLs  interface{}
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				a: &user.User{ID: 1},
			},
			wantLs: []*labelWithTaskID{
				{
					Label: Label{
						ID:          1,
						Title:       "Label #1",
						CreatedByID: 1,
						CreatedBy:   user1,
						Created:     testCreatedTime,
						Updated:     testUpdatedTime,
					},
				},
				{
					Label: Label{
						ID:          2,
						Title:       "Label #2",
						CreatedByID: 1,
						CreatedBy:   user1,
						Created:     testCreatedTime,
						Updated:     testUpdatedTime,
					},
				},
				{
					Label: Label{
						ID:          4,
						Title:       "Label #4 - visible via other task",
						Created:     testCreatedTime,
						Updated:     testUpdatedTime,
						CreatedByID: 2,
						CreatedBy: &user.User{
							ID:       2,
							Username: "user2",
							Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
							Created:  testCreatedTime,
							Updated:  testUpdatedTime,
						},
					},
				},
			},
		},
		{
			name: "invalid user",
			args: args{
				a: &user.User{ID: -1},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Label{
				ID:          tt.fields.ID,
				Title:       tt.fields.Title,
				Description: tt.fields.Description,
				HexColor:    tt.fields.HexColor,
				CreatedByID: tt.fields.CreatedByID,
				CreatedBy:   tt.fields.CreatedBy,
				Created:     tt.fields.Created,
				Updated:     tt.fields.Updated,
				CRUDable:    tt.fields.CRUDable,
				Rights:      tt.fields.Rights,
			}
			db.LoadAndAssertFixtures(t)
			gotLs, _, _, err := l.ReadAll(tt.args.a, tt.args.search, tt.args.page, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("Label.ReadAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff, equal := messagediff.PrettyDiff(gotLs, tt.wantLs); !equal {
				t.Errorf("Label.ReadAll() = %v, want %v, diff: %v", gotLs, tt.wantLs, diff)
			}
		})
	}
}

func TestLabel_ReadOne(t *testing.T) {
	type fields struct {
		ID          int64
		Title       string
		Description string
		HexColor    string
		CreatedByID int64
		CreatedBy   *user.User
		Created     time.Time
		Updated     time.Time
		CRUDable    web.CRUDable
		Rights      web.Rights
	}
	user1 := &user.User{
		ID:       1,
		Username: "user1",
		Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
		IsActive: true,
		Created:  testCreatedTime,
		Updated:  testUpdatedTime,
	}
	tests := []struct {
		name          string
		fields        fields
		want          *Label
		wantErr       bool
		errType       func(error) bool
		auth          web.Auth
		wantForbidden bool
	}{
		{
			name: "Get label #1",
			fields: fields{
				ID: 1,
			},
			want: &Label{
				ID:          1,
				Title:       "Label #1",
				CreatedByID: 1,
				CreatedBy:   user1,
				Created:     testCreatedTime,
				Updated:     testUpdatedTime,
			},
			auth: &user.User{ID: 1},
		},
		{
			name: "Get nonexistant label",
			fields: fields{
				ID: 9999,
			},
			wantErr:       true,
			errType:       IsErrLabelDoesNotExist,
			wantForbidden: true,
			auth:          &user.User{ID: 1},
		},
		{
			name: "no rights",
			fields: fields{
				ID: 3,
			},
			wantForbidden: true,
			auth:          &user.User{ID: 1},
		},
		{
			name: "Get label #4 - other user",
			fields: fields{
				ID: 4,
			},
			want: &Label{
				ID:          4,
				Title:       "Label #4 - visible via other task",
				CreatedByID: 2,
				CreatedBy: &user.User{
					ID:       2,
					Username: "user2",
					Password: "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
					Created:  testCreatedTime,
					Updated:  testUpdatedTime,
				},
				Created: testCreatedTime,
				Updated: testUpdatedTime,
			},
			auth: &user.User{ID: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Label{
				ID:          tt.fields.ID,
				Title:       tt.fields.Title,
				Description: tt.fields.Description,
				HexColor:    tt.fields.HexColor,
				CreatedByID: tt.fields.CreatedByID,
				CreatedBy:   tt.fields.CreatedBy,
				Created:     tt.fields.Created,
				Updated:     tt.fields.Updated,
				CRUDable:    tt.fields.CRUDable,
				Rights:      tt.fields.Rights,
			}

			allowed, _ := l.CanRead(tt.auth)
			if !allowed && !tt.wantForbidden {
				t.Errorf("Label.CanRead() forbidden, want %v", tt.wantForbidden)
			}
			err := l.ReadOne()
			if (err != nil) != tt.wantErr {
				t.Errorf("Label.ReadOne() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("Label.ReadOne() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}
			if diff, equal := messagediff.PrettyDiff(l, tt.want); !equal && !tt.wantErr && !tt.wantForbidden {
				t.Errorf("Label.ReadAll() = %v, want %v, diff: %v", l, tt.want, diff)
			}
		})
	}
}

func TestLabel_Create(t *testing.T) {
	type fields struct {
		ID          int64
		Title       string
		Description string
		HexColor    string
		CreatedByID int64
		CreatedBy   *user.User
		Created     time.Time
		Updated     time.Time
		CRUDable    web.CRUDable
		Rights      web.Rights
	}
	type args struct {
		a web.Auth
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantErr       bool
		wantForbidden bool
	}{
		{
			name: "normal",
			fields: fields{
				Title:       "Test #1",
				Description: "Lorem Ipsum",
				HexColor:    "ffccff",
			},
			args: args{
				a: &user.User{ID: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Label{
				ID:          tt.fields.ID,
				Title:       tt.fields.Title,
				Description: tt.fields.Description,
				HexColor:    tt.fields.HexColor,
				CreatedByID: tt.fields.CreatedByID,
				CreatedBy:   tt.fields.CreatedBy,
				Created:     tt.fields.Created,
				Updated:     tt.fields.Updated,
				CRUDable:    tt.fields.CRUDable,
				Rights:      tt.fields.Rights,
			}
			allowed, _ := l.CanCreate(tt.args.a)
			if !allowed && !tt.wantForbidden {
				t.Errorf("Label.CanCreate() forbidden, want %v", tt.wantForbidden)
			}
			if err := l.Create(tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("Label.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLabel_Update(t *testing.T) {
	type fields struct {
		ID          int64
		Title       string
		Description string
		HexColor    string
		CreatedByID int64
		CreatedBy   *user.User
		Created     time.Time
		Updated     time.Time
		CRUDable    web.CRUDable
		Rights      web.Rights
	}
	tests := []struct {
		name          string
		fields        fields
		wantErr       bool
		auth          web.Auth
		wantForbidden bool
	}{
		{
			name: "normal",
			fields: fields{
				ID:    1,
				Title: "new and better",
			},
			auth: &user.User{ID: 1},
		},
		{
			name: "nonexisting",
			fields: fields{
				ID:    99999,
				Title: "new and better",
			},
			auth:          &user.User{ID: 1},
			wantForbidden: true,
			wantErr:       true,
		},
		{
			name: "no rights",
			fields: fields{
				ID:    3,
				Title: "new and better",
			},
			auth:          &user.User{ID: 1},
			wantForbidden: true,
		},
		{
			name: "no rights other creator but access",
			fields: fields{
				ID:    4,
				Title: "new and better",
			},
			auth:          &user.User{ID: 1},
			wantForbidden: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Label{
				ID:          tt.fields.ID,
				Title:       tt.fields.Title,
				Description: tt.fields.Description,
				HexColor:    tt.fields.HexColor,
				CreatedByID: tt.fields.CreatedByID,
				CreatedBy:   tt.fields.CreatedBy,
				Created:     tt.fields.Created,
				Updated:     tt.fields.Updated,
				CRUDable:    tt.fields.CRUDable,
				Rights:      tt.fields.Rights,
			}
			allowed, _ := l.CanUpdate(tt.auth)
			if !allowed && !tt.wantForbidden {
				t.Errorf("Label.CanUpdate() forbidden, want %v", tt.wantForbidden)
			}
			if err := l.Update(); (err != nil) != tt.wantErr {
				t.Errorf("Label.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLabel_Delete(t *testing.T) {
	type fields struct {
		ID          int64
		Title       string
		Description string
		HexColor    string
		CreatedByID int64
		CreatedBy   *user.User
		Created     time.Time
		Updated     time.Time
		CRUDable    web.CRUDable
		Rights      web.Rights
	}
	tests := []struct {
		name          string
		fields        fields
		wantErr       bool
		auth          web.Auth
		wantForbidden bool
	}{

		{
			name: "normal",
			fields: fields{
				ID: 1,
			},
			auth: &user.User{ID: 1},
		},
		{
			name: "nonexisting",
			fields: fields{
				ID: 99999,
			},
			auth:          &user.User{ID: 1},
			wantForbidden: true, // When the label does not exist, it is forbidden. We should fix this, but for everything.
		},
		{
			name: "no rights",
			fields: fields{
				ID: 3,
			},
			auth:          &user.User{ID: 1},
			wantForbidden: true,
		},
		{
			name: "no rights but visible",
			fields: fields{
				ID: 4,
			},
			auth:          &user.User{ID: 1},
			wantForbidden: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Label{
				ID:          tt.fields.ID,
				Title:       tt.fields.Title,
				Description: tt.fields.Description,
				HexColor:    tt.fields.HexColor,
				CreatedByID: tt.fields.CreatedByID,
				CreatedBy:   tt.fields.CreatedBy,
				Created:     tt.fields.Created,
				Updated:     tt.fields.Updated,
				CRUDable:    tt.fields.CRUDable,
				Rights:      tt.fields.Rights,
			}
			allowed, _ := l.CanDelete(tt.auth)
			if !allowed && !tt.wantForbidden {
				t.Errorf("Label.CanDelete() forbidden, want %v", tt.wantForbidden)
			}
			if err := l.Delete(); (err != nil) != tt.wantErr {
				t.Errorf("Label.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
