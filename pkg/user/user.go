// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"code.vikunja.io/api/pkg/config"

	"code.vikunja.io/api/pkg/db"

	"xorm.io/xorm"

	"code.vikunja.io/web"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Login Object to recive user credentials in JSON format
type Login struct {
	// The username used to log in.
	Username string `json:"username"`
	// The password for the user.
	Password string `json:"password"`
	// The totp passcode of a user. Only needs to be provided when enabled.
	TOTPPasscode string `json:"totp_passcode"`
}

// User holds information about an user
type User struct {
	// The unique, numeric id of this user.
	ID int64 `xorm:"bigint autoincr not null unique pk" json:"id"`
	// The full name of the user.
	Name string `xorm:"text null" json:"name"`
	// The username of the user. Is always unique.
	Username string `xorm:"varchar(250) not null unique" json:"username" valid:"length(1|250)" minLength:"1" maxLength:"250"`
	Password string `xorm:"varchar(250) null" json:"-"`
	// The user's email address.
	Email    string `xorm:"varchar(250) null" json:"email,omitempty" valid:"email,length(0|250)" maxLength:"250"`
	IsActive bool   `xorm:"null" json:"-"`

	PasswordResetToken string `xorm:"varchar(450) null" json:"-"`
	EmailConfirmToken  string `xorm:"varchar(450) null" json:"-"`

	AvatarProvider string `xorm:"varchar(255) null" json:"-"`
	AvatarFileID   int64  `xorm:"null" json:"-"`

	// Issuer and Subject contain the issuer and subject from the source the user authenticated with.
	Issuer  string `xorm:"text null" json:"-"`
	Subject string `xorm:"text null" json:"-"`

	EmailRemindersEnabled        bool  `xorm:"bool default true" json:"-"`
	DiscoverableByName           bool  `xorm:"bool default false index" json:"-"`
	DiscoverableByEmail          bool  `xorm:"bool default false index" json:"-"`
	OverdueTasksRemindersEnabled bool  `xorm:"bool default true index" json:"-"`
	DefaultListID                int64 `xorm:"bigint null index" json:"-"`
	WeekStart                    int   `xorm:"null" json:"-"`

	// A timestamp when this task was created. You cannot change this value.
	Created time.Time `xorm:"created not null" json:"created"`
	// A timestamp when this task was last updated. You cannot change this value.
	Updated time.Time `xorm:"updated not null" json:"updated"`

	web.Auth `xorm:"-" json:"-"`
}

// RouteForMail routes all notifications for a user to its email address
func (u *User) RouteForMail() (string, error) {

	if u.Email == "" {
		s := db.NewSession()
		defer s.Close()
		user, err := getUser(s, &User{ID: u.ID}, true)
		if err != nil {
			return "", err
		}
		return user.Email, nil
	}

	return u.Email, nil
}

// RouteForDB routes all notifications for a user to their id
func (u *User) RouteForDB() int64 {
	return u.ID
}

// GetID implements the Auth interface
func (u *User) GetID() int64 {
	return u.ID
}

// TableName returns the table name for users
func (User) TableName() string {
	return "users"
}

// GetName returns the name if the user has one and the username otherwise.
func (u *User) GetName() string {
	if u.Name != "" {
		return u.Name
	}

	return u.Username
}

// GetNameAndFromEmail returns the name and email address for a user. Useful to use in notifications.
func (u *User) GetNameAndFromEmail() string {
	return u.GetName() + " via Vikunja <" + config.MailerFromEmail.GetString() + ">"
}

// GetFromAuth returns a user object from a web.Auth object and returns an error if the underlying type
// is not a user object
func GetFromAuth(a web.Auth) (*User, error) {
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
func GetUserByID(s *xorm.Session, id int64) (user *User, err error) {
	// Apparently xorm does otherwise look for all users but return only one, which leads to returing one even if the ID is 0
	if id < 1 {
		return &User{}, ErrUserDoesNotExist{}
	}

	return getUser(s, &User{ID: id}, false)
}

// GetUserByUsername gets a user from its user name. This is an extra function to be able to add an extra error check.
func GetUserByUsername(s *xorm.Session, username string) (user *User, err error) {
	if username == "" {
		return &User{}, ErrUserDoesNotExist{}
	}

	return getUser(s, &User{Username: username}, false)
}

// GetUserWithEmail returns a user object with email
func GetUserWithEmail(s *xorm.Session, user *User) (userOut *User, err error) {
	return getUser(s, user, true)
}

// GetUsersByIDs returns a map of users from a slice of user ids
func GetUsersByIDs(s *xorm.Session, userIDs []int64) (users map[int64]*User, err error) {
	users = make(map[int64]*User)

	if len(userIDs) == 0 {
		return users, nil
	}

	err = s.In("id", userIDs).Find(&users)
	if err != nil {
		return
	}

	// Obfuscate all user emails
	for _, u := range users {
		u.Email = ""
	}

	return
}

// getUser is a small helper function to avoid having duplicated code for almost the same use case
func getUser(s *xorm.Session, user *User, withEmail bool) (userOut *User, err error) {
	userOut = &User{} // To prevent a panic if user is nil
	*userOut = *user
	exists, err := s.Get(userOut)
	if err != nil {
		return nil, err
	}
	if !exists {
		return &User{}, ErrUserDoesNotExist{UserID: user.ID}
	}

	if !withEmail {
		userOut.Email = ""
	}

	return userOut, err
}

func getUserByUsernameOrEmail(s *xorm.Session, usernameOrEmail string) (u *User, err error) {
	u = &User{}
	exists, err := s.
		Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).
		Get(u)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserDoesNotExist{}
	}

	u.Email = ""
	return
}

