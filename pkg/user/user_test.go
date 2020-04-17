// Copyright 2018-2020 Vikunja and contriubtors. All rights reserved.
//
// This file is part of Vikunja.
//
// Vikunja is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Vikunja is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Vikunja.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/assert"
	"testing"
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
		createdUser, err := CreateUser(dummyuser)
		assert.NoError(t, err)
		assert.NotZero(t, createdUser.Created)
	})
	t.Run("already existing", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := CreateUser(&User{
			Username: "user1",
			Password: "12345",
			Email:    "email@example.com",
		})
		assert.Error(t, err)
		assert.True(t, IsErrUsernameExists(err))
	})
	t.Run("same email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := CreateUser(&User{
			Username: "testuser",
			Password: "12345",
			Email:    "user1@example.com",
		})
		assert.Error(t, err)
		assert.True(t, IsErrUserEmailExists(err))
	})
	t.Run("no username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := CreateUser(&User{
			Username: "",
			Password: "12345",
			Email:    "user1@example.com",
		})
		assert.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
	t.Run("no password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := CreateUser(&User{
			Username: "testuser",
			Password: "",
			Email:    "user1@example.com",
		})
		assert.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
	t.Run("no email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := CreateUser(&User{
			Username: "testuser",
			Password: "12345",
			Email:    "",
		})
		assert.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
}

func TestGetUser(t *testing.T) {
	t.Run("by name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		theuser, err := GetUser(&User{
			Username: "user1",
		})
		assert.NoError(t, err)
		assert.Equal(t, theuser.ID, int64(1))
		assert.Empty(t, theuser.Email)
	})
	t.Run("by email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		theuser, err := GetUser(&User{
			Email: "user1@example.com",
		})
		assert.NoError(t, err)
		assert.Equal(t, theuser.ID, int64(1))
		assert.Empty(t, theuser.Email)
	})
	t.Run("by id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		theuser, err := GetUserByID(1)
		assert.NoError(t, err)
		assert.Equal(t, theuser.ID, int64(1))
		assert.Equal(t, theuser.Username, "user1")
		assert.Empty(t, theuser.Email)
	})
	t.Run("invalid id", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := GetUserByID(99999)
		assert.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
	t.Run("nonexistant", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := GetUserByID(0)
		assert.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
	t.Run("empty name", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := GetUserByUsername("")
		assert.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
	t.Run("with email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		theuser, err := GetUserWithEmail(&User{ID: 1})
		assert.NoError(t, err)
		assert.Equal(t, theuser.ID, int64(1))
		assert.Equal(t, theuser.Username, "user1")
		assert.NotEmpty(t, theuser.Email)
	})
}

func TestCheckUserCredentials(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := CheckUserCredentials(&Login{Username: "user1", Password: "1234"})
		assert.NoError(t, err)
	})
	t.Run("unverified email", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := CheckUserCredentials(&Login{Username: "user5", Password: "1234"})
		assert.Error(t, err)
		assert.True(t, IsErrEmailNotConfirmed(err))
	})
	t.Run("wrong password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := CheckUserCredentials(&Login{Username: "user1", Password: "12345"})
		assert.Error(t, err)
		assert.True(t, IsErrWrongUsernameOrPassword(err))
	})
	t.Run("nonexistant user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := CheckUserCredentials(&Login{Username: "dfstestuu", Password: "1234"})
		assert.Error(t, err)
		assert.True(t, IsErrWrongUsernameOrPassword(err))
	})
	t.Run("empty password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := CheckUserCredentials(&Login{Username: "user1"})
		assert.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
	t.Run("empty username", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := CheckUserCredentials(&Login{Password: "1234"})
		assert.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
}

func TestUpdateUser(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		uuser, err := UpdateUser(&User{
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
		uuser, err := UpdateUser(&User{
			ID:       1,
			Username: "changedname",
		})
		assert.NoError(t, err)
		assert.Equal(t, "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.", uuser.Password) // Password should not change
		assert.Equal(t, "changedname", uuser.Username)
	})
	t.Run("nonexistant", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		_, err := UpdateUser(&User{
			ID: 99999,
		})
		assert.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
}

func TestUpdateUserPassword(t *testing.T) {

	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		err := UpdateUserPassword(&User{
			ID: 1,
		}, "12345",
		)
		assert.NoError(t, err)
	})
	t.Run("nonexistant user", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		err := UpdateUserPassword(&User{
			ID: 9999,
		}, "12345")
		assert.Error(t, err)
		assert.True(t, IsErrUserDoesNotExist(err))
	})
	t.Run("empty password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		err := UpdateUserPassword(&User{
			ID: 1,
		}, "",
		)
		assert.Error(t, err)
		assert.True(t, IsErrEmptyNewPassword(err))
	})
}

func TestListUsers(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		all, err := ListUsers("user1")
		assert.NoError(t, err)
		assert.True(t, len(all) > 0)
		assert.Equal(t, all[0].Username, "user1")
	})
	t.Run("all users", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		all, err := ListUsers("")
		assert.NoError(t, err)
		assert.Len(t, all, 13)
	})
}

func TestUserPasswordReset(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		reset := &PasswordReset{
			Token:       "passwordresettesttoken",
			NewPassword: "12345",
		}
		err := ResetPassword(reset)
		assert.NoError(t, err)
	})
	t.Run("without password", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		reset := &PasswordReset{
			Token: "passwordresettesttoken",
		}
		err := ResetPassword(reset)
		assert.Error(t, err)
		assert.True(t, IsErrNoUsernamePassword(err))
	})
	t.Run("empty token", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		reset := &PasswordReset{
			Token:       "somethingsomething",
			NewPassword: "12345",
		}
		err := ResetPassword(reset)
		assert.Error(t, err)
		assert.True(t, IsErrInvalidPasswordResetToken(err))
	})
	t.Run("wrong token", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		reset := &PasswordReset{
			Token:       "somethingsomething",
			NewPassword: "12345",
		}
		err := ResetPassword(reset)
		assert.Error(t, err)
		assert.True(t, IsErrInvalidPasswordResetToken(err))
	})
}
