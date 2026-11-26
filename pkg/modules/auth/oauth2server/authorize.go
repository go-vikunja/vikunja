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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
)

// AuthorizeRequest represents the body for the authorize endpoint.
type AuthorizeRequest struct {
	ResponseType        string `json:"response_type"`
	ClientID            string `json:"client_id" doc:"A first-party client name, or an https URL serving the client's OAuth Client ID Metadata Document. Redirect URIs of a URL client must be listed in its document."`
	RedirectURI         string `json:"redirect_uri"`
	State               string `json:"state"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	Scope               string `json:"scope,omitempty" doc:"Space-separated scopes. With mcp:read or mcp:write the client receives an API token for the MCP endpoint instead of a session. Unknown values are ignored. A client whose client_id is a URL always receives an API token, for mcp:read mcp:write when scope and resource are omitted."`
	Resource            string `json:"resource,omitempty" doc:"RFC 8707 resource indicator. Must be this instance's MCP endpoint; implies a scoped grant of mcp:read mcp:write when scope is omitted."`
}

// AuthorizeResponse is returned on successful authorization code creation.
type AuthorizeResponse struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
	State       string `json:"state"`
	Scope       string `json:"scope,omitempty" doc:"The granted scope. Empty when the client receives a session."`
	Issuer      string `json:"iss,omitempty" doc:"The issuer identifier (RFC 9207). Send it as iss with every redirect to redirect_uri, also with an error. Absent while the instance has no public URL."`
}

// Server issues authorization codes and tokens. Scoped grants are API tokens for resource.
type Server struct {
	resource Resource
	metadata *MetadataFetcher
}

func NewServer(resource Resource, metadata *MetadataFetcher) *Server {
	return &Server{
		resource: resource,
		metadata: metadata,
	}
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

	resp, err := srv.Authorize(c.Request().Context(), &req, u.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// AuthorizePreview describes a valid authorize request for the consent screen.
type AuthorizePreview struct {
	Client           AuthorizeClient `json:"client"`
	RedirectURI      string          `json:"redirect_uri" doc:"The validated redirect URI. Only redirect here, also to report a denial."`
	Scopes           []Scope         `json:"scopes" enum:"mcp:read,mcp:write" doc:"The scopes the code would grant, normalized. Empty when the client receives a session."`
	Scoped           bool            `json:"scoped" doc:"Whether the client receives an MCP-only API token instead of a session."`
	LoopbackRedirect bool            `json:"loopback_redirect" doc:"Whether redirect_uri is a loopback http address. Any program on the user's device can listen there and receive the code."`
	Issuer           string          `json:"iss,omitempty" doc:"The issuer identifier (RFC 9207). Send it as iss with every redirect to redirect_uri, also with an error. Absent while the instance has no public URL."`
}

type ClientKind string

const (
	ClientKindLegacy           ClientKind = "legacy"
	ClientKindMetadataDocument ClientKind = "metadata_document"
)

// AuthorizeClient identifies the client asking for authorization.
type AuthorizeClient struct {
	ID   string     `json:"id"`
	Kind ClientKind `json:"kind" enum:"legacy,metadata_document" doc:"metadata_document: the client_id is an https URL whose metadata document Vikunja fetched and validated. legacy: an unverified first-party client id."`
	Name string     `json:"name,omitempty" doc:"The client_name of a metadata_document client. Self-declared by the client; only host is verified."`
	Host string     `json:"host,omitempty" doc:"The host of a metadata_document client's id URL, the part of its identity the client cannot fake."`
}

// Preview validates an authorize request exactly like Authorize, without creating a code.
func (srv *Server) Preview(ctx context.Context, req *AuthorizeRequest, userID int64) (*AuthorizePreview, error) {
	client, scopes, _, err := srv.validate(ctx, req, userID)
	if err != nil {
		return nil, err
	}
	if scopes == nil {
		scopes = []Scope{}
	}
	issuer, _ := Issuer()
	return &AuthorizePreview{
		Client:           client.describe(),
		RedirectURI:      req.RedirectURI,
		Scopes:           scopes,
		Scoped:           len(scopes) > 0,
		LoopbackRedirect: isLoopbackRedirect(req.RedirectURI),
		Issuer:           issuer,
	}, nil
}

// Authorize validates the OAuth authorization parameters for the given
// authenticated user and creates a single-use authorization code, independent
// of the HTTP layer. Callers own request binding and resolving the user.
func (srv *Server) Authorize(ctx context.Context, req *AuthorizeRequest, userID int64) (*AuthorizeResponse, error) {
	_, scopes, resource, err := srv.validate(ctx, req, userID)
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

	issuer, _ := Issuer()
	return &AuthorizeResponse{
		Code:        code,
		RedirectURI: req.RedirectURI,
		State:       req.State,
		Scope:       scope,
		Issuer:      issuer,
	}, nil
}

// validate holds every rule of an authorize request; Authorize and Preview must not diverge.
func (srv *Server) validate(ctx context.Context, req *AuthorizeRequest, userID int64) (client Client, scopes []Scope, resource string, err error) {
	if req.ResponseType != "code" {
		return nil, nil, "", &models.ErrOAuthUnsupportedResponseType{}
	}

	if req.CodeChallenge == "" || req.CodeChallengeMethod != "S256" {
		return nil, nil, "", &models.ErrOAuthMissingPKCE{}
	}

	// After the cheap checks: resolving may fetch the client's metadata document.
	client, err = srv.resolveClient(ctx, req.ClientID, userFetchBudget(userID))
	if err != nil {
		return nil, nil, "", err
	}

	if !client.allowsRedirect(req.RedirectURI) {
		return nil, nil, "", &models.ErrOAuthInvalidRedirectURI{}
	}

	scopes, resource, err = srv.scopedGrant(req, client)
	if err != nil {
		return nil, nil, "", err
	}
	return client, scopes, resource, nil
}

// Not ok while service.publicurl is unset: the URL is then relative and no scoped grant can be issued.
func (srv *Server) resourceURL() (string, bool) {
	resource := srv.resource.ResourceURL()
	_, ok := parseResource(resource)
	return resource, ok
}

// scopedGrant returns no scopes and an empty resource for a first-party session grant.
func (srv *Server) scopedGrant(req *AuthorizeRequest, client Client) (scopes []Scope, resource string, err error) {
	// Generic OAuth libraries send scopes like openid; without an MCP scope they keep getting a session.
	if req.Resource == "" && !requestsMCPScope(req.Scope) && !client.scopedOnly() {
		return nil, "", nil
	}

	resource, ok := srv.resourceURL()
	if !ok {
		return nil, "", &models.ErrOAuthInvalidTarget{}
	}
	if req.Resource != "" && !SameResource(req.Resource, resource) {
		return nil, "", &models.ErrOAuthInvalidTarget{}
	}

	scopes = supportedScopes()
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
