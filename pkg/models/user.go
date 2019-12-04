// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2019 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/mail"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/web"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"reflect"
)

// UserLogin Object to recive user credentials in JSON format
type UserLogin struct {
	// The username used to log in.
	Username string `json:"username"`
	// The password for the user.
	Password string `json:"password"`
}

// User holds information about an user
type User struct {
	// The unique, numeric id of this user.
	ID int64 `xorm:"int(11) autoincr not null unique pk" json:"id"`
	// The username of the user. Is always unique.
	Username string `xorm:"varchar(250) not null unique" json:"username" valid:"length(3|250)" minLength:"3" maxLength:"250"`
	Password string `xorm:"varchar(250) not null" json:"-"`
	// The user's email address.
	Email    string `xorm:"varchar(250) null" json:"email,omitempty" valid:"email,length(0|250)" maxLength:"250"`
	IsActive bool   `xorm:"null" json:"-"`
	// The users md5-hashed email address, used to get the avatar from gravatar and the likes.
	AvatarURL string `xorm:"-" json:"avatarUrl"`

	PasswordResetToken string `xorm:"varchar(450) null" json:"-"`
	EmailConfirmToken  string `xorm:"varchar(450) null" json:"-"`

	// A unix timestamp when this task was created. You cannot change this value.
	Created int64 `xorm:"created not null" json:"created"`
	// A unix timestamp when this task was last updated. You cannot change this value.
	Updated int64 `xorm:"updated not null" json:"updated"`

	web.Auth `xorm:"-" json:"-"`
}

// AfterLoad is used to delete all emails to not have them leaked to the user
func (u *User) AfterLoad() {
	u.AvatarURL = utils.Md5String(u.Email)
}

// GetID implements the Auth interface
func (u *User) GetID() int64 {
	return u.ID
}

// TableName returns the table name for users
func (User) TableName() string {
	return "users"
}

func getUserWithError(a web.Auth) (*User, error) {
	u, is := a.(*User)
	if !is {
		return &User{}, fmt.Errorf("user is not user element, is %s", reflect.TypeOf(a))
	}
	return u, nil
}

// APIUserPassword represents a user object without timestamps and a json password field.
type APIUserPassword struct {
	// The unique, numeric id of this user.
	ID int64 `json:"id"`
	// The username of the username. Is always unique.
	Username string `json:"username" valid:"length(3|250)" minLength:"3" maxLength:"250"`
	// The user's password in clear text. Only used when registering the user.
	Password string `json:"password" valid:"length(8|250)" minLength:"8" maxLength:"250"`
	// The user's email address
	Email string `json:"email" valid:"email,length(0|250)" maxLength:"250"`
}

// APIFormat formats an API User into a normal user struct
func (apiUser *APIUserPassword) APIFormat() *User {
	return &User{
		ID:       apiUser.ID,
		Username: apiUser.Username,
		Password: apiUser.Password,
		Email:    apiUser.Email,
	}
}

// GetUserByID gets informations about a user by its ID
func GetUserByID(id int64) (user *User, err error) {
	// Apparently xorm does otherwise look for all users but return only one, which leads to returing one even if the ID is 0
	if id < 1 {
		return &User{}, ErrUserDoesNotExist{}
	}

	return GetUser(&User{ID: id})
}

// GetUserByUsername gets a user from its user name. This is an extra function to be able to add an extra error check.
func GetUserByUsername(username string) (user *User, err error) {
	if username == "" {
		return &User{}, ErrUserDoesNotExist{}
	}

	return GetUser(&User{Username: username})
}

// GetUser gets a user object
func GetUser(user *User) (userOut *User, err error) {
	return getUser(user, false)
}

// GetUserWithEmail returns a user object with email
func GetUserWithEmail(user *User) (userOut *User, err error) {
	return getUser(user, true)
}

// getUser is a small helper function to avoid having duplicated code for almost the same use case
func getUser(user *User, withEmail bool) (userOut *User, err error) {
	userOut = &User{} // To prevent a panic if user is nil
	*userOut = *user
	exists, err := x.Get(userOut)

	if !exists {
		return &User{}, ErrUserDoesNotExist{UserID: user.ID}
	}

	if !withEmail {
		userOut.Email = ""
	}

	return userOut, err
}

// CheckUserCredentials checks user credentials
func CheckUserCredentials(u *UserLogin) (*User, error) {
	// Check if we have any credentials
	if u.Password == "" || u.Username == "" {
		return &User{}, ErrNoUsernamePassword{}
	}

	// Check if the user exists
	user, err := GetUserByUsername(u.Username)
	if err != nil {
		// hashing the password takes a long time, so we hash something to not make it clear if the username was wrong
		bcrypt.GenerateFromPassword([]byte(u.Username), 14)
		return &User{}, ErrWrongUsernameOrPassword{}
	}

	// User is invalid if it needs to verify its email address
	if !user.IsActive {
		return &User{}, ErrEmailNotConfirmed{UserID: user.ID}
	}

	// Check the users password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return &User{}, ErrWrongUsernameOrPassword{}
		}
		return &User{}, err
	}

	return user, nil
}

