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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"
	"github.com/stretchr/testify/require"
)

func TestCreateUserAsAdminSkipConfirmation(t *testing.T) {
	adminActionsSetup(t)
	oldMailer := config.MailerEnabled.GetBool()
	config.MailerEnabled.Set(true)
	t.Cleanup(func() { config.MailerEnabled.Set(oldMailer) })
	notifications.Fake()
	t.Cleanup(notifications.Unfake)
	s := db.NewSession()
	defer s.Close()
	created, err := CreateUserAsAdmin(s, &user.User{ID: 1}, &CreateUserBody{APIUserPassword: user.APIUserPassword{Username: "skip-confirm", Email: "skip@example.com", Password: "12345678"}, SkipEmailConfirm: true})
	require.NoError(t, err)
	require.Equal(t, user.StatusActive, created.Status)
	notifications.AssertNotSent(t, &user.EmailConfirmNotification{})
	rs := db.NewSession()
	defer rs.Close()
	count, err := rs.Where("user_id = ? AND kind = ?", created.ID, user.TokenEmailConfirm).Count(&user.Token{})
	require.NoError(t, err)
	require.Zero(t, count)
}
