//   Vikunja is a todo-list application to facilitate your life.
//   Copyright 2019 Vikunja and contributors. All rights reserved.
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v1

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"github.com/dgrijalva/jwt-go"
	"time"
)

// These are all valid auth types
const (
	AuthTypeUnknown int = iota
	AuthTypeUser
	AuthTypeLinkShare
)

// NewUserJWTAuthtoken generates and signes a new jwt token for a user. This is a global function to be able to call it from integration tests.
func NewUserJWTAuthtoken(user *models.User) (token string, err error) {
	t := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := t.Claims.(jwt.MapClaims)
	claims["type"] = AuthTypeUser
	claims["id"] = user.ID
	claims["username"] = user.Username
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	claims["avatar"] = user.AvatarURL

	// Generate encoded token and send it as response.
	return t.SignedString([]byte(config.ServiceJWTSecret.GetString()))
}

// NewLinkShareJWTAuthtoken creates a new jwt token from a link share
func NewLinkShareJWTAuthtoken(share *models.LinkSharing) (token string, err error) {
	t := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := t.Claims.(jwt.MapClaims)
	claims["type"] = AuthTypeLinkShare
	claims["id"] = share.ID
	claims["hash"] = share.Hash
	claims["listID"] = share.ListID
	claims["right"] = share.Right
	claims["sharedByID"] = share.SharedByID
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// Generate encoded token and send it as response.
	return t.SignedString([]byte(config.ServiceJWTSecret.GetString()))
}
