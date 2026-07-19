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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

// AuthorizeRequest represents the body for the authorize endpoint.
type AuthorizeRequest struct {
	ResponseType        string `json:"response_type"`
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	State               string `json:"state"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

// AuthorizeResponse is returned on successful authorization code creation.
type AuthorizeResponse struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
	State       string `json:"state"`
}

// HandleAuthorize handles POST /oauth/authorize.
// It validates the OAuth parameters, creates an authorization code, and
// returns it as JSON. Authentication is handled by the token middleware.
func HandleAuthorize(c *echo.Context) error {
	if c.Get("api_token") != nil {
		return echo.NewHTTPError(http.StatusForbidden, "API tokens cannot be used to authorize OAuth clients")
	}

	var req AuthorizeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Get the authenticated user from the middleware
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	resp, err := Authorize(&req, u.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// Authorize validates the OAuth authorization parameters for the given
// authenticated user and creates a single-use authorization code, independent
// of the HTTP layer. Callers own request binding and resolving the user.
func Authorize(req *AuthorizeRequest, userID int64) (*AuthorizeResponse, error) {
	// Validate response_type
	if req.ResponseType != "code" {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "response_type must be 'code'")
	}

	// Validate redirect_uri
	if !ValidateRedirectURI(req.RedirectURI) {
		return nil, &models.ErrOAuthInvalidRedirectURI{}
	}

	// Validate PKCE (required)
	if req.CodeChallenge == "" || req.CodeChallengeMethod != "S256" {
		return nil, &models.ErrOAuthMissingPKCE{}
	}

	s := db.NewSession()
	defer s.Close()

	fullUser, err := user.GetUserByID(s, userID)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	code, err := models.CreateOAuthCode(s, fullUser.ID, req.ClientID, req.RedirectURI, req.CodeChallenge, req.CodeChallengeMethod)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if err := s.Commit(); err != nil {
		return nil, err
	}

	return &AuthorizeResponse{
		Code:        code,
		RedirectURI: req.RedirectURI,
		State:       req.State,
	}, nil
}
