package models

import (
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

	// Delete every preexisting user to have a fresh start
	_, err = x.Where("1 = 1").Delete(&User{})
	assert.NoError(t, err)

	allusers, err := ListUsers("")
	assert.NoError(t, err)
	for _, user := range allusers {
		// Delete it
		err := DeleteUserByID(user.ID, &doer)
		assert.NoError(t, err)
	}

	// Create a new user
	createdUser, err := CreateUser(dummyuser)
	assert.NoError(t, err)

	// Create a second new user
	createdUser2, err := CreateUser(User{Username: dummyuser.Username + "2", Email: dummyuser.Email + "m", Password: dummyuser.Password})
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

	// Passing 0 as ID should return an empty user
	_, err = GetUserByID(0)
	assert.NoError(t, err)

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
	err = UpdateUserPassword(theuser.ID, newpassword, &doer)
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
	err = UpdateUserPassword(9999, newpassword, &doer)
	assert.Error(t, err)
	assert.True(t, IsErrUserDoesNotExist(err))

	// Delete it
	err = DeleteUserByID(theuser.ID, &doer)
	assert.NoError(t, err)

	// Try deleting one with ID = 0
	err = DeleteUserByID(0, &doer)
	assert.Error(t, err)
	assert.True(t, IsErrIDCannotBeZero(err))

	// Try delete the last user (Should fail)
	err = DeleteUserByID(createdUser2.ID, &doer)
	assert.Error(t, err)
	assert.True(t, IsErrCannotDeleteLastUser(err))
}
