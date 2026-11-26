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
	Scope               string `json:"scope,omitempty" doc:"Space-separated scopes. With mcp:read or mcp:write the client receives an API token for the MCP endpoint instead of a session. Unknown values are ignored."`
	Resource            string `json:"resource,omitempty" doc:"RFC 8707 resource indicator. Must be this instance's MCP endpoint; implies a scoped grant of mcp:read mcp:write when scope is omitted."`
}

// AuthorizeResponse is returned on successful authorization code creation.
type AuthorizeResponse struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
	State       string `json:"state"`
	Scope       string `json:"scope,omitempty" doc:"The granted scope. Empty when the client receives a session."`
}

// Server issues authorization codes and tokens. Scoped grants are API tokens for resource.
type Server struct {
	resource Resource
}

func NewServer(resource Resource) *Server {
	return &Server{resource: resource}
}

// HandleAuthorize handles POST /oauth/authorize.
// It validates the OAuth parameters, creates an authorization code, and
// returns it as JSON. Authentication is handled by the token middleware.
// @Summary OAuth 2.0 authorize endpoint
// @Description Creates an authorization code for an OAuth 2.0 client on behalf of the authenticated user. PKCE is required. API tokens cannot be used to authorize a client. With an mcp:read or mcp:write scope, or a resource, the code is exchanged for an API token that only works on the MCP endpoint.
// @tags auth
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param request body AuthorizeRequest true "The authorization request"
// @Success 200 {object} AuthorizeResponse "The authorization code and the redirect URI to return it to."
// @Failure 400 {object} web.HTTPError "response_type is not 'code', the redirect URI is invalid, the PKCE challenge is missing, or the scope or resource is invalid."
// @Failure 403 {object} models.Message "An API token was used to authorize an OAuth client."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /oauth/authorize [post]
func (srv *Server) HandleAuthorize(c *echo.Context) error {
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

	resp, err := srv.Authorize(&req, u.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// AuthorizePreview describes a valid authorize request for the consent screen.
type AuthorizePreview struct {
	Client      AuthorizeClient `json:"client"`
	RedirectURI string          `json:"redirect_uri" doc:"The validated redirect URI. Only redirect here, also to report a denial."`
	Scopes      []Scope         `json:"scopes" enum:"mcp:read,mcp:write" doc:"The scopes the code would grant, normalized. Empty when the client receives a session."`
	Scoped      bool            `json:"scoped" doc:"Whether the client receives an MCP-only API token instead of a session."`
}

// AuthorizeClient identifies the client asking for authorization.
type AuthorizeClient struct {
	ID string `json:"id"`
}

// Preview validates an authorize request exactly like Authorize, without creating a code.
func (srv *Server) Preview(req *AuthorizeRequest) (*AuthorizePreview, error) {
	scopes, _, err := srv.validate(req)
	if err != nil {
		return nil, err
	}
	if scopes == nil {
		scopes = []Scope{}
	}
	return &AuthorizePreview{
		Client:      AuthorizeClient{ID: req.ClientID},
		RedirectURI: req.RedirectURI,
		Scopes:      scopes,
		Scoped:      len(scopes) > 0,
	}, nil
}

// Authorize validates the OAuth authorization parameters for the given
// authenticated user and creates a single-use authorization code, independent
// of the HTTP layer. Callers own request binding and resolving the user.
func (srv *Server) Authorize(req *AuthorizeRequest, userID int64) (*AuthorizeResponse, error) {
	scopes, resource, err := srv.validate(req)
	if err != nil {
		return nil, err
	}
	scope := formatScope(scopes)

	s := db.NewSession()
	defer s.Close()

	fullUser, err := user.GetUserByID(s, userID)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	code, err := models.CreateOAuthCode(s, &models.OAuthCode{
		UserID:              fullUser.ID,
		ClientID:            req.ClientID,
		RedirectURI:         req.RedirectURI,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		Scope:               scope,
		Resource:            resource,
	})
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
		Scope:       scope,
	}, nil
}

// validate holds every rule of an authorize request; Authorize and Preview must not diverge.
func (srv *Server) validate(req *AuthorizeRequest) (scopes []Scope, resource string, err error) {
	if req.ResponseType != "code" {
		return nil, "", &models.ErrOAuthUnsupportedResponseType{}
	}

	if !ValidateRedirectURI(req.RedirectURI) {
		return nil, "", &models.ErrOAuthInvalidRedirectURI{}
	}

	if req.CodeChallenge == "" || req.CodeChallengeMethod != "S256" {
		return nil, "", &models.ErrOAuthMissingPKCE{}
	}

	return srv.scopedGrant(req)
}

// scopedGrant returns no scopes and an empty resource for a first-party session grant.
func (srv *Server) scopedGrant(req *AuthorizeRequest) (scopes []Scope, resource string, err error) {
	// Generic OAuth libraries send scopes like openid; without an MCP scope they keep getting a session.
	if req.Resource == "" && !requestsMCPScope(req.Scope) {
		return nil, "", nil
	}

	resource = srv.resource.ResourceURL()
	// Also rejects every scoped grant while service.publicurl is unset and the URL is relative.
	if _, ok := parseResource(resource); !ok {
		return nil, "", &models.ErrOAuthInvalidTarget{}
	}
	if req.Resource != "" && !SameResource(req.Resource, resource) {
		return nil, "", &models.ErrOAuthInvalidTarget{}
	}

	scopes = []Scope{ScopeMCPRead, ScopeMCPWrite}
	if req.Scope != "" {
		scopes, err = parseScope(req.Scope)
		if err != nil {
			return nil, "", err
		}
	}

	// The client id names the token the user later revokes.
	if req.ClientID == "" {
		return nil, "", &models.ErrOAuthClientNotFound{}
	}

	return scopes, resource, nil
}
