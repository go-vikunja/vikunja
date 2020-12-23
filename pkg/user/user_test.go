// Copyright 2018-2020 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	// Our dummy user for testing
	dummyuser := &User{
		Username: "testuser",
		Password: "1234",
		Email:    "noone@example.com",
	}

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		createdUser, err := CreateUser(s, dummyuser)
		assert.NoError(t, err)
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
		assert.Error(t, err)
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
		assert.Error(t, err)
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
		assert.Error(t, err)
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
		assert.Error(t, err)
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
		assert.Error(t, err)
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
		assert.NoError(t, err)
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
		assert.NoError(t, err)
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
		assert.NoError(t, err)
		assert.Equal(t, theuser.ID, int64(1))
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
		assert.NoError(t, err)
		assert.Equal(t, theuser.ID, int64(1))
		assert.Empty(t, theuser.Email)
	})
	t.Run("by id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		theuser, err := GetUserByID(s, 1)
		assert.NoError(t, err)
		assert.Equal(t, theuser.ID, int64(1))
		assert.Equal(t, theuser.Username, "user1")
		assert.Empty(t, theuser.Email)
	})
	t.Run("invalid id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := GetUserByID(s, 99999)
		assert.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
	t.Run("nonexistant", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := GetUserByID(s, 0)
		assert.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
	t.Run("empty name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := GetUserByUsername(s, "")
		assert.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
	t.Run("with email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		theuser, err := GetUserWithEmail(s, &User{ID: 1})
		assert.NoError(t, err)
		assert.Equal(t, theuser.ID, int64(1))
		assert.Equal(t, theuser.Username, "user1")
		assert.NotEmpty(t, theuser.Email)
	})
}

func TestCheckUserCredentials(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CheckUserCredentials(s, &Login{Username: "user1", Password: "1234"})
		assert.NoError(t, err)
	})
	t.Run("unverified email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CheckUserCredentials(s, &Login{Username: "user5", Password: "1234"})
		assert.Error(t, err)
		assert.True(t, IsErrEmailNotConfirmed(err))
	})
	t.Run("wrong password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CheckUserCredentials(s, &Login{Username: "user1", Password: "12345"})
		assert.Error(t, err)
		assert.True(t, IsErrWrongUsernameOrPassword(err))
	})
	t.Run("nonexistant user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CheckUserCredentials(s, &Login{Username: "dfstestuu", Password: "1234"})
		assert.Error(t, err)
		assert.True(t, IsErrWrongUsernameOrPassword(err))
	})
	t.Run("empty password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CheckUserCredentials(s, &Login{Username: "user1"})
		assert.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
	t.Run("empty username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CheckUserCredentials(s, &Login{Password: "1234"})
		assert.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
	t.Run("email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := CheckUserCredentials(s, &Login{Username: "user1@example.com", Password: "1234"})
		assert.NoError(t, err)
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
		})
		assert.NoError(t, err)
		assert.Equal(t, "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.", uuser.Password) // Password should not change
		assert.Equal(t, "user1", uuser.Username)                                                        // Username should not change either
	})
	t.Run("change username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		uuser, err := UpdateUser(s, &User{
			ID:       1,
			Username: "changedname",
		})
		assert.NoError(t, err)
		assert.Equal(t, "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.", uuser.Password) // Password should not change
		assert.Equal(t, "changedname", uuser.Username)
	})
	t.Run("nonexistant", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		_, err := UpdateUser(s, &User{
			ID: 99999,
		})
		assert.Error(t, err)
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
		assert.NoError(t, err)
	})
	t.Run("nonexistant user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := UpdateUserPassword(s, &User{
			ID: 9999,
		}, "12345")
		assert.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
	t.Run("empty password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		err := UpdateUserPassword(s, &User{
			ID: 1,
		}, "")
		assert.Error(t, err)
		assert.True(t, IsErrEmptyNewPassword(err))
	})
}

func TestListUsers(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := ListUsers(s, "user1")
		assert.NoError(t, err)
		assert.True(t, len(all) > 0)
		assert.Equal(t, all[0].Username, "user1")
	})
	t.Run("all users", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		all, err := ListUsers(s, "")
		assert.NoError(t, err)
		assert.Len(t, all, 14)
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
		assert.NoError(t, err)
	})
	t.Run("without password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		reset := &PasswordReset{
			Token: "passwordresettesttoken",
		}
		err := ResetPassword(s, reset)
		assert.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
	t.Run("empty token", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		reset := &PasswordReset{
			Token:       "somethingsomething",
			NewPassword: "12345",
		}
		err := ResetPassword(s, reset)
		assert.Error(t, err)
		assert.True(t, IsErrInvalidPasswordResetToken(err))
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
		assert.Error(t, err)
		assert.True(t, IsErrInvalidPasswordResetToken(err))
	})
}
