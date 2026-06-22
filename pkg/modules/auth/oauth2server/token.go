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
	"context"

	"net/http"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
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

// TokenRequest holds the parameters of a POST /oauth/token request. v1 binds it
// from JSON; v2 accepts spec-compliant application/x-www-form-urlencoded as well
// (form tags mirror the json names).
type TokenRequest struct {
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
	var req TokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	resp, err := ExchangeToken(c.Request().Context(), &req, c.Request().UserAgent(), c.RealIP())
	if err != nil {
		return err
	}

	c.Response().Header().Set("Cache-Control", "no-store")
	return c.JSON(http.StatusOK, resp)
}

// ExchangeToken runs the grant-type dispatch and token issuance for the OAuth
// token endpoint, independent of the HTTP layer. Callers own request binding and
// the Cache-Control: no-store response header. deviceInfo/ipAddress are recorded
// on the session created for the authorization_code grant.
func ExchangeToken(ctx context.Context, req *TokenRequest, deviceInfo, ipAddress string) (*TokenResponse, error) {
	switch req.GrantType {
	case "authorization_code":
		return exchangeAuthorizationCode(ctx, req, deviceInfo, ipAddress)
	case "refresh_token":
		return exchangeRefreshToken(req)
	default:
		return nil, &models.ErrOAuthInvalidGrantType{}
	}
}

func exchangeAuthorizationCode(ctx context.Context, req *TokenRequest, deviceInfo, ipAddress string) (*TokenResponse, error) {
	s := db.NewSession()
	defer s.Close()

	// Look up and delete the authorization code (single-use)
	oauthCode, err := models.GetAndDeleteOAuthCode(s, req.Code)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	// Validate client_id matches
	if oauthCode.ClientID != req.ClientID {
		_ = s.Rollback()
		return nil, &models.ErrOAuthClientNotFound{}
	}

	// Validate redirect_uri matches
	if oauthCode.RedirectURI != req.RedirectURI {
		_ = s.Rollback()
		return nil, &models.ErrOAuthInvalidRedirectURI{}
	}

	// Verify PKCE
	if !VerifyPKCE(req.CodeVerifier, oauthCode.CodeChallenge, oauthCode.CodeChallengeMethod) {
		_ = s.Rollback()
		return nil, &models.ErrOAuthPKCEVerifyFailed{}
	}

	// Create a session (reuses existing session infrastructure)
	session, err := models.CreateSession(s, oauthCode.UserID, deviceInfo, ipAddress, false, nil)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	u, err := user.GetUserByID(s, oauthCode.UserID)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	// Generate JWT
	accessToken, err := auth.NewUserJWTAuthtoken(u, session.ID)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if err := s.Commit(); err != nil {
		return nil, err
	}

	// The code exchange mints a fresh session, so it is a login for the
	// audit trail, same as NewUserAuthTokenResponse.
	if err := events.DispatchWithContext(ctx, &user.LoginSucceededEvent{User: u}); err != nil {
		log.Errorf("Could not dispatch login succeeded event: %s", err)
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "bearer",
		ExpiresIn:    config.ServiceJWTTTLShort.GetInt64(),
		RefreshToken: session.RefreshToken,
	}, nil
}

func exchangeRefreshToken(req *TokenRequest) (*TokenResponse, error) {
	result, err := auth.RefreshSession(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  result.AccessToken,
		TokenType:    "bearer",
		ExpiresIn:    result.ExpiresIn,
		RefreshToken: result.NewRefreshToken,
	}, nil
}
