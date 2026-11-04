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

package user

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/notifications"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationLinksKeepTokensOutOfTheQuery(t *testing.T) {
	u := &User{Username: "user1"}
	tests := map[string]struct {
		notification interface {
			ToMail(lang string) *notifications.Mail
		}
		param string
	}{
		"email confirm":    {&EmailConfirmNotification{User: u, ConfirmToken: "secrettoken"}, "userEmailConfirm"},
		"password reset":   {&ResetPasswordNotification{User: u, Token: &Token{ClearTextToken: "secrettoken"}}, "userPasswordReset"},
		"account deletion": {&AccountDeletionConfirmNotification{User: u, ConfirmToken: "secrettoken"}, "accountDeletionConfirm"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			opts, err := notifications.RenderMail(tc.notification.ToMail("en"), "en")
			require.NoError(t, err)

			for _, body := range []string{opts.Message, opts.HTMLMessage} {
				assert.Contains(t, body, "#"+tc.param+"=secrettoken")
				assert.NotContains(t, body, "?"+tc.param+"=")
			}
		})
	}
}

func TestShouldNotify(t *testing.T) {
	tests := map[string]struct {
		userID       int64
		notification notifications.Notification
		want         bool
	}{
		"active user":                        {userID: 1, notification: &PasswordChangedNotification{}, want: true},
		"locked user, ordinary notification": {userID: 18, notification: &PasswordChangedNotification{}, want: false},
		"locked user, password reset":        {userID: 18, notification: &ResetPasswordNotification{}, want: true},
		"disabled user, password reset":      {userID: 17, notification: &ResetPasswordNotification{}, want: false},
		"bot user, password reset":           {userID: 23, notification: &ResetPasswordNotification{}, want: false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			u, err := loadUser(s, &User{ID: tt.userID}, false)
			require.NoError(t, err)

			got, err := u.ShouldNotify(tt.notification, s)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIdentityLinkedNotification(t *testing.T) {
	i18n.Init()
	n := &IdentityLinkedNotification{
		User: &User{
			Username: "user1",
			Timezone: "Europe/Berlin",
		},
		ProviderName: "Authentik",
		Linked:       time.Date(2026, 11, 4, 9, 30, 0, 0, time.UTC),
	}

	mail := n.ToMail("en")
	opts, err := notifications.RenderMail(mail, "en")
	require.NoError(t, err)

	assert.Equal(t, "Authentik was linked to your Vikunja account", opts.Subject)
	assert.Contains(t, opts.Message, "Your Authentik account was linked to your Vikunja account on 2026-11-04 10:30 CET.")
	assert.Contains(t, opts.Message, "change your password immediately, then unlink it in your account settings")
	assert.Contains(t, opts.Message, "user/settings/linked-accounts")
}
