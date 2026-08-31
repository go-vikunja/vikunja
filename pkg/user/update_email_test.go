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
	"regexp"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

var emailConfirmTokenRegex = regexp.MustCompile(`\?userEmailConfirm=([^"'\s<&]+)`)

// confirmTokenFromSentMail returns the clear text token from the confirmation link of the
// sent mails. It is only available there, the db only holds its hash.
func confirmTokenFromSentMail(t *testing.T) string {
	for _, sent := range mail.SentMails() {
		matches := emailConfirmTokenRegex.FindStringSubmatch(sent.Message + sent.HTMLMessage)
		if len(matches) == 2 {
			return matches[1]
		}
	}
	t.Fatal("no mail with a confirmation link was sent")
	return ""
}

func resetEmailConfirmationCooldown(t *testing.T, userID int64) {
	require.NoError(t, keyvalue.Del(emailConfirmationCooldownKey(userID)))
}

func backdateToken(t *testing.T, s *xorm.Session, tokenID int64, created time.Time) {
	_, err := s.Where("id = ?", tokenID).Cols("created").Update(&Token{Created: created})
	require.NoError(t, err)
}

func TestUpdateEmail(t *testing.T) {
	t.Run("mailer enabled keeps old email until confirmed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		config.MailerEnabled.Set(true)
		defer config.MailerEnabled.Set(false)

		resetEmailConfirmationCooldown(t, 1)
		mail.ResetSent()
		err := UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "new1@example.com"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		sent := mail.SentMails()
		require.Len(t, sent, 2)
		assert.Equal(t, "new1@example.com", sent[0].To)

		s2 := db.NewSession()
		defer s2.Close()
		updated, err := GetUserWithEmail(s2, &User{ID: 1})
		require.NoError(t, err)
		assert.Equal(t, StatusActive, updated.Status)
		assert.Equal(t, "user1@example.com", updated.Email)
		assert.Equal(t, "new1@example.com", updated.PendingEmail)

		tokens, err := getTokensForKind(s2, updated, TokenEmailConfirm)
		require.NoError(t, err)
		assert.Len(t, tokens, 1)

		// User can still log in with the old address.
		_, err = CheckUserCredentials(context.Background(), s2, &Login{Username: "user1@example.com", Password: "12345678"})
		require.NoError(t, err)
	})

	t.Run("mailer enabled second change replaces pending", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		config.MailerEnabled.Set(true)
		defer config.MailerEnabled.Set(false)

		resetEmailConfirmationCooldown(t, 1)
		require.NoError(t, UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "typo@example.com"}))
		resetEmailConfirmationCooldown(t, 1)
		require.NoError(t, UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "fixed@example.com"}))
		require.NoError(t, s.Commit())

		s2 := db.NewSession()
		defer s2.Close()
		updated, err := GetUserWithEmail(s2, &User{ID: 1})
		require.NoError(t, err)
		assert.Equal(t, "fixed@example.com", updated.PendingEmail)

		tokens, err := getTokensForKind(s2, updated, TokenEmailConfirm)
		require.NoError(t, err)
		assert.Len(t, tokens, 1)
	})

	t.Run("mailer disabled applies immediately", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		config.MailerEnabled.Set(false)

		err := UpdateEmail(s, &EmailUpdate{User: &User{ID: 2}, NewEmail: "new2@example.com"})
		require.NoError(t, err)
		require.NoError(t, s.Commit())

		s2 := db.NewSession()
		defer s2.Close()
		updated, err := GetUserWithEmail(s2, &User{ID: 2})
		require.NoError(t, err)
		assert.Equal(t, StatusActive, updated.Status)
		assert.Equal(t, "new2@example.com", updated.Email)
		assert.Empty(t, updated.PendingEmail)
	})

	t.Run("email already taken", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "user2@example.com"})
		require.Error(t, err)
		assert.True(t, IsErrUserEmailExists(err))
	})

	t.Run("own email rejected", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "user1@example.com"})
		require.Error(t, err)
		assert.True(t, IsErrUserEmailExists(err))
	})

	t.Run("second change within cooldown", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		config.MailerEnabled.Set(true)
		defer config.MailerEnabled.Set(false)

		resetEmailConfirmationCooldown(t, 1)
		require.NoError(t, UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "new1@example.com"}))

		err := UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "other1@example.com"})
		require.Error(t, err)
		assert.True(t, IsErrEmailConfirmationCooldown(err))
	})

	t.Run("cooldown survives cancel", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		config.MailerEnabled.Set(true)
		defer config.MailerEnabled.Set(false)

		resetEmailConfirmationCooldown(t, 2)
		require.NoError(t, UpdateEmail(s, &EmailUpdate{User: &User{ID: 2}, NewEmail: "new2@example.com"}))
		require.NoError(t, CancelEmailUpdate(s, &User{ID: 2}))

		err := UpdateEmail(s, &EmailUpdate{User: &User{ID: 2}, NewEmail: "other2@example.com"})
		require.Error(t, err)
		assert.True(t, IsErrEmailConfirmationCooldown(err))
	})

	t.Run("notifies the current address", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		config.MailerEnabled.Set(true)
		defer config.MailerEnabled.Set(false)

		resetEmailConfirmationCooldown(t, 1)
		mail.ResetSent()
		require.NoError(t, UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "new1@example.com"}))

		sent := mail.SentMails()
		require.Len(t, sent, 2)
		assert.Equal(t, "new1@example.com", sent[0].To)
		assert.Equal(t, "user1@example.com", sent[1].To)
	})
}

