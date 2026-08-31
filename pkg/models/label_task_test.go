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

	"github.com/stretchr/testify/require"
	"gopkg.in/d4l3k/messagediff.v1"
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
			defer s.Close()

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
			wantErr:       true,
			errType:       IsErrLabelDoesNotExist,
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
		{
			// Label 10 is attached only to task 25 in project 16, a child of the
			// team-shared project 33. Task 26 lives in project 17, a child of the
			// team-shared project 34. User 1 has no direct share on either child —
			// both label access and task write are inherited through the parents.
			name: "label and task access inherited via parent project",
			fields: fields{
				TaskID:  26,
				LabelID: 10,
			},
			args: args{
				a: &user.User{ID: 1},
			},
		},
		{
			// Task 1 is writable by user 1, but label 6 is user 13's private
			// label — write access to the task must not grant label access.
			name: "writable task but inaccessible label",
			fields: fields{
				TaskID:  1,
				LabelID: 6,
			},
			args: args{
				a: &user.User{ID: 1},
			},
			wantForbidden: true,
		},
		{
			// Label 11 has no label_tasks row, so only the owner branch can
			// grant this (#3592).
			name: "bot can attach a never-used label created by its owner",
			fields: fields{
				TaskID:  52,
				LabelID: 11,
			},
			args: args{
				a: &user.User{ID: 23},
			},
		},
		{
			// Same writable task, but label 6 belongs to user 13 — inheriting the
			// owner's labels must not widen access to anyone else's.
			name: "bot cannot attach a label unrelated to its owner",
			fields: fields{
				TaskID:  52,
				LabelID: 6,
			},
			args: args{
				a: &user.User{ID: 23},
			},
			wantForbidden: true,
		},
		{
			// Bot 23 can see label 11 (owned by its owner, user 21), but has no
			// share on task 1's project, isolating the task-write conjunct.
			name: "bot cannot attach its owner's label to an unwritable task",
			fields: fields{
				TaskID:  1,
				LabelID: 11,
			},
			args: args{
				a: &user.User{ID: 23},
			},
			wantForbidden: true,
		},
		{
			name: "bot can attach a label created by a sibling bot",
			fields: fields{
				TaskID:  52,
				LabelID: 12,
			},
			args: args{
				a: &user.User{ID: 23},
			},
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
			allowed, err := l.CanCreate(s, tt.args.a)
			if !allowed && !tt.wantForbidden {
				t.Errorf("LabelTask.CanCreate() forbidden, want %v, err %v", tt.wantForbidden, err)
			}
			if allowed && tt.wantForbidden {
				t.Errorf("LabelTask.CanCreate() allowed, want forbidden")
			}
			if tt.wantForbidden {
				if tt.wantErr && !tt.errType(err) {
					t.Errorf("LabelTask.CanCreate() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
				}
				return
			}
			err = l.Create(s, tt.args.a)
			if (err != nil) != tt.wantErr {
				t.Errorf("LabelTask.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) && tt.wantErr && !tt.errType(err) {
				t.Errorf("LabelTask.Create() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
			}
			if !tt.wantErr {
				require.NoError(t, s.Commit())
				db.AssertExists(t, "label_tasks", map[string]interface{}{
					"id":       l.ID,
					"task_id":  l.TaskID,
					"label_id": l.LabelID,
				}, false)
			}
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
			if allowed && tt.wantForbidden {
				t.Errorf("LabelTask.CanDelete() allowed, want forbidden")
			}
			if !tt.wantForbidden {
				err := l.Delete(s, tt.auth)
				if (err != nil) != tt.wantErr {
					t.Errorf("LabelTask.Delete() error = %v, wantErr %v", err, tt.wantErr)
				}
				if (err != nil) && tt.wantErr && !tt.errType(err) {
					t.Errorf("LabelTask.Delete() Wrong error type! Error = %v, want = %v", err, runtime.FuncForPC(reflect.ValueOf(tt.errType).Pointer()).Name())
				}
				require.NoError(t, s.Commit())
				db.AssertMissing(t, "label_tasks", map[string]interface{}{
					"label_id": l.LabelID,
					"task_id":  l.TaskID,
				})
			}
		})
	}
}

// Label changes must advance both the task's and the project's updated
// timestamps so CalDAV delta syncs and ctags pick up CATEGORIES changes.
func TestLabelTaskUpdatedTimestamps(t *testing.T) {
	readTimes := func(t *testing.T) (task Task, project *Project) {
		s := db.NewSession()
		defer s.Close()
		task, err := GetTaskByIDSimple(s, 1)
		require.NoError(t, err)
		project, err = GetProjectSimpleByID(s, 1)
		require.NoError(t, err)
		return
	}

	t.Run("create", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		taskBefore, projectBefore := readTimes(t)

		s := db.NewSession()
		defer s.Close()
		lt := &LabelTask{TaskID: 1, LabelID: 1}
		require.NoError(t, lt.Create(s, &user.User{ID: 1}))
		require.NoError(t, s.Commit())

		taskAfter, projectAfter := readTimes(t)
		require.True(t, taskAfter.Updated.After(taskBefore.Updated), "task updated time must advance")
		require.True(t, projectAfter.Updated.After(projectBefore.Updated), "project updated time must advance")
	})

	t.Run("delete", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		taskBefore, projectBefore := readTimes(t)

		s := db.NewSession()
		defer s.Close()
		lt := &LabelTask{TaskID: 1, LabelID: 4}
		require.NoError(t, lt.Delete(s, &user.User{ID: 1}))
		require.NoError(t, s.Commit())

		taskAfter, projectAfter := readTimes(t)
		require.True(t, taskAfter.Updated.After(taskBefore.Updated), "task updated time must advance")
		require.True(t, projectAfter.Updated.After(projectBefore.Updated), "project updated time must advance")
	})
}

func TestLabelTaskBulk_CreateLinkShare(t *testing.T) {
	// Link share 2 has write on project 2, so it clears the task check and
	// reaches the per-label access check. Label 1 is on no task it can see.
	db.LoadAndAssertFixtures(t)

	s := db.NewSession()
	defer s.Close()

	share := &LinkSharing{ID: 2, Hash: "test2", ProjectID: 2, Permission: PermissionWrite}
	ltb := &LabelTaskBulk{TaskID: 13, Labels: []*Label{{ID: 1}}}

	allowed, err := ltb.CanCreate(s, share)
	require.NoError(t, err)
	require.True(t, allowed, "write link share must pass the task check")

	err = ltb.Create(s, share)
	require.Error(t, err)
	require.True(t, IsErrUserHasNoAccessToLabel(err), "got %#v", err)
}
