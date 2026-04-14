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
	"crypto/rand"
	"net/url"
	"strings"

	"net/http"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"github.com/labstack/echo/v5"
	"github.com/ulule/limiter/v3"
	"xorm.io/xorm"
)

type DynamicClientRegistrationResponse struct {
	ClientID                string            `json:"client_id"`
	ClientSecret            string            `json:"client_secret,omitempty"`
	RegistrationAccessToken string            `json:"registration_access_token,omitempty"`
	RegistrationClientURI   string            `json:"registration_client_uri,omitempty"`
	ClientIDIssuedAt        int64             `json:"client_id_issued_at,omitempty"`
	ClientSecretExpiresAt   int64             `json:"client_secret_expires_at,omitempty"`
	ClientName              string            `json:"client_name,omitempty"`
	ClientNameLocalized     map[string]string `json:"-"`
	RedirectURIs            []string          `json:"redirect_uris,omitempty"`
	GrantTypes              []string          `json:"grant_types,omitempty"`
	TokenEndpointAuthMethod string            `json:"token_endpoint_auth_method,omitempty"`
	LogoURI                 string            `json:"logo_uri,omitempty"`
	JwksURI                 string            `json:"jwks_uri,omitempty"`
}

func RegisterHandler(c *echo.Context) error {

	if !config.AuthEnableDynamicClientRegistration.GetBool() {
		return c.JSON(400, map[string]interface{}{"error": "Dynamic client registration is disabled"})
	}

	request := DynamicClientRegistrationResponse{}
	err := c.Bind(&request)
	if err != nil {
		log.Warningf("error parsing request: %v", err)
		return c.JSON(400, map[string]interface{}{"error": "Error parsing request"})
	}

	if request.ClientName == "" {
		return c.JSON(400, map[string]interface{}{"error": "client_name is required"})
	}

	if len(request.RedirectURIs) == 0 {
		return c.JSON(400, map[string]interface{}{"error": "redirect_uris is required"})
	}

	s := db.NewSession()
	defer func(s *xorm.Session) {
		err := s.Close()
		if err != nil {
			log.Warningf("Failed to close session: %v", err)
		}
	}(s)

	client := models.OAuthClient{
		ClientID:     rand.Text(),
		ClientName:   request.ClientName,
		RedirectURIs: urlEncodeRedirectURIs(request.RedirectURIs),
	}

	err = models.CreateOAuthClient(s, &client)
	if err != nil {
		log.Warningf("Failed to create OAuth client: %v", err)
		return c.JSON(400, map[string]interface{}{"error": "Failed to persist client"})
	}

	request.ClientID = client.ClientID
	request.GrantTypes = []string{"authorization_code", "refresh_token"}

	err = s.Commit()
	if err != nil {
		log.Warningf("Failed to commit session: %v", err)
		return c.JSON(400, map[string]interface{}{"error": "Failed to persist client"})
	}
	return c.JSON(http.StatusOK, request)
}

func RateLimit() limiter.Rate {

	return limiter.Rate{
		Period: 60 * time.Second,
		Limit:  1,
	}
}

func urlEncodeRedirectURIs(uris []string) string {
	encoded := make([]string, len(uris))
	for i, uri := range uris {
		encoded[i] = url.QueryEscape(uri)
	}
	return strings.Join(encoded, ",")
}