// GetCurrentUser returns the current user based on its jwt token
func GetCurrentUser(c echo.Context) (user *User, err error) {
	jwtinf := c.Get("user").(*jwt.Token)
	claims := jwtinf.Claims.(jwt.MapClaims)
	return GetUserFromClaims(claims)
}

// GetUserFromClaims Returns a new user from jwt claims
func GetUserFromClaims(claims jwt.MapClaims) (user *User, err error) {
	userID, ok := claims["id"].(float64)
	if !ok {
		return user, ErrCouldNotGetUserID{}
	}
	user = &User{
		ID:       int64(userID),
		Email:    claims["email"].(string),
		Username: claims["username"].(string),
	}

	return
}

// CreateUser creates a new user and inserts it into the database
func CreateUser(user *User) (newUser *User, err error) {

	newUser = user

	// Check if we have all needed informations
	if newUser.Password == "" || newUser.Username == "" || newUser.Email == "" {
		return &User{}, ErrNoUsernamePassword{}
	}

	// Check if the user already existst with that username
	exists := true
	_, err = GetUserByUsername(newUser.Username)
	if err != nil {
		if IsErrUserDoesNotExist(err) {
			exists = false
		} else {
			return &User{}, err
		}
	}
	if exists {
		return &User{}, ErrUsernameExists{newUser.ID, newUser.Username}
	}

	// Check if the user already existst with that email
	exists = true
	_, err = GetUser(&User{Email: newUser.Email})
	if err != nil {
		if IsErrUserDoesNotExist(err) {
			exists = false
		} else {
			return &User{}, err
		}
	}
	if exists {
		return &User{}, ErrUserEmailExists{newUser.ID, newUser.Email}
	}

	// Hash the password
	newUser.Password, err = hashPassword(user.Password)
	if err != nil {
		return &User{}, err
	}

	newUser.IsActive = true
	if config.MailerEnabled.GetBool() {
		// The new user should not be activated until it confirms his mail address
		newUser.IsActive = false
		// Generate a confirm token
		newUser.EmailConfirmToken = utils.MakeRandomString(400)
	}

	// Insert it
	_, err = x.Insert(newUser)
	if err != nil {
		return &User{}, err
	}

	// Update the metrics
	metrics.UpdateCount(1, metrics.ActiveUsersKey)

	// Get the  full new User
	newUserOut, err := GetUser(newUser)
	if err != nil {
		return &User{}, err
	}

	// Create the user's namespace
	newN := &Namespace{Name: newUserOut.Username, Description: newUserOut.Username + "'s namespace.", Owner: newUserOut}
	err = newN.Create(newUserOut)
	if err != nil {
		return &User{}, err
	}

	// Dont send a mail if we're testing
	if !config.MailerEnabled.GetBool() {
		return newUserOut, err
	}

	// Send the user a mail with a link to confirm the mail
	data := map[string]interface{}{
		"User": newUserOut,
	}

	mail.SendMailWithTemplate(user.Email, newUserOut.Username+" + Vikunja = <3", "confirm-email", data)

	return newUserOut, err
}

// HashPassword hashes a password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 11)
	return string(bytes), err
}

// UpdateUser updates a user
func UpdateUser(user *User) (updatedUser *User, err error) {

	// Check if it exists
	theUser, err := GetUserByID(user.ID)
	if err != nil {
		return &User{}, err
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
		return &User{}, err
	}

	// Get the newly updated user
	updatedUser, err = GetUserByID(user.ID)
	if err != nil {
		return &User{}, err
	}

	return updatedUser, err
}

// UpdateUserPassword updates the password of a user
func UpdateUserPassword(user *User, newPassword string) (err error) {

	if newPassword == "" {
		return ErrEmptyNewPassword{}
	}

	// Get all user details
	theUser, err := GetUserByID(user.ID)
	if err != nil {
		return err
	}

	// Hash the new password and set it
	hashed, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	theUser.Password = hashed

	// Update it
	_, err = x.Id(user.ID).Update(theUser)
	if err != nil {
		return err
	}

	return err
}

// DeleteUserByID deletes a user by its ID
func DeleteUserByID(id int64, doer *User) error {
	// Check if the id is 0
	if id == 0 {
		return ErrIDCannotBeZero{}
	}

	// Delete the user
	_, err := x.Id(id).Delete(&User{})

	if err != nil {
		return err
	}

	// Update the metrics
	metrics.UpdateCount(-1, metrics.ActiveUsersKey)

	return err
}
