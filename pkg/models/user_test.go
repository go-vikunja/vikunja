package models

import (
	"code.vikunja.io/api/pkg/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateUser(t *testing.T) {
	// Create test database
	//assert.NoError(t, PrepareTestDatabase())

	// Get our doer
	doer, err := GetUserByID(1)
	assert.NoError(t, err)

	// Our dummy user for testing
	dummyuser := User{
		Username: "testuu",
		Password: "1234",
		Email:    "noone@example.com",
	}

	// Create a new user
	createdUser, err := CreateUser(dummyuser)
	assert.NoError(t, err)

	// Create a second new user
	_, err = CreateUser(User{Username: dummyuser.Username + "2", Email: dummyuser.Email + "m", Password: dummyuser.Password})
	assert.NoError(t, err)

	// Check if it fails to create the same user again
	_, err = CreateUser(dummyuser)
	assert.Error(t, err)

	// Check if it fails to create a user with just the same username
	_, err = CreateUser(User{Username: dummyuser.Username, Password: "fsdf"})
	assert.Error(t, err)
	assert.True(t, IsErrUsernameExists(err))

	// Check if it fails to create one with the same email
	_, err = CreateUser(User{Username: "noone", Password: "1234", Email: dummyuser.Email})
	assert.Error(t, err)
	assert.True(t, IsErrUserEmailExists(err))

	// Check if it fails to create a user without password and username
	_, err = CreateUser(User{})
	assert.Error(t, err)
	assert.True(t, IsErrNoUsernamePassword(err))

	// Check if he exists
	theuser, err := GetUser(createdUser)
	assert.NoError(t, err)

	// Get by his ID
	_, err = GetUserByID(theuser.ID)
	assert.NoError(t, err)

	// Passing 0 as ID should return an error
	_, err = GetUserByID(0)
	assert.Error(t, err)
	assert.True(t, IsErrUserDoesNotExist(err))

	// Check the user credentials
	user, err := CheckUserCredentials(&UserLogin{"testuu", "1234"})
	assert.NoError(t, err)
	assert.Equal(t, "testuu", user.Username)

	// Check wrong password (should also fail)
	_, err = CheckUserCredentials(&UserLogin{"testuu", "12345"})
	assert.Error(t, err)

	// Check usercredentials for a nonexistent user (should fail)
	_, err = CheckUserCredentials(&UserLogin{"dfstestuu", "1234"})
	assert.Error(t, err)
	assert.True(t, IsErrUserDoesNotExist(err))

	// Update the user
	uuser, err := UpdateUser(User{ID: theuser.ID, Password: "444444"})
	assert.NoError(t, err)
	assert.Equal(t, theuser.Password, uuser.Password) // Password should not change
	assert.Equal(t, theuser.Username, uuser.Username) // Username should not change either

	// Try updating one which does not exist
	_, err = UpdateUser(User{ID: 99999, Username: "dg"})
	assert.Error(t, err)
	assert.True(t, IsErrUserDoesNotExist(err))

	// Update a users password
	newpassword := "55555"
	err = UpdateUserPassword(&theuser, newpassword)
	assert.NoError(t, err)

	// Check if it was changed
	user, err = CheckUserCredentials(&UserLogin{theuser.Username, newpassword})
	assert.NoError(t, err)

	// Check if the searchterm works
	all, err := ListUsers("test")
	assert.NoError(t, err)
	assert.True(t, len(all) > 0)

	all, err = ListUsers("")
	assert.NoError(t, err)
	assert.True(t, len(all) > 0)

	// Try updating the password of a nonexistent user (should fail)
	err = UpdateUserPassword(&User{ID: 9999}, newpassword)
	assert.Error(t, err)
	assert.True(t, IsErrUserDoesNotExist(err))

	// Delete it
	err = DeleteUserByID(theuser.ID, &doer)
	assert.NoError(t, err)

	// Try deleting one with ID = 0
	err = DeleteUserByID(0, &doer)
	assert.Error(t, err)
	assert.True(t, IsErrIDCannotBeZero(err))
}

func TestUserPasswordReset(t *testing.T) {
	// Request a new token
	tr := &PasswordTokenRequest{
		Username: "user1",
	}
	err := RequestUserPasswordResetToken(tr)
	assert.NoError(t, err)

	// Get the token / inside the user object
	userWithToken, err := GetUserByID(1)
	assert.NoError(t, err)

	// Try resetting it
	reset := &PasswordReset{
		Token: userWithToken.PasswordResetToken,
	}

	// Try resetting it without a password
	reset.NewPassword = ""
	err = UserPasswordReset(reset)
	assert.True(t, IsErrNoUsernamePassword(err))

	// Reset it
	reset.NewPassword = "1234"
	err = UserPasswordReset(reset)
	assert.NoError(t, err)

	// Try resetting it with a wrong token
	reset.Token = utils.MakeRandomString(400)
	err = UserPasswordReset(reset)
	assert.Error(t, err)
	assert.True(t, IsErrInvalidPasswordResetToken(err))
}
