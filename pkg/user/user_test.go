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

	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateUser_RejectsBotPrefix(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	_, err := CreateUser(s, &User{
		Username: "bot-evil",
		Password: "12345678",
		Email:    "x@example.com",
	})
	require.Error(t, err)
	assert.True(t, IsErrUsernameReserved(err))
}

func TestUser_IsBot(t *testing.T) {
	t.Run("regular user", func(t *testing.T) {
		u := &User{ID: 1}
		assert.False(t, u.IsBot())
	})
	t.Run("bot user", func(t *testing.T) {
		u := &User{ID: 2, BotOwnerID: 1}
		assert.True(t, u.IsBot())
	})
}

func TestCreateUser(t *testing.T) {
	// Our dummy user for testing
	dummyuser := &User{
		Username: "testuser",
		Password: "12345678",
		Email:    "noone@example.com",
	}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		createdUser, err := CreateUser(s, dummyuser)
		require.NoError(t, err)
		assert.NotZero(t, createdUser.Created)
	})
	t.Run("already existing", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CreateUser(s, &User{
			Username: "user1",
			Password: "12345",
			Email:    "email@example.com",
		})
		require.Error(t, err)
		assert.True(t, IsErrUsernameExists(err))
	})
	t.Run("same email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CreateUser(s, &User{
			Username: "testuser",
			Password: "12345",
			Email:    "user1@example.com",
		})
		require.Error(t, err)
		assert.True(t, IsErrUserEmailExists(err))
	})
	t.Run("no username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CreateUser(s, &User{
			Username: "",
			Password: "12345",
			Email:    "user1@example.com",
		})
		require.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
	t.Run("no password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CreateUser(s, &User{
			Username: "testuser",
			Password: "",
			Email:    "user1@example.com",
		})
		require.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
	t.Run("no email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CreateUser(s, &User{
			Username: "testuser",
			Password: "12345",
			Email:    "",
		})
		require.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
	t.Run("same email but different issuer", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CreateUser(s, &User{
			Username: "somenewuser",
			Email:    "user1@example.com",
			Issuer:   "https://some.site",
			Subject:  "12345",
		})
		require.NoError(t, err)
	})
	t.Run("same subject but different issuer", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CreateUser(s, &User{
			Username: "somenewuser",
			Email:    "somenewuser@example.com",
			Issuer:   "https://some.site",
			Subject:  "12345",
		})
		require.NoError(t, err)
	})
	t.Run("space in username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CreateUser(s, &User{
			Username: "user name",
			Password: "12345",
			Email:    "user1@example.com",
		})
		require.Error(t, err)
		assert.True(t, IsErrUsernameMustNotContainSpaces(err))
	})
	t.Run("reserved link-share username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CreateUser(s, &User{
			Username: "link-share-123",
			Password: "12345678",
			Email:    "user2@example.com",
		})
		require.Error(t, err)
		assert.True(t, IsErrUsernameReserved(err))
	})
	t.Run("reserved link-share username with single digit", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CreateUser(s, &User{
			Username: "link-share-1",
			Password: "12345678",
			Email:    "user3@example.com",
		})
		require.Error(t, err)
		assert.True(t, IsErrUsernameReserved(err))
	})
}