func TestCancelEmailUpdate(t *testing.T) {
	t.Run("discards pending change", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		config.MailerEnabled.Set(true)
		defer config.MailerEnabled.Set(false)

		resetEmailConfirmationCooldown(t, 1)
		require.NoError(t, UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "new1@example.com"}))
		require.NoError(t, CancelEmailUpdate(s, &User{ID: 1}))
		require.NoError(t, s.Commit())

		s2 := db.NewSession()
		defer s2.Close()
		updated, err := GetUserWithEmail(s2, &User{ID: 1})
		require.NoError(t, err)
		assert.Equal(t, "user1@example.com", updated.Email)
		assert.Empty(t, updated.PendingEmail)

		tokens, err := getTokensForKind(s2, updated, TokenEmailConfirm)
		require.NoError(t, err)
		assert.Empty(t, tokens)
	})

	t.Run("nothing pending", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := CancelEmailUpdate(s, &User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrNoPendingEmail(err))
	})
}

func TestResendEmailConfirmation(t *testing.T) {
	t.Run("nothing pending", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := ResendEmailConfirmation(s, &User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrNoPendingEmail(err))
	})

	t.Run("pending rotates token", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		config.MailerEnabled.Set(true)
		defer config.MailerEnabled.Set(false)

		resetEmailConfirmationCooldown(t, 1)
		require.NoError(t, UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "new1@example.com"}))
		before, err := getTokensForKind(s, &User{ID: 1}, TokenEmailConfirm)
		require.NoError(t, err)
		require.Len(t, before, 1)
		resetEmailConfirmationCooldown(t, 1)

		mail.ResetSent()
		require.NoError(t, ResendEmailConfirmation(s, &User{ID: 1}))
		after, err := getTokensForKind(s, &User{ID: 1}, TokenEmailConfirm)
		require.NoError(t, err)
		require.Len(t, after, 1)
		assert.NotEqual(t, before[0].Token, after[0].Token)

		sent := mail.LastSent()
		require.NotNil(t, sent)
		assert.Equal(t, "new1@example.com", sent.To)
	})

	t.Run("within cooldown", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		config.MailerEnabled.Set(true)
		defer config.MailerEnabled.Set(false)

		resetEmailConfirmationCooldown(t, 1)
		require.NoError(t, UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "new1@example.com"}))
		before, err := getTokensForKind(s, &User{ID: 1}, TokenEmailConfirm)
		require.NoError(t, err)
		require.Len(t, before, 1)

		mail.ResetSent()
		err = ResendEmailConfirmation(s, &User{ID: 1})
		require.Error(t, err)
		assert.True(t, IsErrEmailConfirmationCooldown(err))
		assert.Nil(t, mail.LastSent())

		after, err := getTokensForKind(s, &User{ID: 1}, TokenEmailConfirm)
		require.NoError(t, err)
		require.Len(t, after, 1)
		assert.Equal(t, before[0].Token, after[0].Token)
	})
}
