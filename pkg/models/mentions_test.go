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
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindMentionedUsersInText(t *testing.T) {
	user1 := &user.User{
		ID: 1,
	}
	user2 := &user.User{
		ID: 2,
	}

	tests := []struct {
		name      string
		text      string
		wantUsers []*user.User
		wantErr   bool
	}{
		{
			name: "no users mentioned",
			text: "Lorem Ipsum dolor sit amet",
		},
		{
			name:      "one user at the beginning",
			text:      "@user1 Lorem Ipsum",
			wantUsers: []*user.User{user1},
		},
		{
			name:      "one user at the end",
			text:      "Lorem Ipsum @user1",
			wantUsers: []*user.User{user1},
		},
		{
			name:      "one user in the middle",
			text:      "Lorem @user1 Ipsum",
			wantUsers: []*user.User{user1},
		},
		{
			name:      "same user multiple times",
			text:      "Lorem @user1 Ipsum @user1 @user1",
			wantUsers: []*user.User{user1},
		},
		{
			name:      "Multiple users",
			text:      "Lorem @user1 Ipsum @user2",
			wantUsers: []*user.User{user1, user2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			gotUsers, err := FindMentionedUsersInText(s, tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindMentionedUsersInText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for _, u := range tt.wantUsers {
				_, has := gotUsers[u.ID]
				if !has {
					t.Errorf("wanted user %d but did not get it", u.ID)
				}
			}
		})
	}
}

func TestSendingMentionNotification(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("should send notifications to all users having access", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task, err := GetTaskByIDSimple(s, 32)
		require.NoError(t, err)
		tc := &TaskComment{
			Comment: "Lorem Ipsum @user1 @user2 @user3 @user4 @user5 @user6",
			TaskID:  32, // user2 has access to the project that task belongs to
		}
		err = tc.Create(s, u)
		require.NoError(t, err)
		n := &TaskCommentNotification{
			Doer:    u,
			Task:    &task,
			Comment: tc,
		}

		_, err = notifyMentionedUsers(s, &task, tc.Comment, n)
		require.NoError(t, err)

		db.AssertExists(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 1,
			"name":          n.Name(),
		}, false)
		db.AssertExists(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 2,
			"name":          n.Name(),
		}, false)
		db.AssertExists(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 3,
			"name":          n.Name(),
		}, false)
		db.AssertMissing(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 4,
			"name":          n.Name(),
		})
		db.AssertMissing(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 5,
			"name":          n.Name(),
		})
		db.AssertMissing(t, "notifications", map[string]interface{}{
			"subject_id":    tc.ID,
			"notifiable_id": 6,
			"name":          n.Name(),
		})
	})
	t.Run("should not send notifications multiple times", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		task, err := GetTaskByIDSimple(s, 32)
		require.NoError(t, err)
		tc := &TaskComment{
			Comment: "Lorem Ipsum @user2",
			TaskID:  32, // user2 has access to the project that task belongs to
		}
		err = tc.Create(s, u)
		require.NoError(t, err)
		n := &TaskCommentNotification{
			Doer:    u,
			Task:    &task,
			Comment: tc,
		}

		_, err = notifyMentionedUsers(s, &task, tc.Comment, n)
		require.NoError(t, err)

		_, err = notifyMentionedUsers(s, &task, "Lorem Ipsum @user2 @user3", n)
		require.NoError(t, err)

		// The second time mentioning the user in the same task should not create another notification
		dbNotifications, err := notifications.GetNotificationsForNameAndUser(s, 2, n.Name(), tc.ID)
		require.NoError(t, err)
		assert.Len(t, dbNotifications, 1)
	})
}
