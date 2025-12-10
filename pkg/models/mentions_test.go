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
			text: "<p>Lorem Ipsum dolor sit amet</p>",
		},
		{
			name:      "one user at the beginning",
			text:      `<p><mention-user data-id="user1">@user1</mention-user> Lorem Ipsum</p>`,
			wantUsers: []*user.User{user1},
		},
		{
			name:      "one user at the end",
			text:      `<p>Lorem Ipsum <mention-user data-id="user1">@user1</mention-user></p>`,
			wantUsers: []*user.User{user1},
		},
		{
			name:      "one user in the middle",
			text:      `<p>Lorem <mention-user data-id="user1">@user1</mention-user> Ipsum</p>`,
			wantUsers: []*user.User{user1},
		},
		{
			name:      "same user multiple times",
			text:      `<p>Lorem <mention-user data-id="user1">@user1</mention-user> Ipsum <mention-user data-id="user1">@user1</mention-user> <mention-user data-id="user1">@user1</mention-user></p>`,
			wantUsers: []*user.User{user1},
		},
		{
			name:      "Multiple users",
			text:      `<p>Lorem <mention-user data-id="user1">@user1</mention-user> Ipsum <mention-user data-id="user2">@user2</mention-user></p>`,
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
			Comment: `<p>Lorem Ipsum <mention-user data-id="user1">@user1</mention-user> <mention-user data-id="user2">@user2</mention-user> <mention-user data-id="user3">@user3</mention-user> <mention-user data-id="user4">@user4</mention-user> <mention-user data-id="user5">@user5</mention-user> <mention-user data-id="user6">@user6</mention-user></p>`,
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
			Comment: `<p>Lorem Ipsum <mention-user data-id="user2">@user2</mention-user></p>`,
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

		_, err = notifyMentionedUsers(s, &task, `<p>Lorem Ipsum <mention-user data-id="user2">@user2</mention-user> <mention-user data-id="user3">@user3</mention-user></p>`, n)
		require.NoError(t, err)

		// The second time mentioning the user in the same task should not create another notification
		dbNotifications, err := notifications.GetNotificationsForNameAndUser(s, 2, n.Name(), tc.ID)
		require.NoError(t, err)
		assert.Len(t, dbNotifications, 1)
	})
}

func TestFormatMentionsForEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no mentions",
			input:    "<p>Lorem Ipsum dolor sit amet</p>",
			expected: "<p>Lorem Ipsum dolor sit amet</p>",
		},
		{
			name:     "single mention with data-label (new format)",
			input:    `<p><mention-user data-id="frederick" data-label="Frederick" data-mention-suggestion-char="@"></mention-user> hello</p>`,
			expected: `<p><strong>@Frederick</strong> hello</p>`,
		},
		{
			name:     "single mention with full name in data-label",
			input:    `<p><mention-user data-id="johndoe" data-label="John Doe" data-mention-suggestion-char="@"></mention-user> please help</p>`,
			expected: `<p><strong>@John Doe</strong> please help</p>`,
		},
		{
			name:     "mention without data-label (fallback to data-id)",
			input:    `<p><mention-user data-id="johndoe"></mention-user> test</p>`,
			expected: `<p><strong>@johndoe</strong> test</p>`,
		},
		{
			name:     "old format with text node inside",
			input:    `<p><mention-user data-id="user1">@user1</mention-user> Lorem Ipsum</p>`,
			expected: `<p><strong>@user1</strong> Lorem Ipsum</p>`,
		},
		{
			name:     "old format with text node (data-id takes precedence over text)",
			input:    `<p><mention-user data-id="actualuser">@differentuser</mention-user> text</p>`,
			expected: `<p><strong>@actualuser</strong> text</p>`,
		},
		{
			name:     "multiple mentions in one paragraph",
			input:    `<p>Hey <mention-user data-id="john" data-label="John"></mention-user> and <mention-user data-id="jane" data-label="Jane Doe"></mention-user>, please review</p>`,
			expected: `<p>Hey <strong>@John</strong> and <strong>@Jane Doe</strong>, please review</p>`,
		},
		{
			name:     "mention at beginning",
			input:    `<p><mention-user data-id="user1" data-label="User One"></mention-user> Lorem Ipsum</p>`,
			expected: `<p><strong>@User One</strong> Lorem Ipsum</p>`,
		},
		{
			name:     "mention at end",
			input:    `<p>Lorem Ipsum <mention-user data-id="user1" data-label="User One"></mention-user></p>`,
			expected: `<p>Lorem Ipsum <strong>@User One</strong></p>`,
		},
		{
			name:     "mention in middle",
			input:    `<p>Lorem <mention-user data-id="user1" data-label="User One"></mention-user> Ipsum</p>`,
			expected: `<p>Lorem <strong>@User One</strong> Ipsum</p>`,
		},
		{
			name:     "same user mentioned multiple times",
			input:    `<p><mention-user data-id="user1" data-label="User"></mention-user> and <mention-user data-id="user1" data-label="User"></mention-user> again</p>`,
			expected: `<p><strong>@User</strong> and <strong>@User</strong> again</p>`,
		},
		{
			name:     "HTML preservation with links",
			input:    `<p>Check <a href="http://example.com">this link</a> and ask <mention-user data-id="expert" data-label="Expert"></mention-user></p>`,
			expected: `<p>Check <a href="http://example.com">this link</a> and ask <strong>@Expert</strong></p>`,
		},
		{
			name:     "HTML preservation with multiple paragraphs",
			input:    `<p>First paragraph with <mention-user data-id="user1" data-label="User"></mention-user></p><p>Second paragraph</p>`,
			expected: `<p>First paragraph with <strong>@User</strong></p><p>Second paragraph</p>`,
		},
		{
			name:     "HTML preservation with bold and italic",
			input:    `<p><strong>Bold text</strong> and <em>italic</em> with <mention-user data-id="user1" data-label="User"></mention-user></p>`,
			expected: `<p><strong>Bold text</strong> and <em>italic</em> with <strong>@User</strong></p>`,
		},
		{
			name:     "special characters in data-label",
			input:    `<p><mention-user data-id="user1" data-label="O'Brien"></mention-user> test</p>`,
			expected: `<p><strong>@O&#39;Brien</strong> test</p>`,
		},
		{
			name:     "special characters - ampersand in data-label",
			input:    `<p><mention-user data-id="user1" data-label="Tom &amp; Jerry"></mention-user> test</p>`,
			expected: `<p><strong>@Tom &amp; Jerry</strong> test</p>`,
		},
		{
			name:     "special characters - quotes in data-label",
			input:    `<p><mention-user data-id="user1" data-label="&quot;Nickname&quot;"></mention-user> test</p>`,
			expected: `<p><strong>@&#34;Nickname&#34;</strong> test</p>`,
		},
		{
			name:     "mixed old and new format",
			input:    `<p><mention-user data-id="new" data-label="New User"></mention-user> and <mention-user data-id="old">@old</mention-user></p>`,
			expected: `<p><strong>@New User</strong> and <strong>@old</strong></p>`,
		},
		{
			name:     "self-closing tag format (XML-style)",
			input:    `<p><mention-user data-id="user" data-label="User"/> hello</p>`,
			expected: `<p><strong>@User</strong></p>`,
		},
		{
			name:     "mention with only text content (no attributes) - old format edge case",
			input:    `<p><mention-user>@someuser</mention-user> test</p>`,
			expected: `<p><strong>@someuser</strong> test</p>`,
		},
		{
			name:     "data-label takes precedence over data-id",
			input:    `<p><mention-user data-id="username123" data-label="John Smith"></mention-user> test</p>`,
			expected: `<p><strong>@John Smith</strong> test</p>`,
		},
		{
			name:     "unicode characters in data-label",
			input:    `<p><mention-user data-id="user" data-label="Müller François"></mention-user> test</p>`,
			expected: `<p><strong>@Müller François</strong> test</p>`,
		},
		{
			name:     "emoji in data-label",
			input:    `<p><mention-user data-id="user" data-label="Cool User 😎"></mention-user> test</p>`,
			expected: `<p><strong>@Cool User 😎</strong> test</p>`,
		},
		{
			name:     "nested HTML structure",
			input:    `<div><p>Text with <mention-user data-id="user" data-label="User"></mention-user> in div</p></div>`,
			expected: `<div><p>Text with <strong>@User</strong> in div</p></div>`,
		},
		{
			name:     "mention in list",
			input:    `<ul><li>Item with <mention-user data-id="user" data-label="User"></mention-user></li></ul>`,
			expected: `<ul><li>Item with <strong>@User</strong></li></ul>`,
		},
		{
			name:     "very long name",
			input:    `<p><mention-user data-id="user" data-label="Christopher Montgomery Bartholomew Johnson-Smith III"></mention-user> test</p>`,
			expected: `<p><strong>@Christopher Montgomery Bartholomew Johnson-Smith III</strong> test</p>`,
		},
		{
			name:     "empty data-label and data-id with text content",
			input:    `<p><mention-user>@fallback</mention-user> test</p>`,
			expected: `<p><strong>@fallback</strong> test</p>`,
		},
		{
			name:     "whitespace in data-label",
			input:    `<p><mention-user data-id="user" data-label="  Spaces  "></mention-user> test</p>`,
			expected: `<p><strong>@  Spaces  </strong> test</p>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			result := formatMentionsForEmail(s, tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatMentionsForEmail_MalformedHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "unclosed tag - returns original",
			input: `<p>Test <mention-user data-id="user" data-label="User">`,
		},
		{
			name:  "invalid HTML entities",
			input: `<p>Test &invalid; entity</p>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			result := formatMentionsForEmail(s, tt.input)
			// For malformed HTML, we expect it to either be fixed by the parser or returned as-is
			// The key is that it shouldn't panic or error
			assert.NotEmpty(t, result)
		})
	}
}
