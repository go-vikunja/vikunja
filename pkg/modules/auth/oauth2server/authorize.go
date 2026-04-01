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
	log.Debugf("OAuth authorize request: method=%s, path=%s, query=%s", c.Request().Method, c.Request().URL.Path, c.Request().URL.RawQuery)

	println("test log")

	var req authorizeRequest
	if err := c.Bind(&req); err != nil {
		log.Debugf("OAuth authorize: failed to bind request: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	log.Debugf("OAuth authorize request body: response_type=%s, client_id=%s, redirect_uri=%s, state=%s, code_challenge=%s, code_challenge_method=%s",
		req.ResponseType, req.ClientID, req.RedirectURI, req.State, req.CodeChallenge, req.CodeChallengeMethod)

	// Validate response_type
	if req.ResponseType != "code" {
		log.Debugf("OAuth authorize: invalid response_type: %s", req.ResponseType)
		return echo.NewHTTPError(http.StatusBadRequest, "response_type must be 'code'")
	}

	// Validate redirect_uri
	if !ValidateRedirectURI(req.RedirectURI) {
		log.Debugf("OAuth authorize: invalid redirect_uri: %s", req.RedirectURI)
		return &models.ErrOAuthInvalidRedirectURI{}
	}

	// Validate PKCE (required)
	if req.CodeChallenge == "" || req.CodeChallengeMethod != "S256" {
		log.Debugf("OAuth authorize: missing or invalid PKCE: code_challenge=%s, code_challenge_method=%s", req.CodeChallenge, req.CodeChallengeMethod)
		return &models.ErrOAuthMissingPKCE{}
	}

	// Get the authenticated user from the middleware
	u, err := user.GetCurrentUser(c)
	if err != nil {
		log.Debugf("OAuth authorize: failed to get current user: %v", err)
		return err
	}

	log.Debugf("OAuth authorize: authenticated user ID: %d", u.ID)

	s := db.NewSession()
	defer s.Close()

	fullUser, err := user.GetUserByID(s, u.ID)
	if err != nil {
		log.Debugf("OAuth authorize: failed to get full user by ID %d: %v", u.ID, err)
		_ = s.Rollback()
		return err
	}

	log.Debugf("OAuth authorize: creating OAuth code for user %d, client_id=%s, redirect_uri=%s", fullUser.ID, req.ClientID, req.RedirectURI)

	code, err := models.CreateOAuthCode(s, fullUser.ID, req.ClientID, req.RedirectURI, req.CodeChallenge, req.CodeChallengeMethod)
	if err != nil {
		log.Debugf("OAuth authorize: failed to create OAuth code: %v", err)
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		log.Debugf("OAuth authorize: failed to commit transaction: %v", err)
		return err
	}

	log.Debugf("OAuth authorize: successfully created code, returning response")

	return c.JSON(http.StatusOK, AuthorizeResponse{
		Code:        code,
		RedirectURI: req.RedirectURI,
		State:       req.State,
	})
}
