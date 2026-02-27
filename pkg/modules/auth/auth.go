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
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
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

const RefreshTokenCookieName = "vikunja_refresh_token"      //nolint:gosec // not a credential
const refreshTokenCookiePath = "/api/v1/user/token/refresh" //nolint:gosec // not a credential

// SetRefreshTokenCookie sets an HttpOnly cookie containing the refresh token.
// The cookie is path-scoped to the refresh endpoint so the browser only sends
// it on refresh requests. HttpOnly prevents JavaScript access (XSS protection).
func SetRefreshTokenCookie(c *echo.Context, token string, maxAge int) {
	secure := strings.HasPrefix(config.ServicePublicURL.GetString(), "https")
	c.SetCookie(&http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    token,
		Path:     refreshTokenCookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})
}

// ClearRefreshTokenCookie removes the refresh token cookie.
func ClearRefreshTokenCookie(c *echo.Context) {
	SetRefreshTokenCookie(c, "", -1)
}

// NewUserAuthTokenResponse creates a new user auth token response from a user object.
func NewUserAuthTokenResponse(u *user.User, c *echo.Context, long bool) error {
	s := db.NewSession()
	defer s.Close()

	deviceInfo := c.Request().UserAgent()
	ipAddress := c.RealIP()

	session, err := models.CreateSession(s, u.ID, deviceInfo, ipAddress, long)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	t, err := NewUserJWTAuthtoken(u, session.ID)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	// Set the refresh token as an HttpOnly cookie. The cookie is path-scoped
	// to the refresh endpoint, so the browser only sends it there. JavaScript
	// never sees the refresh token — this protects it from XSS.
	cookieMaxAge := int(config.ServiceJWTTTL.GetInt64())
	if long {
		cookieMaxAge = int(config.ServiceJWTTTLLong.GetInt64())
	}
	SetRefreshTokenCookie(c, session.RefreshToken, cookieMaxAge)

	c.Response().Header().Set("Cache-Control", "no-store")
	return c.JSON(http.StatusOK, Token{Token: t})
}

// NewUserJWTAuthtoken generates and signs a new short-lived jwt token for a user.
// The token includes the session UUID as the `sid` claim. This is a global
// function to be able to call it from web tests.
func NewUserJWTAuthtoken(u *user.User, sessionID string) (token string, err error) {
	t := jwt.New(jwt.SigningMethodHS256)

	var ttl = time.Duration(config.ServiceJWTTTLShort.GetInt64())
	var exp = time.Now().Add(time.Second * ttl).Unix()

	claims := t.Claims.(jwt.MapClaims)
	claims["type"] = AuthTypeUser
	claims["id"] = u.ID
	claims["username"] = u.Username
	claims["exp"] = exp
	claims["sid"] = sessionID
	claims["jti"] = uuid.New().String()

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

	// Generate encoded token and send it as response.
	return t.SignedString([]byte(config.ServiceJWTSecret.GetString()))
}

// GetAuthFromClaims returns a web.Auth object from jwt claims
func GetAuthFromClaims(c *echo.Context) (a web.Auth, err error) {
	// check if we have a token in context and use it if that's the case
	if c.Get("api_token") != nil {
		apiToken := c.Get("api_token").(*models.APIToken)
		s := db.NewSession()
		defer s.Close()
		u, err := user.GetUserByID(s, apiToken.OwnerID)
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
	typFloat, is := claims["type"].(float64)
	if !is {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid JWT token.")
	}
	typ := int(typFloat)
	if typ == AuthTypeLinkShare && config.ServiceEnableLinkSharing.GetBool() {
		return models.GetLinkShareFromClaims(claims)
	}
	if typ == AuthTypeUser {
		return user.GetUserFromClaims(claims)
	}
	return nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid JWT token.")
}

// ValidateAPITokenString looks up an API token by its raw string, checks expiry,
// and returns the token and its owner. This is the shared validation logic used
// by both the HTTP middleware and WebSocket auth.
func ValidateAPITokenString(tokenString string) (*models.APIToken, *user.User, error) {
	s := db.NewSession()
	defer s.Close()

	token, err := models.GetTokenFromTokenString(s, tokenString)
	if err != nil {
		return nil, nil, err
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, nil, fmt.Errorf("API token %d expired on %s", token.ID, token.ExpiresAt.String())
	}

	u, err := user.GetUserByID(s, token.OwnerID)
	if err != nil {
		return nil, nil, err
	}

	return token, u, nil
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
