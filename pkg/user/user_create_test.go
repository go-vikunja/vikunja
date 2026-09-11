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
	"context"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/mail"
	"github.com/stretchr/testify/require"
)

func TestCreateUserConfirmationDeferred(t *testing.T) {
	i18n.Init()
	for _, tc := range []struct {
		name           string
		language       string
		welcome        string
		skip, rollback bool
	}{
		{name: "skip", skip: true},
		{name: "confirm", language: "en", welcome: "Welcome to Vikunja!"},
		{name: "localized", language: "de-DE", welcome: "Willkommen bei Vikunja!"},
		{name: "rollback", rollback: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			oldMailer := config.MailerEnabled.GetBool()
			config.MailerEnabled.Set(true)
			t.Cleanup(func() { config.MailerEnabled.Set(oldMailer) })
			mail.Fake()
			t.Cleanup(mail.ResetSent)
			events.Fake()
			s := db.NewSession()
			defer s.Close()
			defer events.CleanupPending(s)
			created, err := CreateUser(s, &User{Username: "invite-new", Email: "invite-new@example.com", Password: "12345678", Language: tc.language}, CreateUserOptions{SkipEmailConfirm: tc.skip})
			require.NoError(t, err)
			require.Nil(t, mail.LastSent())
			count, err := s.Where("user_id = ? AND kind = ?", created.ID, TokenEmailConfirm).Count(&Token{})
			require.NoError(t, err)
			if tc.skip {
				require.Equal(t, StatusActive, created.Status)
				require.Zero(t, count)
			} else {
				require.Equal(t, StatusEmailConfirmationRequired, created.Status)
				require.EqualValues(t, 1, count)
			}
			if tc.rollback {
				require.NoError(t, s.Rollback())
				events.CleanupPending(s)
				rs := db.NewSession()
				defer rs.Close()
				exists, err := rs.ID(created.ID).Exist(&User{})
				require.NoError(t, err)
				require.False(t, exists)
				n, err := rs.Where("user_id = ?", created.ID).Count(&Token{})
				require.NoError(t, err)
				require.Zero(t, n)
			} else {
				require.NoError(t, s.Commit())
			}
			if tc.skip || tc.rollback {
				require.Empty(t, mail.SentMails())
			} else {
				require.Len(t, mail.SentMails(), 1, "confirmation must be queued before Commit returns")
				sent := mail.LastSent()
				require.Equal(t, "invite-new@example.com", sent.To)
				require.Contains(t, sent.Message, tc.welcome)
				require.Contains(t, sent.Message, "userEmailConfirm=")
			}
			events.DispatchPending(context.Background(), s)
			if !tc.rollback {
				require.Equal(t, 1, events.CountDispatchedEvents((&CreatedEvent{}).Name()))
			}
		})
	}
}
