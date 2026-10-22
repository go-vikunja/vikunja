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
