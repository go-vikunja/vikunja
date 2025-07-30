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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"
	"gopkg.in/d4l3k/messagediff.v1"
)

func TestListUsersFromProject(t *testing.T) {
	testuser1 := &user.User{
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
	testuser2 := &user.User{
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
	testuser3 := &user.User{
		ID:                           3,
		Username:                     "user3",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		DefaultProjectID:             4,
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	testuser4 := &user.User{
		ID:                           4,
		Username:                     "user4",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Status:                       user.StatusEmailConfirmationRequired,
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	testuser5 := &user.User{
		ID:                           5,
		Username:                     "user5",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Status:                       user.StatusEmailConfirmationRequired,
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	testuser6 := &user.User{
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
	testuser7 := &user.User{
		ID:                           7,
		Username:                     "user7",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		DiscoverableByEmail:          true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	testuser8 := &user.User{
		ID:                           8,
		Username:                     "user8",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	testuser9 := &user.User{
		ID:                           9,
		Username:                     "user9",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	testuser10 := &user.User{
		ID:                           10,
		Username:                     "user10",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	testuser11 := &user.User{
		ID:                           11,
		Username:                     "user11",
		Name:                         "Some one else",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	testuser12 := &user.User{
		ID:                           12,
		Username:                     "user12",
		Name:                         "Name with spaces",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		DiscoverableByName:           true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}
	testuser13 := &user.User{
		ID:                           13,
		Username:                     "user13",
		Password:                     "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6",
		Issuer:                       "local",
		EmailRemindersEnabled:        true,
		OverdueTasksRemindersEnabled: true,
		OverdueTasksRemindersTime:    "09:00",
		Created:                      testCreatedTime,
		Updated:                      testUpdatedTime,
	}

	type args struct {
		l      *Project
		search string
	}
	tests := []struct {
		name      string
		args      args
		wantUsers []*user.User
		wantErr   bool
	}{
		{
			name:      "Check owner only",
			args:      args{l: &Project{ID: 18, OwnerID: 7}},
			wantUsers: []*user.User{testuser7},
		},
		{
			// This project has another different user shared for each possible method
			name: "Check with owner and other users",
			args: args{l: &Project{ID: 19, OwnerID: 7}},
			wantUsers: []*user.User{
				testuser1, // Shared Via Team readonly
				testuser2, // Shared Via Team write
				testuser3, // Shared Via Team admin

				testuser4, // Shared Via User readonly
				testuser5, // Shared Via User write
				testuser6, // Shared Via User admin

				testuser7, // Owner

				testuser8,  // Shared Via Parent Project Team readonly
				testuser9,  // Shared Via Parent Project Team write
				testuser10, // Shared Via Parent Project Team admin

				testuser11, // Shared Via Parent Project User readonly
				testuser12, // Shared Via Parent Project User write
				testuser13, // Shared Via Parent Project User admin
			},
		},
		{
			name: "search for user1",
			args: args{l: &Project{ID: 19, OwnerID: 7}, search: "user1"},
			wantUsers: []*user.User{
				testuser1, // Shared Via Team readonly

				testuser10, // Matches Partially, Shared Via NamespaceTeam admin
				testuser11, // Matches Partially, Shared Via NamespaceUser readonly
				testuser12, // Matches Partially, Shared Via NamespaceUser write
				testuser13, // Matches Partially, Shared Via NamespaceUser admin
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			gotUsers, err := ListUsersFromProject(s, tt.args.l, testuser1, tt.args.search)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListUsersFromProject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff, equal := messagediff.PrettyDiff(tt.wantUsers, gotUsers); !equal {
				t.Errorf("Test %s, LabelTask.ReadAll() = %v, want %v, \ndiff: %v", tt.name, gotUsers, tt.wantUsers, diff)
			}
		})
	}
}
