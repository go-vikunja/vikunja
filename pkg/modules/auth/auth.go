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
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/humabridge"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"xorm.io/xorm"
)

// All valid auth types; AuthTypeUser lives in pkg/user since
// that package parses claims and can't import this one.
const (
	AuthTypeUnknown   int = 0
	AuthTypeUser          = user.AuthTypeUser
	AuthTypeLinkShare int = 2
)

// Token represents an authentication token
type Token struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"`
}

const RefreshTokenCookieName = "vikunja_refresh_token" //nolint:gosec // not a credential

const (
	RefreshTokenPathV1 = "/api/v1/user/token/refresh" //nolint:gosec // a route path, not a credential
	RefreshTokenPathV2 = "/api/v2/user/token/refresh" //nolint:gosec // a route path, not a credential
)

// getRefreshTokenCookiePaths prefixes each refresh endpoint with the base
// path from service.publicurl.
func getRefreshTokenCookiePaths() []string {
	basePath := ""
	if u, err := url.Parse(config.ServicePublicURL.GetString()); err == nil {
		basePath = strings.TrimRight(u.Path, "/")
	}

	refreshTokenPaths := []string{RefreshTokenPathV1, RefreshTokenPathV2}
	paths := make([]string, len(refreshTokenPaths))
	for i, p := range refreshTokenPaths {
		paths[i] = basePath + p
	}
	return paths
}

// SetRefreshTokenCookie sets one HttpOnly cookie per refresh endpoint (v1, v2),
// path-scoped so the browser only sends it on refresh requests. A single
// cookie at Path=/api would ship the long-lived token on every API request.
// Browsers match cookie paths by prefix, so each endpoint needs its own.
func SetRefreshTokenCookie(c *echo.Context, token string, maxAge int) {
	secure := strings.HasPrefix(config.ServicePublicURL.GetString(), "https")
	// SameSite=None so the cookie survives split-origin deployments where the
	// frontend and API are on different hosts, but browsers only accept that
	// with Secure=true; fall back to Lax on plain HTTP (local dev, E2E tests).
	sameSite := http.SameSiteLaxMode
	if secure {
		sameSite = http.SameSiteNoneMode
	}
	for _, path := range getRefreshTokenCookiePaths() {
		c.SetCookie(&http.Cookie{ //nolint:gosec // G124: Secure/SameSite are intentionally conditional on the https scheme (see above); HttpOnly is always set.
			Name:     RefreshTokenCookieName,
			Value:    token,
			Path:     path,
			MaxAge:   maxAge,
			HttpOnly: true,
			Secure:   secure,
			SameSite: sameSite,
		})
	}
}

// ClearRefreshTokenCookie removes the refresh token cookie.
func ClearRefreshTokenCookie(c *echo.Context) {
	SetRefreshTokenCookie(c, "", -1)
}

// IssuedUserToken bundles a freshly minted access token with the matching
// refresh token and the cookie max-age both v1 and v2 use to set the
// HttpOnly refresh cookie.
type IssuedUserToken struct {
	AccessToken  string
	RefreshToken string
	CookieMaxAge int
}

// IssueUserToken creates a session for the user and mints a JWT access token plus
// a refresh token for it. It is the transport-agnostic core both v1 (which writes
// the echo response) and v2 (Huma) call; callers set the refresh cookie and the
// Cache-Control header themselves via WriteUserAuthCookies. Pass oidc for
// OpenID Connect logins to store the logout data; nil otherwise.
func IssueUserToken(ctx context.Context, u *user.User, deviceInfo, ipAddress string, long bool, oidc *models.SessionOIDCData) (*IssuedUserToken, error) {
	s := db.NewSession()
	defer s.Close()

	session, err := models.CreateSession(s, u.ID, deviceInfo, ipAddress, long, oidc)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	t, err := NewUserJWTAuthtoken(u, session.ID)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if err := events.DispatchWithContext(ctx, &user.LoginSucceededEvent{User: u}); err != nil {
		log.Errorf("Could not dispatch login succeeded event: %s", err)
	}

	cookieMaxAge := int(config.ServiceJWTTTL.GetInt64())
	if long {
		cookieMaxAge = int(config.ServiceJWTTTLLong.GetInt64())
	}

	return &IssuedUserToken{
		AccessToken:  t,
		RefreshToken: session.RefreshToken,
		CookieMaxAge: cookieMaxAge,
	}, nil
}

// WriteUserAuthCookies sets the HttpOnly refresh-token cookie and the
// Cache-Control: no-store header on a response. The cookie is path-scoped to the
// refresh endpoint, so the browser only sends it there; JavaScript never sees the
// refresh token, which protects it from XSS. Shared by the v1 echo handlers and
// the v2 Huma handlers (which reach the echo context via the humabridge
// EchoContextKey stash on their request context).
func WriteUserAuthCookies(c *echo.Context, token *IssuedUserToken) {
	SetRefreshTokenCookie(c, token.RefreshToken, token.CookieMaxAge)
	c.Response().Header().Set("Cache-Control", "no-store")
}

