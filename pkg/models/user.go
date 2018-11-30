//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/web"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"golang.org/x/crypto/bcrypt"
	"reflect"
)

// UserLogin Object to recive user credentials in JSON format
type UserLogin struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

// User holds information about an user
type User struct {
	ID       int64  `xorm:"int(11) autoincr not null unique pk" json:"id"`
	Username string `xorm:"varchar(250) not null unique" json:"username" valid:"length(3|250)"`
	Password string `xorm:"varchar(250) not null" json:"-"`
	Email    string `xorm:"varchar(250)" json:"email" valid:"email,length(0|250)"`
	IsActive bool   `json:"-"`

	PasswordResetToken string `xorm:"varchar(450)" json:"-"`
	EmailConfirmToken  string `xorm:"varchar(450)" json:"-"`

	Created int64 `xorm:"created" json:"created"`
	Updated int64 `xorm:"updated" json:"updated"`

	web.Auth `xorm:"-" json:"-"`
}

// AuthDummy implements the auth of the crud handler
func (User) AuthDummy() {}

// TableName returns the table name for users
func (User) TableName() string {
	return "users"
}

func getUserForRights(a web.Auth) *User {
	u, err := getUserWithError(a)
	if err != nil {
		log.Log.Error(err.Error())
	}
	return u
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
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

// APIFormat formats an API User into a normal user struct
func (apiUser *APIUserPassword) APIFormat() User {
	return User{
		ID:       apiUser.ID,
		Username: apiUser.Username,
		Password: apiUser.Password,
		Email:    apiUser.Email,
	}
}

// GetUserByID gets informations about a user by its ID
func GetUserByID(id int64) (user User, err error) {
	// Apparently xorm does otherwise look for all users but return only one, which leads to returing one even if the ID is 0
	if id < 1 {
		return User{}, ErrUserDoesNotExist{}
	}

	return GetUser(User{ID: id})
}

// GetUser gets a user object
func GetUser(user User) (userOut User, err error) {
	userOut = user
	exists, err := x.Get(&userOut)

	if !exists {
		return User{}, ErrUserDoesNotExist{UserID: user.ID}
	}

	return userOut, err
}

// CheckUserCredentials checks user credentials
func CheckUserCredentials(u *UserLogin) (User, error) {
	// Check if we have any credentials
	if u.Password == "" || u.Username == "" {
		return User{}, ErrNoUsernamePassword{}
	}

	// Check if the user exists
	user, err := GetUser(User{Username: u.Username})
	if err != nil {
		return User{}, err
	}

	// User is invalid if it needs to verify its email address
	if !user.IsActive {
		return User{}, ErrEmailNotConfirmed{UserID: user.ID}
	}

	// Check the users password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return User{}, ErrWrongUsernameOrPassword{}
		}
		return User{}, err
	}

	return user, nil
}

// GetCurrentUser returns the current user based on its jwt token
func GetCurrentUser(c echo.Context) (user *User, err error) {
	jwtinf := c.Get("user").(*jwt.Token)
	claims := jwtinf.Claims.(jwt.MapClaims)
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
