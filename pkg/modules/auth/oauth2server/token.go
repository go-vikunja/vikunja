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
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

// TokenResponse is the OAuth 2.0 token response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// tokenRequest holds the form-decoded body of a POST /oauth/token request.
type tokenRequest struct {
	GrantType    string `json:"grant_type" form:"grant_type"`
	Code         string `json:"code" form:"code"`
	ClientID     string `json:"client_id" form:"client_id"`
	RedirectURI  string `json:"redirect_uri" form:"redirect_uri"`
	CodeVerifier string `json:"code_verifier" form:"code_verifier"`
	RefreshToken string `json:"refresh_token" form:"refresh_token"`
}

// HandleToken handles POST /oauth/token.
// Supports grant_type=authorization_code and grant_type=refresh_token.
func HandleToken(c *echo.Context) error {
	var req tokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	switch req.GrantType {
	case "authorization_code":
		return handleAuthorizationCodeGrant(c, &req)
	case "refresh_token":
		return handleRefreshTokenGrant(c, &req)
	default:
		return &models.ErrOAuthInvalidGrantType{}
	}
}

func handleAuthorizationCodeGrant(c *echo.Context, req *tokenRequest) error {
	// Validate client_id
	if !ValidateClient(req.ClientID) {
		return &models.ErrOAuthClientNotFound{}
	}

	s := db.NewSession()
	defer s.Close()

	// Look up and delete the authorization code (single-use)
	oauthCode, err := models.GetAndDeleteOAuthCode(s, req.Code)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	// Validate client_id matches
	if oauthCode.ClientID != req.ClientID {
		_ = s.Rollback()
		return &models.ErrOAuthClientNotFound{}
	}

	// Validate redirect_uri matches
	if oauthCode.RedirectURI != req.RedirectURI {
		_ = s.Rollback()
		return &models.ErrOAuthInvalidRedirectURI{}
	}

	// Verify PKCE
	if !VerifyPKCE(req.CodeVerifier, oauthCode.CodeChallenge, oauthCode.CodeChallengeMethod) {
		_ = s.Rollback()
		return &models.ErrOAuthPKCEVerifyFailed{}
	}

	// Create a session (reuses existing session infrastructure)
	deviceInfo := c.Request().UserAgent()
	ipAddress := c.RealIP()
	session, err := models.CreateSession(s, oauthCode.UserID, deviceInfo, ipAddress, false)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	// Get the user for JWT claims
	u, err := user.GetUserByID(s, oauthCode.UserID)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	// Check account status
	if u.Status == user.StatusDisabled {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusForbidden, "Account disabled.")
	}

	// Generate JWT
	accessToken, err := auth.NewUserJWTAuthtoken(u, session.ID)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		return err
	}

	c.Response().Header().Set("Cache-Control", "no-store")
	return c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "bearer",
		ExpiresIn:    config.ServiceJWTTTLShort.GetInt64(),
		RefreshToken: session.RefreshToken,
	})
}

func handleRefreshTokenGrant(c *echo.Context, req *tokenRequest) error {
	// Validate client_id
	if !ValidateClient(req.ClientID) {
		return &models.ErrOAuthClientNotFound{}
	}

	s := db.NewSession()
	defer s.Close()

	// Look up session by refresh token hash
	session, err := models.GetSessionByRefreshToken(s, req.RefreshToken)
	if err != nil {
		_ = s.Rollback()
		if models.IsErrSessionNotFound(err) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired refresh token.")
		}
		return err
	}

	// Check session expiry
	maxAge := time.Duration(config.ServiceJWTTTL.GetInt64()) * time.Second
	if session.IsLongSession {
		maxAge = time.Duration(config.ServiceJWTTTLLong.GetInt64()) * time.Second
	}
	if time.Since(session.LastActive) > maxAge {
		_, _ = s.Where("id = ?", session.ID).Delete(&models.Session{})
		_ = s.Commit()
		return echo.NewHTTPError(http.StatusUnauthorized, "Session expired.")
	}

	// Update last active
	if err := models.UpdateSessionLastActive(s, session.ID); err != nil {
		_ = s.Rollback()
		return err
	}

	// Rotate refresh token
	newRawToken, err := models.RotateRefreshToken(s, session)
	if err != nil {
		_ = s.Rollback()
		if models.IsErrSessionNotFound(err) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Refresh token already used.")
		}
		return err
	}

	// Get user
	u, err := user.GetUserByID(s, session.UserID)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	// Check account status
	if u.Status == user.StatusDisabled {
		_, _ = s.Where("id = ?", session.ID).Delete(&models.Session{})
		_ = s.Commit()
		return echo.NewHTTPError(http.StatusUnauthorized, "Account disabled.")
	}

	// Generate new JWT
	accessToken, err := auth.NewUserJWTAuthtoken(u, session.ID)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		return err
	}

	c.Response().Header().Set("Cache-Control", "no-store")
	return c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "bearer",
		ExpiresIn:    config.ServiceJWTTTLShort.GetInt64(),
		RefreshToken: newRawToken,
	})
}
