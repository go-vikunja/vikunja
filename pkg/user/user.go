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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/web"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// Login Object to recive user credentials in JSON format
type Login struct {
	// The username used to log in.
	Username string `json:"username"`
	// The password for the user.
	Password string `json:"password"`
	// The totp passcode of a user. Only needs to be provided when enabled.
	TOTPPasscode string `json:"totp_passcode"`
	// If true, the token returned will be valid a lot longer than default. Useful for "remember me" style logins.
	LongToken bool `json:"long_token"`
}

type Status int

func (s Status) String() string {
	switch s {
	case StatusActive:
		return "Active"
	case StatusEmailConfirmationRequired:
		return "Email Confirmation required"
	case StatusDisabled:
		return "Disabled"
	}

	return "Unknown"
}

const (
	StatusActive Status = iota
	StatusEmailConfirmationRequired
	StatusDisabled
)

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
	Email string `xorm:"varchar(250) null" json:"email,omitempty" valid:"email,length(0|250)" maxLength:"250"`

	Status Status `xorm:"default 0" json:"-"`

	AvatarProvider string `xorm:"varchar(255) null" json:"-"`
	AvatarFileID   int64  `xorm:"null" json:"-"`

	// Issuer and Subject contain the issuer and subject from the source the user authenticated with.
	Issuer  string `xorm:"text null" json:"-"`
	Subject string `xorm:"text null" json:"-"`

	EmailRemindersEnabled        bool   `xorm:"bool default true" json:"-"`
	DiscoverableByName           bool   `xorm:"bool default false index" json:"-"`
	DiscoverableByEmail          bool   `xorm:"bool default false index" json:"-"`
	OverdueTasksRemindersEnabled bool   `xorm:"bool default true index" json:"-"`
	OverdueTasksRemindersTime    string `xorm:"varchar(5) not null default '09:00'" json:"-"`
	DefaultProjectID             int64  `xorm:"bigint null index" json:"-"`
	WeekStart                    int    `xorm:"null" json:"-"`
	Language                     string `xorm:"varchar(50) null" json:"-" valid:"language"`
	Timezone                     string `xorm:"varchar(255) null" json:"-"`

	DeletionScheduledAt      time.Time `xorm:"datetime null" json:"-"`
	DeletionLastReminderSent time.Time `xorm:"datetime null" json:"-"`

	FrontendSettings   interface{}    `xorm:"json null" json:"-"`
	ExtraSettingsLinks map[string]any `xorm:"json null" json:"-"`

	ExportFileID int64 `xorm:"bigint null" json:"-"`

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

func (u *User) ShouldNotify() (bool, error) {
	s := db.NewSession()
	defer s.Close()
	user, err := getUser(s, &User{ID: u.ID}, true)
	if err != nil {
		return false, err
	}

	return user.Status != StatusDisabled, err
}

func (u *User) Lang() string {
	return u.Language
}

// GetID implements the Auth interface
func (u *User) GetID() int64 {
	return u.ID
}

