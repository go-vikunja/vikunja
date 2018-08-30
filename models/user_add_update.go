package models

import (
	"golang.org/x/crypto/bcrypt"
)

// CreateUser creates a new user and inserts it into the database
func CreateUser(user User) (newUser User, err error) {

	newUser = user

	// Check if we have all needed informations
	if newUser.Password == "" || newUser.Username == "" {
		return User{}, ErrNoUsernamePassword{}
	}

	// Check if the user already existst with that username
	var exists bool
	existingUser, err := GetUser(User{Username: newUser.Username})
	if err != nil {
		if IsErrUserDoesNotExist(err) {
			exists = true
		} else {
			return User{}, err
		}
	}
	if exists {
		return User{}, ErrUsernameExists{existingUser.ID, existingUser.Username}
	}

	// Check if the user already existst with that email
	existingUser, err = GetUser(User{Email: newUser.Email})
	if err != nil && !IsErrUserDoesNotExist(err) {
		if IsErrUserDoesNotExist(err) {
			exists = true
		} else {
			return User{}, err
		}
	}
	if exists {
		return User{}, ErrUserEmailExists{existingUser.ID, existingUser.Email}
	}

	// Hash the password
	newUser.Password, err = hashPassword(user.Password)
	if err != nil {
		return User{}, err
	}

	// Insert it
	_, err = x.Insert(newUser)
	if err != nil {
		return User{}, err
	}

	// Get the  full new User
	newUserOut, err := GetUser(newUser)
	if err != nil {
		return User{}, err
	}

	// Create the user's namespace
	newN := &Namespace{Name: newUserOut.Username, Description: newUserOut.Username + "'s namespace.", Owner: newUserOut}
	err = newN.Create(&newUserOut)
	if err != nil {
		return User{}, err
	}

	return newUserOut, err
}

// HashPassword hashes a password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// UpdateUser updates a user
func UpdateUser(user User) (updatedUser User, err error) {

	// Check if it exists
	theUser, err := GetUserByID(user.ID)
	if err != nil {
		return User{}, err
	}

	// Check if we have at least a username
	if user.Username == "" {
		//return User{}, ErrNoUsername{user.ID}
		user.Username = theUser.Username // Dont change the username if we dont have one
	}

	user.Password = theUser.Password // set the password to the one in the database to not accedently resetting it

	// Update it
	_, err = x.Id(user.ID).Update(user)
	if err != nil {
		return User{}, err
	}

	// Get the newly updated user
	updatedUser, err = GetUserByID(user.ID)
	if err != nil {
		return User{}, err
	}

	return updatedUser, err
}

// UpdateUserPassword updates the password of a user
func UpdateUserPassword(userID int64, newPassword string, doer *User) (err error) {

	// Get all user details
	user, err := GetUserByID(userID)
	if err != nil {
		return err
	}

	// Hash the new password and set it
	hashed, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	user.Password = hashed

	// Update it
	_, err = x.Id(user.ID).Update(user)
	if err != nil {
		return err
	}

	return err
}