// NewUserAuthTokenResponse creates a new user auth token response from a user object.
// Pass oidc for OpenID Connect logins to store the logout data; nil otherwise.
func NewUserAuthTokenResponse(u *user.User, c *echo.Context, long bool, oidc *models.SessionOIDCData) error {
	token, err := IssueUserToken(c.Request().Context(), u, c.Request().UserAgent(), c.RealIP(), long, oidc)
	if err != nil {
		return err
	}

	WriteUserAuthCookies(c, token)
	return c.JSON(http.StatusOK, Token{Token: token.AccessToken})
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
	claims["is_admin"] = u.IsAdmin
	claims["exp"] = exp
	claims["sid"] = sessionID
	claims["jti"] = uuid.New().String()

	return t.SignedString([]byte(config.ServiceSecret.GetString()))
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
	return t.SignedString([]byte(config.ServiceSecret.GetString()))
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
		s := db.NewSession()
		defer s.Close()
		return models.GetLinkShareFromClaims(s, claims)
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
		if user.IsErrUserStatusError(err) {
			return nil, nil, fmt.Errorf("API token %d owner account is disabled or locked", token.ID)
		}
		return nil, nil, err
	}

	return token, u, nil
}

// GetUserIDFromToken parses a raw JWT token string and returns the user ID.
// Only regular user tokens are accepted (not link shares).
// Returns 0 and an error if the token is invalid.
func GetUserIDFromToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(_ *jwt.Token) (any, error) {
		return []byte(config.ServiceSecret.GetString()), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, jwt.ErrTokenInvalidClaims
	}

	typ, ok := claims["type"].(float64)
	if !ok || int(typ) != AuthTypeUser {
		return 0, jwt.ErrTokenInvalidClaims
	}

	userIDFloat, ok := claims["id"].(float64)
	if !ok {
		return 0, jwt.ErrTokenInvalidClaims
	}

	return int64(userIDFloat), nil
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

// RefreshResult holds the result of a successful session refresh.
type RefreshResult struct {
	AccessToken     string
	NewRefreshToken string
	ExpiresIn       int64
	IsLongSession   bool
	SessionID       string
}

// RefreshSession looks up a session by its raw refresh token, validates it,
// rotates the refresh token, fetches the user, and generates a new JWT.
// It handles its own DB session (open/commit/rollback).
//
// On user status errors (disabled/locked), the session is deleted before
// returning the error so the caller can handle cleanup (e.g. clearing cookies).
func RefreshSession(rawRefreshToken string) (*RefreshResult, error) {
	s := db.NewSession()
	defer s.Close()

	session, err := models.GetSessionByRefreshToken(s, rawRefreshToken)
	if err != nil {
		_ = s.Rollback()
		if models.IsErrSessionNotFound(err) {
			return nil, &models.ErrInvalidRefreshToken{}
		}
		return nil, err
	}

	maxAge := time.Duration(config.ServiceJWTTTL.GetInt64()) * time.Second
	if session.IsLongSession {
		maxAge = time.Duration(config.ServiceJWTTTLLong.GetInt64()) * time.Second
	}
	if time.Since(session.LastActive) > maxAge {
		if _, err := s.Where("id = ?", session.ID).Delete(&models.Session{}); err != nil {
			_ = s.Rollback()
			return nil, err
		}
		if err := s.Commit(); err != nil {
			return nil, err
		}
		return nil, &models.ErrSessionExpired{}
	}

	if err := models.UpdateSessionLastActive(s, session.ID); err != nil {
		_ = s.Rollback()
		return nil, err
	}

	newRawToken, err := models.RotateRefreshToken(s, session)
	if err != nil {
		_ = s.Rollback()
		if models.IsErrSessionNotFound(err) {
			return nil, &models.ErrRefreshTokenAlreadyUsed{}
		}
		return nil, err
	}

	u, err := user.GetUserByID(s, session.UserID)
	if err != nil {
		if user.IsErrUserStatusError(err) {
			if _, delErr := s.Where("id = ?", session.ID).Delete(&models.Session{}); delErr != nil {
				_ = s.Rollback()
				return nil, delErr
			}
			if commitErr := s.Commit(); commitErr != nil {
				return nil, commitErr
			}
			return nil, err
		}
		_ = s.Rollback()
		return nil, err
	}

	accessToken, err := NewUserJWTAuthtoken(u, session.ID)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if err := s.Commit(); err != nil {
		return nil, err
	}

	return &RefreshResult{
		AccessToken:     accessToken,
		NewRefreshToken: newRawToken,
		ExpiresIn:       config.ServiceJWTTTLShort.GetInt64(),
		IsLongSession:   session.IsLongSession,
		SessionID:       session.ID,
	}, nil
}

// IsUnusableRefreshToken reports whether the caller should clear the refresh
// cookie. Transient errors must not clear it, nor a token rotated away by a
// concurrent refresh: that would delete the cookie the winning refresh just set.
func IsUnusableRefreshToken(err error) bool {
	var expired *models.ErrSessionExpired
	return errors.As(err, &expired) || user.IsErrUserStatusError(err)
}

// SessionIDFromContext reads the session id (the `sid` claim) off the user JWT
// in the echo context. It returns "" when there is no user JWT or no sid claim
// (API tokens and link shares carry no session), which callers treat as a no-op.
func SessionIDFromContext(c *echo.Context) string {
	raw := c.Get("user")
	if raw == nil {
		return ""
	}
	jwtinf, ok := raw.(*jwt.Token)
	if !ok {
		return ""
	}
	claims, ok := jwtinf.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}
	sid, _ := claims["sid"].(string)
	return sid
}

// GetAuthFromContext retrieves the authenticated web.Auth from a plain
// context.Context, bridging Huma handlers to Vikunja's echo JWT flow. The
// humabridge group middleware stashes the *echo.Context under EchoContextKey
// first.
func GetAuthFromContext(ctx context.Context) (web.Auth, error) {
	ec, ok := ctx.Value(humabridge.EchoContextKey).(*echo.Context)
	if !ok {
		return nil, fmt.Errorf("no echo.Context on request context; are you calling GetAuthFromContext from a Huma handler mounted via humabridge?")
	}
	return GetAuthFromClaims(ec)
}