// CheckUserCredentials checks user credentials
func CheckUserCredentials(s *xorm.Session, u *Login) (*User, error) {
	// Check if we have any credentials
	if u.Password == "" || u.Username == "" {
		return nil, ErrNoUsernamePassword{}
	}

	// Check if the user exists
	user, err := getUserByUsernameOrEmail(s, u.Username)
	if err != nil {
		// hashing the password takes a long time, so we hash something to not make it clear if the username was wrong
		_, _ = bcrypt.GenerateFromPassword([]byte(u.Username), 14)
		return nil, ErrWrongUsernameOrPassword{}
	}

	// The user is invalid if they need to verify their email address
	if !user.IsActive {
		return &User{}, ErrEmailNotConfirmed{UserID: user.ID}
	}

	// Check the users password
	err = CheckUserPassword(user, u.Password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CheckUserPassword checks and verifies a user's password. The user object needs to contain the hashed password from the database.
func CheckUserPassword(user *User, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrWrongUsernameOrPassword{}
		}
		return err
	}

	return nil
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
		Name:     claims["name"].(string),
	}

	return
}

// UpdateUser updates a user
func UpdateUser(s *xorm.Session, user *User) (updatedUser *User, err error) {

	// Check if it exists
	theUser, err := GetUserWithEmail(s, &User{ID: user.ID})
	if err != nil {
		return &User{}, err
	}

	// Check if we have at least a username
	if user.Username == "" {
		user.Username = theUser.Username // Dont change the username if we dont have one
	} else {
		// Check if the new username already exists
		uu, err := GetUserByUsername(s, user.Username)
		if err != nil && !IsErrUserDoesNotExist(err) {
			return nil, err
		}
		if uu.ID != 0 && uu.ID != user.ID {
			return nil, &ErrUsernameExists{Username: user.Username, UserID: uu.ID}
		}
	}

	// Check if we have a name
	if user.Name == "" {
		user.Name = theUser.Name
	}

	// Check if the email is already used
	if user.Email == "" {
		user.Email = theUser.Email
	} else {
		uu, err := getUser(s, &User{
			Email:   user.Email,
			Issuer:  user.Issuer,
			Subject: user.Subject,
		}, true)
		if err != nil && !IsErrUserDoesNotExist(err) {
			return nil, err
		}
		if uu.ID != 0 && uu.ID != user.ID {
			return nil, &ErrUserEmailExists{Email: user.Email, UserID: uu.ID}
		}
	}

	// Validate the avatar type
	if user.AvatarProvider != "" {
		if user.AvatarProvider != "default" &&
			user.AvatarProvider != "gravatar" &&
			user.AvatarProvider != "initials" &&
			user.AvatarProvider != "upload" {
			return updatedUser, &ErrInvalidAvatarProvider{AvatarProvider: user.AvatarProvider}
		}
	}

	// Update it
	_, err = s.
		ID(user.ID).
		Cols(
			"username",
			"email",
			"avatar_provider",
			"avatar_file_id",
			"is_active",
			"name",
			"email_reminders_enabled",
			"discoverable_by_name",
			"discoverable_by_email",
			"overdue_tasks_reminders_enabled",
			"default_list_id",
			"week_start",
		).
		Update(user)
	if err != nil {
		return &User{}, err
	}

	// Get the newly updated user
	updatedUser, err = GetUserByID(s, user.ID)
	if err != nil {
		return &User{}, err
	}

	return updatedUser, err
}

// UpdateUserPassword updates the password of a user
func UpdateUserPassword(s *xorm.Session, user *User, newPassword string) (err error) {

	if newPassword == "" {
		return ErrEmptyNewPassword{}
	}

	// Get all user details
	theUser, err := GetUserByID(s, user.ID)
	if err != nil {
		return err
	}

	// Hash the new password and set it
	hashed, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	theUser.Password = hashed

	// Update it
	_, err = s.ID(user.ID).Update(theUser)
	if err != nil {
		return err
	}

	return err
}
