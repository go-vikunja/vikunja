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
	"gopkg.in/d4l3k/messagediff.v1"
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
		Permissions web.Permissions
	}
	type args struct {
		search string
		a      web.Auth
		page   int
	}
	user1 := &user.User{
		ID:                           1,
		Username:                     "user1",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
		ExportFileID:                 1,
	}
	user2 := &user.User{
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
	}
	user6 := &user.User{
		ID:                           6,
		Username:                     "user6",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantLs  []*LabelWithTaskID
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				a: &user.User{ID: 1},
			},
			wantLs: []*LabelWithTaskID{
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
						CreatedBy:   user2,
					},
				},
				{
					// Task 35 (project 21, owned by user1); archiving a project doesn't hide its tasks.
					Label: Label{
						ID:          5,
						Title:       "Label #5",
						CreatedByID: 2,
						CreatedBy:   user2,
						Created:     testCreatedTime,
						Updated:     testUpdatedTime,
					},
				},
				{
					Label: Label{
						ID:          7,
						Title:       "Label #7 - created by user 1, no task attachment",
						CreatedByID: 1,
						CreatedBy:   user1,
						Created:     testCreatedTime,
						Updated:     testUpdatedTime,
					},
				},
				{
					Label: Label{
						ID:          8,
						Title:       "Label #8 - user 1 creator, only attached to inaccessible task",
						CreatedByID: 1,
						CreatedBy:   user1,
						Created:     testCreatedTime,
						Updated:     testUpdatedTime,
					},
				},
				{
					// Attached to task 25 in project 16, visible via the team 1
					// share on the parent project 33.
					Label: Label{
						ID:          10,
						Title:       "Label #10 - attached in child project only",
						CreatedByID: 6,
						CreatedBy:   user6,
						Created:     testCreatedTime,
						Updated:     testUpdatedTime,
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
				Permissions: tt.fields.Permissions,
			}
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()
			gotLs, _, _, err := l.ReadAll(s, tt.args.a, tt.args.search, tt.args.page, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("Label.ReadAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := gotLs.([]*LabelWithTaskID)

			if diff, equal := messagediff.PrettyDiff(got, tt.wantLs); !equal {
				t.Errorf("Label.ReadAll() = %v, want %v, diff: %v", gotLs, tt.wantLs, diff)
			}
		})
	}
}

// Separate from the table above: its args are built before a session exists, so the share can't be loaded yet.
func TestLabel_ReadAll_LinkShare(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	share, err := GetLinkShareByID(s, 1)
	require.NoError(t, err)

	l := &Label{}
	gotLs, _, _, err := l.ReadAll(s, share, "", 0, 0)
	require.NoError(t, err)

	labels, ok := gotLs.([]*LabelWithTaskID)
	require.True(t, ok)

	ids := make([]int64, 0, len(labels))
	for _, lb := range labels {
		ids = append(ids, lb.ID)
	}
	// Label #4: on tasks #1/#2 (project 1) - visible via the share.
	// Label #5: only on project-21 tasks and soft-deleted task #51 - stays out.
	assert.ElementsMatch(t, []int64{4}, ids, "link share for project 1 must see exactly {4}; got %v", ids)
}

