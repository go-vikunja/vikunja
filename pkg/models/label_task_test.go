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
	"gopkg.in/d4l3k/messagediff.v1"

	"code.vikunja.io/api/pkg/web"
)

func TestLabelTask_ReadAll(t *testing.T) {
	label := Label{
		ID:          4,
		Title:       "Label #4 - visible via other task",
		Created:     testCreatedTime,
		Updated:     testUpdatedTime,
		CreatedByID: 2,
		CreatedBy: &user.User{
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
	}

	type fields struct {
		ID          int64
		TaskID      int64
		LabelID     int64
		Created     time.Time
		CRUDable    web.CRUDable
		Permissions web.Permissions
	}
	type args struct {
		search string
		a      web.Auth
		page   int
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantLabels interface{}
		wantErr    bool
		errType    func(error) bool
	}{
		{
			name: "normal",
			fields: fields{
				TaskID: 1,
			},
			args: args{
				a: &user.User{ID: 1},
			},
			wantLabels: []*LabelWithTaskID{
				{
					TaskID: 1,
					Label:  label,
				},
			},
		},
		{
			name: "no permission to see the task",
			fields: fields{
				TaskID: 14,
			},
			args: args{
				a: &user.User{ID: 1},
			},
			wantErr: true,
			errType: IsErrNoPermissionToSeeTask,
		},
		{
			name: "nonexistant task",
			fields: fields{
				TaskID: 9999,
			},
			args: args{
				a: &user.User{ID: 1},
			},
			wantErr: true,
			errType: IsErrTaskDoesNotExist,
		},
		{
			name: "search",
			fields: fields{
				TaskID: 1,
			},
			args: args{
				a:      &user.User{ID: 1},
				search: "VISIBLE",
			},
			wantLabels: []*LabelWithTaskID{
				{
					TaskID: 1,
					Label:  label,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()

			l := &LabelTask{
				ID:          tt.fields.ID,
				TaskID:      tt.fields.TaskID,
				LabelID:     tt.fields.LabelID,
				Created:     tt.fields.Created,
				CRUDable:    tt.fields.CRUDable,
				Permissions: tt.fields.Permissions,
			}
			gotLabels, _, _, err := l.ReadAll(s, tt.args.a, tt.args.search, tt.args.page, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("LabelTask.ReadAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("LabelTask.ReadAll() Wrong error type! Error = %v, want = %v, got = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name(), err)
			}
			if diff, equal := messagediff.PrettyDiff(gotLabels, tt.wantLabels); !equal {
				t.Errorf("LabelTask.ReadAll() = %v, want %v, diff: %v", l, tt.wantLabels, diff)
			}

			s.Close()
		})
	}
}

func TestLabelTask_Create(t *testing.T) {
	type fields struct {
		ID          int64
		TaskID      int64
		LabelID     int64
		Created     time.Time
		CRUDable    web.CRUDable
		Permissions web.Permissions
	}
	type args struct {
		a web.Auth
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantErr       bool
		errType       func(error) bool
		wantForbidden bool
	}{
		{
			name: "normal",
			fields: fields{
				TaskID:  1,
				LabelID: 1,
			},
			args: args{
				a: &user.User{ID: 1},
			},
		},
		{
			name: "already existing",
			fields: fields{
				TaskID:  1,
				LabelID: 4,
			},
			args: args{
				a: &user.User{ID: 1},
			},
			wantErr: true,
			errType: IsErrLabelIsAlreadyOnTask,
		},
		{
			name: "nonexisting label",
			fields: fields{
				TaskID:  1,
				LabelID: 9999,
			},
			args: args{
				a: &user.User{ID: 1},
			},
			wantForbidden: true,
		},
		{
			name: "nonexisting task",
			fields: fields{
				TaskID:  9999,
				LabelID: 1,
			},
			args: args{
				a: &user.User{ID: 1},
			},
			wantForbidden: true,
			wantErr:       true,
			errType:       IsErrTaskDoesNotExist,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)

			s := db.NewSession()

			l := &LabelTask{
				ID:          tt.fields.ID,
				TaskID:      tt.fields.TaskID,
				LabelID:     tt.fields.LabelID,
				Created:     tt.fields.Created,
				CRUDable:    tt.fields.CRUDable,
				Permissions: tt.fields.Permissions,
			}
			allowed, err := l.CanCreate(s, tt.args.a)
			if !allowed && !tt.wantForbidden {
				t.Errorf("LabelTask.CanCreate() forbidden, want %v, err %v", tt.wantForbidden, err)
			}
			err = l.Create(s, tt.args.a)
			if (err != nil) != tt.wantErr {
				t.Errorf("LabelTask.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("LabelTask.Create() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}
			if !tt.wantErr {
				db.AssertExists(t, "label_tasks", map[string]interface{}{
					"id":       l.ID,
					"task_id":  l.TaskID,
					"label_id": l.LabelID,
				}, false)
			}
			s.Close()
		})
	}
}

func TestLabelTask_Delete(t *testing.T) {
	type fields struct {
		ID          int64
		TaskID      int64
		LabelID     int64
		Created     time.Time
		CRUDable    web.CRUDable
		Permissions web.Permissions
	}
	tests := []struct {
		name          string
		fields        fields
		wantErr       bool
		errType       func(error) bool
		auth          web.Auth
		wantForbidden bool
	}{
		{
			name: "normal",
			fields: fields{
				TaskID:  1,
				LabelID: 4,
			},
			auth: &user.User{ID: 1},
		},
		{
			name: "delete nonexistant",
			fields: fields{
				TaskID:  1,
				LabelID: 1,
			},
			auth:          &user.User{ID: 1},
			wantForbidden: true,
		},
		{
			name: "nonexisting label",
			fields: fields{
				TaskID:  1,
				LabelID: 9999,
			},
			auth:          &user.User{ID: 1},
			wantForbidden: true,
		},
		{
			name: "nonexisting task",
			fields: fields{
				TaskID:  9999,
				LabelID: 1,
			},
			auth:          &user.User{ID: 1},
			wantForbidden: true,
		},
		{
			name: "existing, but forbidden task",
			fields: fields{
				TaskID:  14,
				LabelID: 1,
			},
			auth:          &user.User{ID: 1},
			wantForbidden: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)

			s := db.NewSession()
			defer s.Close()

			l := &LabelTask{
				ID:          tt.fields.ID,
				TaskID:      tt.fields.TaskID,
				LabelID:     tt.fields.LabelID,
				Created:     tt.fields.Created,
				CRUDable:    tt.fields.CRUDable,
				Permissions: tt.fields.Permissions,
			}
			allowed, _ := l.CanDelete(s, tt.auth)
			if !allowed && !tt.wantForbidden {
				t.Errorf("LabelTask.CanDelete() forbidden, want %v", tt.wantForbidden)
			}
			if !tt.wantForbidden {
				err := l.Delete(s, tt.auth)
				if (err != nil) != tt.wantErr {
					t.Errorf("LabelTask.Delete() error = %v, wantErr %v", err, tt.wantErr)
				}
				if (err != nil) && tt.wantErr && !tt.errType(err) {
					t.Errorf("LabelTask.Delete() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
				}
				db.AssertMissing(t, "label_tasks", map[string]interface{}{
					"label_id": l.LabelID,
					"task_id":  l.TaskID,
				})
			}
		})
	}
}
