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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/mail"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserEmailConfirm(t *testing.T) {
	type args struct {
		c *EmailConfirm
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		errType func(error) bool
	}{
		{
			name: "Test Empty token",
			args: args{
				c: &EmailConfirm{
					Token: "",
				},
			},
			wantErr: true,
			errType: IsErrInvalidEmailConfirmToken,
		},
		{
			name: "Test invalid token",
			args: args{
				c: &EmailConfirm{
					Token: "invalid",
				},
			},
			wantErr: true,
			errType: IsErrInvalidEmailConfirmToken,
		},
		{
			name: "Test valid token",
			args: args{
				c: &EmailConfirm{
					Token: "tiepiQueed8ahc7zeeFe1eveiy4Ein8osooxegiephauph2Ael",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			if err := ConfirmEmail(s, tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("ConfirmEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfirmEmailAppliesPendingEmail(t *testing.T) {
	t.Run("pending email becomes email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		config.MailerEnabled.Set(true)
		defer config.MailerEnabled.Set(false)

		resetEmailConfirmationCooldown(t, 1)
		mail.ResetSent()
		require.NoError(t, UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "new1@example.com"}))
		tokens, err := getTokensForKind(s, &User{ID: 1}, TokenEmailConfirm)
		require.NoError(t, err)
		require.Len(t, tokens, 1)

		require.NoError(t, ConfirmEmail(s, &EmailConfirm{Token: confirmTokenFromSentMail(t)}))
		require.NoError(t, s.Commit())

		s2 := db.NewSession()
		defer s2.Close()
		updated, err := GetUserWithEmail(s2, &User{ID: 1})
		require.NoError(t, err)
		assert.Equal(t, "new1@example.com", updated.Email)
		assert.Empty(t, updated.PendingEmail)
		assert.Equal(t, StatusActive, updated.Status)
	})

	t.Run("pending email taken meanwhile", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		config.MailerEnabled.Set(true)
		defer config.MailerEnabled.Set(false)

		resetEmailConfirmationCooldown(t, 1)
		mail.ResetSent()
		require.NoError(t, UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "new1@example.com"}))
		confirmToken := confirmTokenFromSentMail(t)

		_, err := s.Where("id = ?", 2).Cols("email").Update(&User{Email: "new1@example.com"})
		require.NoError(t, err)

		err = ConfirmEmail(s, &EmailConfirm{Token: confirmToken})
		require.Error(t, err)
		assert.True(t, IsErrUserEmailExists(err))

		// The token survives so the user can retry once the conflict is gone.
		tokens, err := getTokensForKind(s, &User{ID: 1}, TokenEmailConfirm)
		require.NoError(t, err)
		assert.Len(t, tokens, 1)
	})

	t.Run("expired token", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		config.MailerEnabled.Set(true)
		defer config.MailerEnabled.Set(false)

		resetEmailConfirmationCooldown(t, 1)
		mail.ResetSent()
		require.NoError(t, UpdateEmail(s, &EmailUpdate{User: &User{ID: 1}, NewEmail: "new1@example.com"}))
		confirmToken := confirmTokenFromSentMail(t)

		tokens, err := getTokensForKind(s, &User{ID: 1}, TokenEmailConfirm)
		require.NoError(t, err)
		require.Len(t, tokens, 1)
		backdateToken(t, s, tokens[0].ID, time.Now().Add(-25*time.Hour))

		err = ConfirmEmail(s, &EmailConfirm{Token: confirmToken})
		require.Error(t, err)
		assert.True(t, IsErrInvalidEmailConfirmToken(err))

		updated, err := GetUserWithEmail(s, &User{ID: 1})
		require.NoError(t, err)
		assert.Equal(t, "user1@example.com", updated.Email)
		assert.Equal(t, "new1@example.com", updated.PendingEmail)

		tokens, err = getTokensForKind(s, &User{ID: 1}, TokenEmailConfirm)
		require.NoError(t, err)
		assert.Empty(t, tokens)
	})

	t.Run("locked user keeps their email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := s.Where("id = ?", 18).Cols("pending_email").Update(&User{PendingEmail: "new18@example.com"})
		require.NoError(t, err)

		before, err := GetUserWithEmail(s, &User{ID: 18})
		require.Error(t, err)
		require.True(t, IsErrAccountLocked(err))

		token, err := generateToken(s, &User{ID: 18}, TokenEmailConfirm)
		require.NoError(t, err)

		err = ConfirmEmail(s, &EmailConfirm{Token: token.ClearTextToken})
		require.Error(t, err)
		assert.True(t, IsErrAccountLocked(err))
		require.NoError(t, s.Commit())

		s2 := db.NewSession()
		defer s2.Close()
		updated, err := GetUserWithEmail(s2, &User{ID: 18})
		require.Error(t, err)
		assert.True(t, IsErrAccountLocked(err))
		assert.Equal(t, StatusAccountLocked, updated.Status)
		assert.Equal(t, before.Email, updated.Email)
		assert.Equal(t, "new18@example.com", updated.PendingEmail)

		tokens, err := getTokensForKind(s2, &User{ID: 18}, TokenEmailConfirm)
		require.NoError(t, err)
		assert.Len(t, tokens, 1)
	})
}
