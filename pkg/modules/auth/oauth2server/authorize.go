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

package oauth2server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

// HandleAuthorize handles GET /oauth/authorize.
// It validates the OAuth parameters, checks for an authenticated session,
// creates an authorization code, and redirects to the redirect_uri.
func HandleAuthorize(c *echo.Context) error {
	// Parse and validate query parameters
	responseType := c.QueryParam("response_type")
	clientID := c.QueryParam("client_id")
	redirectURI := c.QueryParam("redirect_uri")
	state := c.QueryParam("state")
	codeChallenge := c.QueryParam("code_challenge")
	codeChallengeMethod := c.QueryParam("code_challenge_method")

	// Validate response_type
	if responseType != "code" {
		return echo.NewHTTPError(http.StatusBadRequest, "response_type must be 'code'")
	}

	// Validate client_id
	if !ValidateClient(clientID) {
		return &models.ErrOAuthClientNotFound{}
	}

	// Validate redirect_uri
	if !ValidateRedirectURI(clientID, redirectURI) {
		return &models.ErrOAuthInvalidRedirectURI{}
	}

	// Validate PKCE (required)
	if codeChallenge == "" || codeChallengeMethod != "S256" {
		return &models.ErrOAuthMissingPKCE{}
	}

	// Check for an existing authenticated session via JWT bearer token
	u, err := getUserFromRequest(c)
	if err != nil || u == nil {
		// Not authenticated — redirect to frontend login with a redirect parameter
		// so the user returns here after login.
		frontendURL := strings.TrimSuffix(config.ServicePublicURL.GetString(), "/")
		authorizeURL := buildAuthorizeURL(c)
		loginURL := fmt.Sprintf("%s/login?redirect=%s", frontendURL, url.QueryEscape(authorizeURL))
		return c.Redirect(http.StatusFound, loginURL)
	}

	// User is authenticated — create the authorization code
	s := db.NewSession()
	defer s.Close()

	code, err := models.CreateOAuthCode(s, u.ID, clientID, redirectURI, codeChallenge, codeChallengeMethod)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		return err
	}

	// Redirect to the client's redirect_uri with the code and state
	redirectTo, err := url.Parse(redirectURI)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid redirect_uri")
	}

	q := redirectTo.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	redirectTo.RawQuery = q.Encode()

	return c.Redirect(http.StatusFound, redirectTo.String())
}

// getUserFromRequest attempts to extract an authenticated user from the request.
// Checks for a JWT bearer token in the Authorization header or a valid
// refresh token cookie. Returns nil if no valid session is found.
func getUserFromRequest(c *echo.Context) (*user.User, error) {
	// Try JWT bearer token first
	authHeader := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.ServiceJWTSecret.GetString()), nil
		})
		if err == nil && token.Valid {
			claims, ok := token.Claims.(jwt.MapClaims)
			if ok {
				return user.GetUserFromClaims(claims)
			}
		}
	}

	// Try refresh token cookie — the user may have an active browser session
	// but the short-lived JWT may have expired.
	cookie, err := c.Cookie("vikunja_refresh_token")
	if err == nil && cookie.Value != "" {
		s := db.NewSession()
		defer s.Close()

		session, err := models.GetSessionByRefreshToken(s, cookie.Value)
		if err == nil {
			u, err := user.GetUserByID(s, session.UserID)
			if err == nil {
				return u, nil
			}
		}
	}

	return nil, nil
}

// buildAuthorizeURL reconstructs the full authorize URL from the current request
// so we can pass it as a redirect parameter after login.
func buildAuthorizeURL(c *echo.Context) string {
	publicURL := strings.TrimSuffix(config.ServicePublicURL.GetString(), "/")
	return fmt.Sprintf("%s%s?%s", publicURL, c.Request().URL.Path, c.Request().URL.RawQuery)
}
