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
	"xorm.io/builder"
)

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

func TestListUsers(t *testing.T) {
	user1 := &User{ID: 1}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := ListUsers(s, "user1", user1, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, all)
		assert.Equal(t, "user1", all[0].Username)
	})
	t.Run("case insensitive", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := ListUsers(s, "uSEr1", user1, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, all)
		assert.Equal(t, "user1", all[0].Username)
	})
	t.Run("all users", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := ListAllUsers(s)
		require.NoError(t, err)
		assert.Len(t, all, 16)
	})
	t.Run("no search term", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := ListUsers(s, "", user1, nil)
		require.NoError(t, err)
		assert.Empty(t, all)
	})
	t.Run("not discoverable by email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := ListUsers(s, "user1@example.com", user1, nil)
		require.NoError(t, err)
		assert.Empty(t, all)
		db.AssertExists(t, "users", map[string]interface{}{
			"email":                 "user1@example.com",
			"discoverable_by_email": false,
		}, false)
	})
	t.Run("not discoverable by name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := ListUsers(s, "one else", user1, nil)
		require.NoError(t, err)
		assert.Empty(t, all)
		db.AssertExists(t, "users", map[string]interface{}{
			"name":                 "Some one else",
			"discoverable_by_name": false,
		}, false)
	})
	t.Run("discoverable by email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := ListUsers(s, "user7@example.com", user1, nil)
		require.NoError(t, err)
		assert.Len(t, all, 1)
		assert.Equal(t, int64(7), all[0].ID)
		db.AssertExists(t, "users", map[string]interface{}{
			"email":                 "user7@example.com",
			"discoverable_by_email": true,
		}, false)
	})
	t.Run("discoverable by partial name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := ListUsers(s, "with space", user1, nil)
		require.NoError(t, err)
		assert.Len(t, all, 1)
		assert.Equal(t, int64(12), all[0].ID)
		db.AssertExists(t, "users", map[string]interface{}{
			"name":                 "Name with spaces",
			"discoverable_by_name": true,
		}, false)
	})
	t.Run("discoverable by email with extra condition", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := ListUsers(s, "user7@example.com", user1, &ProjectUserOpts{AdditionalCond: builder.In("id", 7)})
		require.NoError(t, err)
		assert.Len(t, all, 1)
		assert.Equal(t, int64(7), all[0].ID)
		db.AssertExists(t, "users", map[string]interface{}{
			"email":                 "user7@example.com",
			"discoverable_by_email": true,
		}, false)
	})
	t.Run("discoverable by exact username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := ListUsers(s, "user7", user1, nil)
		require.NoError(t, err)
		assert.Len(t, all, 1)
		assert.Equal(t, int64(7), all[0].ID)
		db.AssertExists(t, "users", map[string]interface{}{
			"username": "user7",
		}, false)
	})
	t.Run("not discoverable by partial username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := ListUsers(s, "user", user1, nil)
		require.NoError(t, err)
		assert.Empty(t, all)
		db.AssertExists(t, "users", map[string]interface{}{
			"username": "user7",
		}, false)
	})
	t.Run("discoverable by partial username, email and name when matching fuzzily", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := ListUsers(s, "user", user1, &ProjectUserOpts{
			MatchFuzzily: true,
		})
		require.NoError(t, err)
		assert.Len(t, all, 16)
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
		err := ResetPassword(s, reset)
		require.NoError(t, err)
	})
	t.Run("without password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		reset := &PasswordReset{
			Token: "passwordresettesttoken",
		}
		err := ResetPassword(s, reset)
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
		err := ResetPassword(s, reset)
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
		err := ResetPassword(s, reset)
		require.Error(t, err)
		assert.True(t, IsErrInvalidPasswordResetToken(err))
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

		db.AssertMissing(t, "user_tokens", map[string]interface{}{
			"token": token,
			"kind":  TokenAccountDeletion,
		})
	})
}
