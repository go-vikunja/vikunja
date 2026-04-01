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
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

// authorizeRequest represents the JSON body for the authorize endpoint.
type authorizeRequest struct {
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
	
	var req authorizeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate response_type
	if req.ResponseType != "code" {
		return echo.NewHTTPError(http.StatusBadRequest, "response_type must be 'code'")
	}
	s := db.NewSession()
	defer s.Close()

	client, err := models.GetOAuthClientByClientID(s, req.ClientID)
	if err != nil {
		log.Warningf("error getting OAuth client: %s %v", req.ClientID, err)
		return &models.ErrOAuthClientNotFound{}
	}

	// Validate redirect_uri
	if !ValidateRedirectURI(req, client) {
		return &models.ErrOAuthInvalidRedirectURI{}
	}

	// Validate PKCE (required)
	if req.CodeChallenge == "" || req.CodeChallengeMethod != "S256" {
		return &models.ErrOAuthMissingPKCE{}
	}

	// Get the authenticated user from the middleware
	u, err := user.GetCurrentUser(c)
	if err != nil {
		return err
	}

	fullUser, err := user.GetUserByID(s, u.ID)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	code, err := models.CreateOAuthCode(s, fullUser.ID, req.ClientID, req.RedirectURI, req.CodeChallenge, req.CodeChallengeMethod)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, AuthorizeResponse{
		Code:        code,
		RedirectURI: req.RedirectURI,
		State:       req.State,
	})
}