// TableName returns the table name for users
func (*User) TableName() string {
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

func (u *User) GetFailedTOTPAttemptsKey() string {
	return "failed_totp_attempts_" + strconv.FormatInt(u.ID, 10)
}

func (u *User) GetFailedPasswordAttemptsKey() string {
	return "failed_password_attempts_" + strconv.FormatInt(u.ID, 10)
}

// GetFromAuth returns a user object from a web.Auth object and returns an error if the underlying type
// is not a user object
func GetFromAuth(a web.Auth) (*User, error) {
	u, is := a.(*User)
	if !is {
		typ := reflect.TypeOf(a)
		if typ.String() == "*models.LinkSharing" {
			return nil, &ErrMustNotBeLinkShare{}
		}
		return &User{}, fmt.Errorf("user is not user element, is %s", typ)
	}
	return u, nil
}

// APIUserPassword represents a user object without timestamps and a json password field.
type APIUserPassword struct {
	// The user's username. Cannot contain anything that looks like an url or whitespaces.
	Username string `json:"username" valid:"length(3|250),username" minLength:"3" maxLength:"250"`
	// The user's password in clear text. Only used when registering the user. The maximum limi is 72 bytes, which may be less than 72 characters. This is due to the limit in the bcrypt hashing algorithm used to store passwords in Vikunja.
	Password string `json:"password" valid:"bcrypt_password" minLength:"8" maxLength:"72"`
	// The user's email address
	Email string `json:"email" valid:"email,length(0|250)" maxLength:"250"`
}

// GetUserByID returns user by its ID
func GetUserByID(s *xorm.Session, id int64) (user *User, err error) {
	// Apparently xorm does otherwise look for all users but return only one, which leads to returning one even if the ID is 0
	if id < 1 {
		return &User{}, ErrUserDoesNotExist{}
	}

	return getUser(s, &User{ID: id}, false)
}

// GetUserByUsername gets a user from its username. This is an extra function to be able to add an extra error check.
func GetUserByUsername(s *xorm.Session, username string) (user *User, err error) {
	if username == "" {
		return &User{}, ErrUserDoesNotExist{}
	}

	return getUser(s, &User{Username: username}, false)
}

// GetUsersByUsername returns a slice of users with the provided usernames
func GetUsersByUsername(s *xorm.Session, usernames []string, withEmails bool) (users map[int64]*User, err error) {
	if len(usernames) == 0 {
		return
	}

	users = make(map[int64]*User)
	err = s.In("username", usernames).Find(&users)
	if err != nil {
		return
	}

	if !withEmails {
		for _, u := range users {
			u.Email = ""
		}
	}

	return
}

// GetUserWithEmail returns a user object with email
func GetUserWithEmail(s *xorm.Session, user *User) (userOut *User, err error) {
	return getUser(s, user, true)
}

// GetUsersByIDs returns a map of users from a slice of user ids
func GetUsersByIDs(s *xorm.Session, userIDs []int64) (users map[int64]*User, err error) {
	if len(userIDs) == 0 {
		return users, nil
	}

	return GetUsersByCond(s, builder.In("id", userIDs))
}

func GetUsersByCond(s *xorm.Session, cond builder.Cond) (users map[int64]*User, err error) {
	users = make(map[int64]*User)

	err = s.Where(cond).Find(&users)
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

	if userOut.OverdueTasksRemindersTime == "" {
		userOut.OverdueTasksRemindersTime = "9:00"
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

	if user.Issuer != IssuerLocal {
		return user, &ErrAccountIsNotLocal{UserID: user.ID}
	}

	// The user is invalid if they need to verify their email address
	if user.Status == StatusEmailConfirmationRequired {
		return &User{}, ErrEmailNotConfirmed{UserID: user.ID}
	}

	// Check the users password
	err = CheckUserPassword(user, u.Password)
	if err != nil {
		if IsErrWrongUsernameOrPassword(err) {
			handleFailedPassword(user)
		}
		return user, err
	}

	return user, nil
}

func (u *User) IsLocalUser() bool {
	return u.Issuer == IssuerLocal
}

func handleFailedPassword(user *User) {
	key := user.GetFailedPasswordAttemptsKey()
	err := keyvalue.IncrBy(key, 1)
	if err != nil {
		log.Errorf("Could not set failed password attempts: %s", err)
		return
	}

	a, _, err := keyvalue.Get(key)
	if err != nil {
		log.Errorf("Could not get failed password attempts for user %d: %s", user.ID, err)
		return
	}
	attempts, ok := a.(int64)
	if !ok {
		attemptsStr, ok := a.(string)
		if !ok {
			log.Errorf("Unexpected type for failed password attempts: %v", a)
			return
		}
		var err error
		attempts, err = strconv.ParseInt(attemptsStr, 10, 64)
		if err != nil {
			log.Errorf("Could not convert failed password attempts to int64: %v, value: %s", err, attemptsStr)
			return
		}
	}
	if attempts != 3 {
		return
	}

	err = notifications.Notify(user, &FailedLoginAttemptNotification{
		User: user,
	})
	if err != nil {
		log.Errorf("Could not send invalid password mail to user: %s", err)
		return
	}

	err = keyvalue.Del(key)
	if err != nil {
		log.Errorf("Could not remove failed password attempts: %s", err)
		return
	}
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

// GetCurrentUserFromDB gets a user from jwt claims and returns the full user from the db.
func GetCurrentUserFromDB(s *xorm.Session, c echo.Context) (user *User, err error) {
	u, err := GetCurrentUser(c)
	if err != nil {
		return nil, err
	}

	return GetUserByID(s, u.ID)
}

// GetCurrentUser returns the current user based on its jwt token
func GetCurrentUser(c echo.Context) (user *User, err error) {
	if apiUser, ok := c.Get("api_user").(*User); ok {
		return apiUser, nil
	}

	jwtinf, is := c.Get("user").(*jwt.Token)
	if jwtinf == nil {
		log.Error("No user found in context")
		return nil, ErrInvalidUserContext{Reason: "no user found in context"}
	}

	if !is {
		log.Errorf("User in context is not a JWT token, got type: %T", jwtinf)
		return nil, ErrInvalidUserContext{Reason: "user in context is not a JWT token"}
	}

	claims := jwtinf.Claims.(jwt.MapClaims)
	return GetUserFromClaims(claims)
}

// GetUserFromClaims Returns a new user from jwt claims
func GetUserFromClaims(claims jwt.MapClaims) (user *User, err error) {
	userID, err := getClaimAsInt(claims, "id")
	if err != nil {
		return nil, err
	}
	email, err := getClaimAsString(claims, "email")
	if err != nil {
		return nil, err
	}
	username, err := getClaimAsString(claims, "username")
	if err != nil {
		return nil, err
	}
	name, err := getClaimAsString(claims, "name")
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       userID,
		Email:    email,
		Username: username,
		Name:     name,
	}, nil
}

func getClaimAsInt(claims jwt.MapClaims, field string) (int64, error) {
	_, exists := claims[field]
	if !exists {
		return 0, &ErrInvalidClaimData{
			Field: field,
			Type:  "missing",
		}
	}

	value, ok := claims[field].(float64)
	if !ok {
		return 0, &ErrInvalidClaimData{
			Field: field,
			Type:  reflect.TypeOf(claims[field]).String(),
		}
	}
	return int64(value), nil
}

func getClaimAsString(claims jwt.MapClaims, field string) (string, error) {
	_, exists := claims[field]
	if !exists {
		return "", &ErrInvalidClaimData{
			Field: field,
			Type:  "missing",
		}
	}

	value, ok := claims[field].(string)
	if !ok {
		return "", &ErrInvalidClaimData{
			Field: field,
			Type:  reflect.TypeOf(claims[field]).String(),
		}
	}
	return value, nil
}

// UpdateUser updates a user
func UpdateUser(s *xorm.Session, user *User, forceOverride bool) (updatedUser *User, err error) {

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
	if user.Name == "" && !forceOverride {
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
			user.AvatarProvider != "upload" &&
			user.AvatarProvider != "marble" &&
			user.AvatarProvider != "ldap" &&
			user.AvatarProvider != "openid" {
			return updatedUser, &ErrInvalidAvatarProvider{AvatarProvider: user.AvatarProvider}
		}
	}

	// Check if we have a valid time zone
	if user.Timezone == "" {
		user.Timezone = config.GetTimeZone().String()
	}

	_, err = time.LoadLocation(user.Timezone)
	if err != nil {
		return nil, &ErrInvalidTimezone{Name: user.Timezone, LoadError: err}
	}

	frontendSettingsJSON, err := json.Marshal(user.FrontendSettings)
	if err != nil {
		return nil, err
	}
	user.FrontendSettings = frontendSettingsJSON

	// Update it
	_, err = s.
		ID(user.ID).
		Cols(
			"username",
			"email",
			"avatar_provider",
			"avatar_file_id",
			"status",
			"name",
			"email_reminders_enabled",
			"discoverable_by_name",
			"discoverable_by_email",
			"overdue_tasks_reminders_enabled",
			"default_project_id",
			"week_start",
			"language",
			"timezone",
			"overdue_tasks_reminders_time",
			"frontend_settings",
			"extra_settings_links",
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

func SetUserStatus(s *xorm.Session, user *User, status Status) (err error) {
	_, err = s.Where("id = ?", user.ID).
		Cols("status").
		Update(&User{Status: status})
	return
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
	_, err = s.
		Where("id = ?", user.ID).
		Update(&User{Password: hashed})
	if err != nil {
		return err
	}

	return err
}

// SetStatus sets a users status in the database
func (u *User) SetStatus(s *xorm.Session, status Status) (err error) {
	u.Status = status
	_, err = s.
		Where("id = ?", u.ID).
		Cols("status").
		Update(u)
	return
}