func TestGetUser(t *testing.T) {
	t.Run("by name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		theuser, err := getUser(
			s,
			&User{
				Username: "user1",
			},
			false,
		)
		require.NoError(t, err)
		assert.Equal(t, int64(1), theuser.ID)
		assert.Empty(t, theuser.Email)
	})
	t.Run("by email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		theuser, err := getUser(
			s,
			&User{
				Email: "user1@example.com",
			},
			false)
		require.NoError(t, err)
		assert.Equal(t, int64(1), theuser.ID)
		assert.Empty(t, theuser.Email)
	})
	t.Run("by id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		theuser, err := GetUserByID(s, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), theuser.ID)
		assert.Equal(t, "user1", theuser.Username)
		assert.Empty(t, theuser.Email)
	})
	t.Run("invalid id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := GetUserByID(s, 99999)
		require.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
	t.Run("nonexistant", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := GetUserByID(s, 0)
		require.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
	t.Run("empty name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := GetUserByUsername(s, "")
		require.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
	t.Run("with email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		theuser, err := GetUserWithEmail(s, &User{ID: 1})
		require.NoError(t, err)
		assert.Equal(t, int64(1), theuser.ID)
		assert.Equal(t, "user1", theuser.Username)
		assert.NotEmpty(t, theuser.Email)
	})
}

func TestCheckUserCredentials(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CheckUserCredentials(s, &Login{Username: "user1", Password: "12345678"})
		require.NoError(t, err)
	})
	t.Run("unverified email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CheckUserCredentials(s, &Login{Username: "user5", Password: "12345678"})
		require.Error(t, err)
		assert.True(t, IsErrEmailNotConfirmed(err))
	})
	t.Run("wrong password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CheckUserCredentials(s, &Login{Username: "user1", Password: "12345"})
		require.Error(t, err)
		assert.True(t, IsErrWrongUsernameOrPassword(err))
	})
	t.Run("nonexistant user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CheckUserCredentials(s, &Login{Username: "dfstestuu", Password: "12345678"})
		require.Error(t, err)
		assert.True(t, IsErrWrongUsernameOrPassword(err))
	})
	t.Run("empty password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CheckUserCredentials(s, &Login{Username: "user1"})
		require.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
	t.Run("empty username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CheckUserCredentials(s, &Login{Password: "12345678"})
		require.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
	t.Run("email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CheckUserCredentials(s, &Login{Username: "user1@example.com", Password: "12345678"})
		require.NoError(t, err)
	})
	t.Run("disabled user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// user17 is disabled (status=2), password is "12345678"
		_, err := CheckUserCredentials(s, &Login{Username: "user17", Password: "12345678"})
		require.Error(t, err)
		assert.True(t, IsErrAccountDisabled(err))
	})
	t.Run("locked user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// user18 is locked (status=3), password is "12345678"
		_, err := CheckUserCredentials(s, &Login{Username: "user18", Password: "12345678"})
		require.Error(t, err)
		assert.True(t, IsErrAccountLocked(err))
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		uuser, err := UpdateUser(s, &User{
			ID:       1,
			Password: "LoremIpsum",
			Email:    "testing@example.com",
		}, false)
		require.NoError(t, err)
		assert.Equal(t, "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6", uuser.Password) // Password should not change
		assert.Equal(t, "user1", uuser.Username)                                                        // Username should not change either
	})
	t.Run("change username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		uuser, err := UpdateUser(s, &User{
			ID:       1,
			Username: "changedname",
		}, false)
		require.NoError(t, err)
		assert.Equal(t, "$2a$04$X4aRMEt0ytgPwMIgv36cI..7X9.nhY/.tYwxpqSi0ykRHx2CwQ0S6", uuser.Password) // Password should not change
		assert.Equal(t, "changedname", uuser.Username)
	})
	t.Run("nonexistant", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := UpdateUser(s, &User{
			ID: 99999,
		}, false)
		require.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
}

func TestUpdateUserPassword(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := UpdateUserPassword(s, &User{
			ID: 1,
		}, "12345")
		require.NoError(t, err)
	})
	t.Run("nonexistant user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := UpdateUserPassword(s, &User{
			ID: 9999,
		}, "12345")
		require.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
	t.Run("empty password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := UpdateUserPassword(s, &User{
			ID: 1,
		}, "")
		require.Error(t, err)
		assert.True(t, IsErrEmptyNewPassword(err))
	})
}

func TestUserPasswordReset(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		reset := &PasswordReset{
			Token:       "passwordresettesttoken",
			NewPassword: "12345",
		}
		_, err := ResetPassword(s, reset)
		require.NoError(t, err)
	})
	t.Run("removes password reset token after use", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		token := "passwordresettesttoken"

		reset := &PasswordReset{
			Token:       token,
			NewPassword: "12345",
		}
		_, err := ResetPassword(s, reset)
		require.NoError(t, err)

		err = s.Commit()
		require.NoError(t, err)

		db.AssertMissing(t, "user_tokens", map[string]interface{}{
			"token": token,
			"kind":  TokenPasswordReset,
		})
	})
	t.Run("without password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		reset := &PasswordReset{
			Token: "passwordresettesttoken",
		}
		_, err := ResetPassword(s, reset)
		require.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
	t.Run("empty token", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		reset := &PasswordReset{
			Token:       "",
			NewPassword: "12345",
		}
		_, err := ResetPassword(s, reset)
		require.Error(t, err)
		assert.True(t, IsErrNoPasswordResetToken(err))
	})
	t.Run("wrong token", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		reset := &PasswordReset{
			Token:       "somethingsomething",
			NewPassword: "12345",
		}
		_, err := ResetPassword(s, reset)
		require.Error(t, err)
		assert.True(t, IsErrInvalidPasswordResetToken(err))
	})
	t.Run("disabled user cannot reset password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		reset := &PasswordReset{
			Token:       "disableduserpasswordresettoken",
			NewPassword: "12345678",
		}
		_, err := ResetPassword(s, reset)
		require.Error(t, err)
		assert.True(t, IsErrAccountDisabled(err))
	})
}