func TestLabel_ReadAll_SearchMatchesDescription(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	_, err := s.ID(7).Cols("description").Update(&Label{Description: "haystackneedle"})
	require.NoError(t, err)

	l := &Label{}
	gotLs, _, _, err := l.ReadAll(s, &user.User{ID: 1}, "haystackneedle", 0, 0)
	require.NoError(t, err)

	labels, ok := gotLs.([]*LabelWithTaskID)
	require.True(t, ok)
	require.Len(t, labels, 1)
	assert.Equal(t, int64(7), labels[0].ID)
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
		Permissions web.Permissions
	}
	user1 := &user.User{
		ID:                           1,
		Username:                     "user1",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
		ExportFileID:                 1,
	}
	tests := []struct {
		name                string
		fields              fields
		want                *Label
		wantErr             bool
		errType             func(error) bool
		auth                web.Auth
		wantForbidden       bool
		assertMaxPermission bool
		wantMaxPermission   int
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
			auth:                &user.User{ID: 1},
			assertMaxPermission: true,
			wantMaxPermission:   int(PermissionAdmin),
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
			name: "no permissions",
			fields: fields{
				ID: 3,
			},
			wantForbidden: true,
			auth:          &user.User{ID: 1},
		},
		{
			// Label 4 is owned by user 2; user 1 can read it via a shared task
			// but is not the owner, so max permission is read.
			name: "Get label #4 - other user",
			fields: fields{
				ID: 4,
			},
			want: &Label{
				ID:          4,
				Title:       "Label #4 - visible via other task",
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
				Created: testCreatedTime,
				Updated: testUpdatedTime,
			},
			auth:                &user.User{ID: 1},
			assertMaxPermission: true,
			wantMaxPermission:   int(PermissionRead),
		},
		{
			// PoC for GHSA-hj5c-mhh2-g7jq: label 6 is reachable only via task
			// 34 in the private project 20, user 1 must not see it.
			name: "PoC GHSA-hj5c-mhh2-g7jq: label 6 attached only to unreachable task must be forbidden",
			fields: fields{
				ID: 6,
			},
			wantForbidden: true,
			auth:          &user.User{ID: 1},
		},
		{
			// Creator of an unattached label must still be able to read it.
			name: "creator can read own label with no task attachment",
			fields: fields{
				ID: 7,
			},
			want: &Label{
				ID:          7,
				Title:       "Label #7 - created by user 1, no task attachment",
				CreatedByID: 1,
				CreatedBy:   user1,
				Created:     testCreatedTime,
				Updated:     testUpdatedTime,
			},
			auth:                &user.User{ID: 1},
			assertMaxPermission: true,
			wantMaxPermission:   int(PermissionAdmin),
		},
		{
			// Label 8's only label_tasks row points at inaccessible task 34, so
			// access comes from the creator branch; as the owner, user 1's max
			// permission is admin.
			name: "creator can read own label only attached to inaccessible task",
			fields: fields{
				ID: 8,
			},
			want: &Label{
				ID:          8,
				Title:       "Label #8 - user 1 creator, only attached to inaccessible task",
				CreatedByID: 1,
				CreatedBy:   user1,
				Created:     testCreatedTime,
				Updated:     testUpdatedTime,
			},
			auth:                &user.User{ID: 1},
			assertMaxPermission: true,
			wantMaxPermission:   int(PermissionAdmin),
		},
		{
			// Non-creator must not be able to read an unattached label owned
			// by someone else — label 3 in fixtures.
			name: "non-creator cannot read label with no task attachment",
			fields: fields{
				ID: 3,
			},
			wantForbidden: true,
			auth:          &user.User{ID: 1},
		},
		{
			// Label 10 is attached only to task 25 in project 16; user 1's
			// access comes from the team 1 share on the parent project 33.
			name: "label attached to task in child project readable via parent share",
			fields: fields{
				ID: 10,
			},
			want: &Label{
				ID:          10,
				Title:       "Label #10 - attached in child project only",
				CreatedByID: 6,
				CreatedBy: &user.User{
					ID:                           6,
					Username:                     "user6",
					Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
					Issuer:                       "local",
					EmailRemindersEnabled:        true,
					OverdueTasksRemindersEnabled: true,
					OverdueTasksRemindersTime:    "09:00",
					Created:                      testCreatedTime,
					Updated:                      testUpdatedTime,
				},
				Created: testCreatedTime,
				Updated: testUpdatedTime,
			},
			auth:                &user.User{ID: 1},
			assertMaxPermission: true,
			wantMaxPermission:   int(PermissionRead),
		},
		{
			// Label 9 was created by bot 23, whose owner is user 21. The
			// bot owner inherits admin-level access.
			name: "bot owner can read label created by their bot",
			fields: fields{
				ID: 9,
			},
			want: &Label{
				ID:          9,
				Title:       "Label #9 - created by bot 23 owned by user 21",
				CreatedByID: 23,
				CreatedBy: &user.User{
					ID:                           23,
					Name:                         "Owner A Assistant",
					Username:                     "bot-owner-a-assistant",
					Issuer:                       "local",
					BotOwnerID:                   21,
					EmailRemindersEnabled:        true,
					OverdueTasksRemindersEnabled: true,
					OverdueTasksRemindersTime:    "09:00",
					Created:                      testCreatedTime,
					Updated:                      testUpdatedTime,
				},
				Created: testCreatedTime,
				Updated: testUpdatedTime,
			},
			auth:                &user.User{ID: 21},
			assertMaxPermission: true,
			wantMaxPermission:   int(PermissionAdmin),
		},
		{
			// User 22 owns a different bot and must not see another owner's
			// bot's label.
			name: "non-owner cannot read label created by someone else's bot",
			fields: fields{
				ID: 9,
			},
			wantForbidden: true,
			auth:          &user.User{ID: 22},
		},
		{
			// Label 11 is unattached, so only the owner branch can grant this (#3592).
			name: "bot can read never-used label created by its owner",
			fields: fields{
				ID: 11,
			},
			want: &Label{
				ID:          11,
				Title:       "Label #11 - created by user 21, owner of bot 23, no task attachment",
				CreatedByID: 21,
				CreatedBy: &user.User{
					ID:                           21,
					Username:                     "user_bot_owner_a",
					Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
					Issuer:                       "local",
					EmailRemindersEnabled:        true,
					OverdueTasksRemindersEnabled: true,
					OverdueTasksRemindersTime:    "09:00",
					Created:                      testCreatedTime,
					Updated:                      testUpdatedTime,
				},
				Created: testCreatedTime,
				Updated: testUpdatedTime,
			},
			auth:                &user.User{ID: 23, BotOwnerID: 21},
			assertMaxPermission: true,
			wantMaxPermission:   int(PermissionRead),
		},
		{
			// Bot 24 belongs to user 22, so user 21's label stays out of reach.
			name: "bot cannot read label created by a different bot owner",
			fields: fields{
				ID: 11,
			},
			wantForbidden: true,
			auth:          &user.User{ID: 24, BotOwnerID: 22},
		},
		{
			// JWT auth carries no BotOwnerID, so the identity must resolve from the database.
			name: "bot can read label created by a sibling bot",
			fields: fields{
				ID: 12,
			},
			want: &Label{
				ID:          12,
				Title:       "Label #12 - created by bot 25, sibling of bot 23",
				CreatedByID: 25,
				CreatedBy: &user.User{
					ID:                           25,
					Name:                         "Owner A Scheduler",
					Username:                     "bot-owner-a-scheduler",
					Issuer:                       "local",
					BotOwnerID:                   21,
					EmailRemindersEnabled:        true,
					OverdueTasksRemindersEnabled: true,
					OverdueTasksRemindersTime:    "09:00",
					Created:                      testCreatedTime,
					Updated:                      testUpdatedTime,
				},
				Created: testCreatedTime,
				Updated: testUpdatedTime,
			},
			auth:                &user.User{ID: 23},
			assertMaxPermission: true,
			wantMaxPermission:   int(PermissionRead),
		},
		{
			name:          "other owner's bot cannot read a sibling bot's label",
			fields:        fields{ID: 12},
			wantForbidden: true,
			auth:          &user.User{ID: 24, BotOwnerID: 22},
		},
		{
			name:          "unrelated user cannot read a sibling bot's label",
			fields:        fields{ID: 12},
			wantForbidden: true,
			auth:          &user.User{ID: 1},
		},
		{
			name:          "other bot owner cannot read a sibling bot's label",
			fields:        fields{ID: 12},
			wantForbidden: true,
			auth:          &user.User{ID: 22},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
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
				Permissions: tt.fields.Permissions,
			}

			s := db.NewSession()
			defer s.Close()

			allowed, maxPermission, err := l.CanRead(s, tt.auth)
			require.NoError(t, err)
			if !allowed && !tt.wantForbidden {
				t.Errorf("Label.CanRead() forbidden, want %v", tt.wantForbidden)
			}
			if allowed && tt.wantForbidden {
				t.Errorf("Label.CanRead() allowed, want forbidden")
			}
			if tt.assertMaxPermission && maxPermission != tt.wantMaxPermission {
				t.Errorf("Label.CanRead() maxPermission = %d, want %d", maxPermission, tt.wantMaxPermission)
			}
			err = l.ReadOne(s, tt.auth)
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
			db.LoadAndAssertFixtures(t)
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
				Permissions: tt.fields.Permissions,
			}
			s := db.NewSession()
			defer s.Close()
			allowed, _ := l.CanCreate(s, tt.args.a)
			if !allowed && !tt.wantForbidden {
				t.Errorf("Label.CanCreate() forbidden, want %v", tt.wantForbidden)
			}
			if allowed && tt.wantForbidden {
				t.Errorf("Label.CanCreate() allowed, want forbidden")
			}
			if err := l.Create(s, tt.args.a); (err != nil) != tt.wantErr {
				t.Errorf("Label.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				require.NoError(t, s.Commit())
				db.AssertExists(t, "labels", map[string]interface{}{
					"id":          l.ID,
					"title":       l.Title,
					"description": l.Description,
					"hex_color":   l.HexColor,
				}, false)
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
		Permissions web.Permissions
	}
	tests := []struct {
		name              string
		fields            fields
		wantErr           bool
		auth              web.Auth
		wantForbidden     bool
		permissionErrType func(error) bool
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
			auth:              &user.User{ID: 1},
			wantForbidden:     true,
			wantErr:           true,
			permissionErrType: IsErrLabelDoesNotExist,
		},
		{
			name: "no permissions",
			fields: fields{
				ID:    3,
				Title: "new and better",
			},
			auth:          &user.User{ID: 1},
			wantForbidden: true,
		},
		{
			name: "no permissions other creator but access",
			fields: fields{
				ID:    4,
				Title: "new and better",
			},
			auth:          &user.User{ID: 1},
			wantForbidden: true,
		},
		{
			// Label 9 was created by bot 23 (owned by user 21). The bot's
			// owner inherits update permission.
			name: "bot owner can update label created by their bot",
			fields: fields{
				ID:    9,
				Title: "new and better",
			},
			auth: &user.User{ID: 21},
		},
		{
			// User 22 owns a different bot and must not be able to update
			// another owner's bot's label.
			name: "non-owner cannot update label created by someone else's bot",
			fields: fields{
				ID:    9,
				Title: "new and better",
			},
			auth:          &user.User{ID: 22},
			wantForbidden: true,
		},
		{
			name: "bot owner can update label created by their second bot",
			fields: fields{
				ID:    12,
				Title: "new and better",
			},
			auth: &user.User{ID: 21},
		},
		{
			name: "bot cannot update label created by a sibling bot",
			fields: fields{
				ID:    12,
				Title: "new and better",
			},
			auth:          &user.User{ID: 23},
			wantForbidden: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
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
				Permissions: tt.fields.Permissions,
			}
			s := db.NewSession()
			defer s.Close()
			allowed, err := l.CanUpdate(s, tt.auth)
			if tt.permissionErrType != nil {
				require.Error(t, err)
				require.True(t, tt.permissionErrType(err))
			} else {
				require.NoError(t, err)
			}
			if !allowed && !tt.wantForbidden {
				t.Errorf("Label.CanUpdate() forbidden, want %v", tt.wantForbidden)
			}
			if allowed && tt.wantForbidden {
				t.Errorf("Label.CanUpdate() allowed, want forbidden")
			}
			if tt.wantForbidden {
				return
			}
			if err := l.Update(s, tt.auth); (err != nil) != tt.wantErr {
				t.Errorf("Label.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				require.NoError(t, s.Commit())
				db.AssertExists(t, "labels", map[string]interface{}{
					"id":    tt.fields.ID,
					"title": tt.fields.Title,
				}, false)
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
		Permissions web.Permissions
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
			name: "no permissions",
			fields: fields{
				ID: 3,
			},
			auth:          &user.User{ID: 1},
			wantForbidden: true,
		},
		{
			name: "no permissions but visible",
			fields: fields{
				ID: 4,
			},
			auth:          &user.User{ID: 1},
			wantForbidden: true,
		},
		{
			// Label 9 was created by bot 23 (owned by user 21). The bot's
			// owner inherits delete permission.
			name: "bot owner can delete label created by their bot",
			fields: fields{
				ID: 9,
			},
			auth: &user.User{ID: 21},
		},
		{
			// User 22 owns a different bot and must not be able to delete
			// another owner's bot's label.
			name: "non-owner cannot delete label created by someone else's bot",
			fields: fields{
				ID: 9,
			},
			auth:          &user.User{ID: 22},
			wantForbidden: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
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
				Permissions: tt.fields.Permissions,
			}
			s := db.NewSession()
			defer s.Close()
			allowed, _ := l.CanDelete(s, tt.auth)
			if !allowed && !tt.wantForbidden {
				t.Errorf("Label.CanDelete() forbidden, want %v", tt.wantForbidden)
			}
			if allowed && tt.wantForbidden {
				t.Errorf("Label.CanDelete() allowed, want forbidden")
			}
			if tt.wantForbidden {
				return
			}
			if err := l.Delete(s, tt.auth); (err != nil) != tt.wantErr {
				t.Errorf("Label.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				require.NoError(t, s.Commit())
				db.AssertMissing(t, "labels", map[string]interface{}{
					"id": l.ID,
				})
			}
		})
	}
}

