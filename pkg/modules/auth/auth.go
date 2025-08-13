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

package auth

import (
	"fmt"
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// These are all valid auth types
const (
	AuthTypeUnknown int = iota
	AuthTypeUser
	AuthTypeLinkShare
)

// Token represents an authentication token
type Token struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

// NewUserAuthTokenResponse creates a new user auth token response from a user object.
func NewUserAuthTokenResponse(u *user.User, c echo.Context, long bool) error {
	t, err := NewUserJWTAuthtoken(u, long)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Token{Token: t})
}

// NewUserJWTAuthtoken generates and signs a new jwt token for a user. This is a global function to be able to call it from web tests.
func NewUserJWTAuthtoken(u *user.User, long bool) (token string, err error) {
	t := jwt.New(jwt.SigningMethodHS256)

	var ttl = time.Duration(config.ServiceJWTTTL.GetInt64())
	if long {
		ttl = time.Duration(config.ServiceJWTTTLLong.GetInt64())
	}
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
	claims["long"] = long

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
	claims["project_id"] = share.ProjectID
	claims["permission"] = share.Permission
	claims["sharedByID"] = share.SharedByID
	claims["exp"] = exp
	claims["isLocalUser"] = true // Link shares are always local

	// Generate encoded token and send it as response.
	return t.SignedString([]byte(config.ServiceJWTSecret.GetString()))
}

// GetAuthFromClaims returns a web.Auth object from jwt claims
func GetAuthFromClaims(c echo.Context) (a web.Auth, err error) {
	// check if we have a token in context and use it if that's the case
	if c.Get("api_token") != nil {
		apiToken := c.Get("api_token").(*models.APIToken)
		u, err := user.GetUserByID(db.NewSession(), apiToken.OwnerID)
		if err != nil {
			return nil, err
		}
		return u, nil
	}

	jwtinf, is := c.Get("user").(*jwt.Token)
	if !is {
		return nil, fmt.Errorf("user in context is not jwt token")
	}
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

func CreateUserWithRandomUsername(s *xorm.Session, uu *user.User) (u *user.User, err error) {
	// Check if we actually have a preferred username and generate a random one right away if we don't
	for {
		if uu.Username == "" {
			uu.Username = petname.Generate(3, "-")
		}

		u, err = user.CreateUser(s, uu)
		if err == nil {
			break
		}

		if !user.IsErrUsernameExists(err) {
			return nil, err
		}

		// If their preferred username is already taken, generate a new one
		uu.Username = petname.Generate(3, "-")
	}

	// And create their project
	err = models.CreateNewProjectForUser(s, u)
	return
}
