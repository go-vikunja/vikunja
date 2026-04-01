package oauth2server

import (
	"crypto/rand"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/ulule/limiter/v3"
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
	request := DynamicClientRegistrationResponse{}
	err := c.Bind(&request)
	if err != nil {
		return err
	}

	request.ClientID = rand.Text()
	request.GrantTypes = []string{"authorization_code", "refresh_token"}

	return c.JSON(http.StatusOK, request)
}

func RateLimit() limiter.Rate {

	return limiter.Rate{
		Period: 60 * time.Second,
		Limit:  1,
	}
}