func TestRequestPasswordResetTokenDisabledUser(t *testing.T) {
	t.Run("disabled user cannot request password reset token", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := RequestUserPasswordResetTokenByEmail(s, &PasswordTokenRequest{
			Email: "user17@example.com",
		})
		require.Error(t, err)
		assert.True(t, IsErrAccountDisabled(err))
	})
}

func TestCleanupOldTokens(t *testing.T) {
	t.Run("deletes old tokens and keeps recent ones", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		// Insert a recent password reset token that should NOT be deleted
		recentToken := &Token{
			UserID: 1,
			Token:  "recenttoken",
			Kind:   TokenPasswordReset,
		}
		_, err := s.Insert(recentToken)
		require.NoError(t, err)

		deleted, err := CleanupOldTokens(s)
		require.NoError(t, err)

		// Fixtures have three old tokens that should be cleaned up:
		// id=1 (kind=1, TokenPasswordReset, created 2021), id=4 (kind=3, TokenAccountDeletion, created 2021),
		// and id=5 (kind=1, TokenPasswordReset for disabled user, created 2024)
		assert.Equal(t, int64(3), deleted)

		err = s.Commit()
		require.NoError(t, err)

		// The old password reset token from fixtures should be gone
		db.AssertMissing(t, "user_tokens", map[string]interface{}{
			"id": 1,
		})
		// The old account deletion token from fixtures should be gone
		db.AssertMissing(t, "user_tokens", map[string]interface{}{
			"id": 4,
		})
		// The recent token should still exist
		db.AssertExists(t, "user_tokens", map[string]interface{}{
			"token": "recenttoken",
			"kind":  TokenPasswordReset,
		}, false)
	})
	t.Run("does not delete email confirm tokens", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CleanupOldTokens(s)
		require.NoError(t, err)

		err = s.Commit()
		require.NoError(t, err)

		// The old email confirm tokens (kind=2) from fixtures should still exist
		db.AssertExists(t, "user_tokens", map[string]interface{}{
			"id":   2,
			"kind": TokenEmailConfirm,
		}, false)
		db.AssertExists(t, "user_tokens", map[string]interface{}{
			"id":   3,
			"kind": TokenEmailConfirm,
		}, false)
	})
}

func TestConfirmDeletion(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		user := &User{ID: 1}
		err := ConfirmDeletion(s, user, "deletiontesttoken")
		require.NoError(t, err)

		updatedUser, err := GetUserByID(s, 1)
		require.NoError(t, err)
		assert.False(t, updatedUser.DeletionScheduledAt.IsZero())
	})
	t.Run("invalid token", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		user := &User{ID: 1}
		err := ConfirmDeletion(s, user, "invalidtoken")
		require.Error(t, err)
		assert.True(t, IsErrInvalidDeletionToken(err))

		invalidErr := err.(ErrInvalidDeletionToken)
		assert.Equal(t, "invalidtoken", invalidErr.Token)
	})
	t.Run("token user mismatch", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		user := &User{ID: 3}
		err := ConfirmDeletion(s, user, "deletiontesttoken")
		require.Error(t, err)
		assert.True(t, IsErrTokenUserMismatch(err))

		mismatchErr := err.(ErrTokenUserMismatch)
		assert.Equal(t, int64(1), mismatchErr.TokenUserID)
		assert.Equal(t, int64(3), mismatchErr.UserID)
	})
	t.Run("removes token after successful confirmation", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		token := "deletiontesttoken"

		user := &User{ID: 1}
		err := ConfirmDeletion(s, user, token)
		require.NoError(t, err)

		err = s.Commit()
		require.NoError(t, err)

		db.AssertMissing(t, "user_tokens", map[string]interface{}{
			"token": token,
			"kind":  TokenAccountDeletion,
		})
	})
}

func TestGetUserByID_DisabledUser(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// user17 is disabled (status=2)
	u, err := GetUserByID(s, 17)
	require.Error(t, err)
	assert.True(t, IsErrAccountDisabled(err), "GetUserByID should return ErrAccountDisabled, got: %v", err)
	// User should still be returned alongside the error
	assert.NotNil(t, u)
	assert.Equal(t, int64(17), u.ID)
}

func TestGetUserByID_ActiveUser(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// user1 is active
	u, err := GetUserByID(s, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), u.ID)
}