// The list and the single-label read must agree; they drifted apart before,
// letting GET /labels/{id} return a label the list hid.
func TestLabel_ReadAllMatchesCanRead(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	share, err := GetLinkShareByID(s, 1)
	require.NoError(t, err)

	var allLabels []*Label
	require.NoError(t, s.Find(&allLabels))
	require.NotEmpty(t, allLabels)

	auths := map[string]web.Auth{
		"user 1":       &user.User{ID: 1},
		"user 13":      &user.User{ID: 13},
		"user 21":      &user.User{ID: 21},
		"link share 1": share,
	}

	// Pinned independently of labelVisibleCond, so a widening of the shared
	// cond (e.g. dropping taskNotDeletedCond or the bot-identity branch)
	// can't slip through just because both sides of the parity check moved
	// together. Cross-checked against TestLabel_ReadAll and
	// TestLabel_ReadAll_LinkShare, and against labels.yml/label_tasks.yml/
	// tasks.yml/link_shares.yml.
	want := map[string][]int64{
		// Owner (1, 2, 7): unattached, readable via the creator branch.
		// Task access (4, 5, 10): attached to tasks in projects 1/21/16 that
		// user 1 owns or reaches via the team share on parent project 33.
		// Owner despite an inaccessible attachment (8): only label_tasks row
		// points at task 34 in project 20, which user 1 can't reach.
		// Excluded: 3 (other user, unattached), 6/9/11/12 (other
		// users'/bots' identities), 13 (only attached to soft-deleted task 51).
		"user 1": {1, 2, 4, 5, 7, 8, 10},
		// Owner (6, 13): user 13's own labels, readable regardless of attachment.
		// Task access (8): attached to task 34 in project 20, which user 13 owns.
		"user 13": {6, 8, 13},
		// Bot-identity branch only: 9 (bot 23), 11 (self), 12 (bot 25) all
		// resolve to owner 21 via SameBotIdentityCond. User 21's own project
		// (44) has no labeled tasks, so nothing arrives via task access.
		"user 21": {9, 11, 12},
		// Link share 1 grants read on project 1. Only tasks 1 and 2 there
		// carry a label (4); task 51 is soft-deleted and excluded, and link
		// shares don't get the bot-identity branch.
		"link share 1": {4},
	}

	for name, a := range auths {
		t.Run(name, func(t *testing.T) {
			l := &Label{}
			gotLs, _, _, err := l.ReadAll(s, a, "", 0, 0)
			require.NoError(t, err)
			labels, ok := gotLs.([]*LabelWithTaskID)
			require.True(t, ok)

			listed := make([]int64, 0, len(labels))
			for _, lb := range labels {
				listed = append(listed, lb.ID)
			}

			readable := []int64{}
			for _, lb := range allLabels {
				single := &Label{ID: lb.ID}
				can, _, err := single.CanRead(s, a)
				require.NoError(t, err)
				if can {
					readable = append(readable, lb.ID)
				}
			}

			assert.ElementsMatch(t, readable, listed, "ReadAll returned %v but CanRead allows %v", listed, readable)
			assert.ElementsMatch(t, want[name], listed, "ReadAll for %s = %v, want %v", name, listed, want[name])
		})
	}
}

// A cond that came back invalid would be dropped by the callers' builder.And,
// widening their queries to every label instead of erroring.
func TestLabelVisibleCondIsValid(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	share, err := GetLinkShareByID(s, 1)
	require.NoError(t, err)

	auths := map[string]web.Auth{
		"user":                     &user.User{ID: 1},
		"link share":               share,
		"user without any project": &user.User{ID: 17},
	}

	for name, a := range auths {
		t.Run(name, func(t *testing.T) {
			cond, err := labelVisibleCond(s, a)
			require.NoError(t, err)
			require.NotNil(t, cond)
			assert.True(t, cond.IsValid())
		})
	}
}
