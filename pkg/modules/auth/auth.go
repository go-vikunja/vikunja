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

package auth

import (
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

// These are all valid auth types
const (
	AuthTypeUnknown int = iota
	AuthTypeUser
	AuthTypeLinkShare
)

// Token represents an authentification token
type Token struct {
	Token string `json:"token"`
}

// NewUserAuthTokenResponse creates a new user auth token response from a user object.
func NewUserAuthTokenResponse(u *user.User, c echo.Context) error {
	t, err := NewUserJWTAuthtoken(u)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Token{Token: t})
}

// NewUserJWTAuthtoken generates and signes a new jwt token for a user. This is a global function to be able to call it from integration tests.
func NewUserJWTAuthtoken(u *user.User) (token string, err error) {
	t := jwt.New(jwt.SigningMethodHS256)

	var ttl = time.Duration(config.ServiceJWTTTL.GetInt64())
	var exp = time.Now().Add(time.Second * ttl).Unix()

	// Set claims
	claims := t.Claims.(jwt.MapClaims)
	claims["type"] = AuthTypeUser
	claims["id"] = u.ID
	claims["username"] = u.Username
	claims["email"] = u.Email
	claims["exp"] = exp
	claims["name"] = u.Name
	claims["emailRemindersEnabled"] = u.EmailRemindersEnabled
	claims["isLocalUser"] = u.Issuer == user.IssuerLocal

	// Generate encoded token and send it as response.
	return t.SignedString([]byte(config.ServiceJWTSecret.GetString()))
}

// NewLinkShareJWTAuthtoken creates a new jwt token from a link share
func NewLinkShareJWTAuthtoken(share *models.LinkSharing) (token string, err error) {
	t := jwt.New(jwt.SigningMethodHS256)

	var ttl = time.Duration(config.ServiceJWTTTL.GetInt64())
	var exp = time.Now().Add(time.Second * ttl).Unix()

	// Set claims
	claims := t.Claims.(jwt.MapClaims)
	claims["type"] = AuthTypeLinkShare
	claims["id"] = share.ID
	claims["hash"] = share.Hash
	claims["list_id"] = share.ListID
	claims["right"] = share.Right
	claims["sharedByID"] = share.SharedByID
	claims["exp"] = exp
	claims["isLocalUser"] = true // Link shares are always local

	// Generate encoded token and send it as response.
	return t.SignedString([]byte(config.ServiceJWTSecret.GetString()))
}

// GetAuthFromClaims returns a web.Auth object from jwt claims
func GetAuthFromClaims(c echo.Context) (a web.Auth, err error) {
	jwtinf := c.Get("user").(*jwt.Token)
	claims := jwtinf.Claims.(jwt.MapClaims)
	typ := int(claims["type"].(float64))
	if typ == AuthTypeLinkShare && config.ServiceEnableLinkSharing.GetBool() {
		return models.GetLinkShareFromClaims(claims)
	}
	if typ == AuthTypeUser {
		return user.GetUserFromClaims(claims)
	}
	return nil, echo.NewHTTPError(http.StatusBadRequest, models.Message{Message: "Invalid JWT token."})
}
