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

package saml

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/xml"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"

	petname "github.com/dustinkirkland/golang-petname"
	"github.com/labstack/echo/v5"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"

	"xorm.io/xorm"
)

// Provider represents a SAML identity provider configuration
type Provider struct {
	Name         string                 `json:"name"`
	Key          string                 `json:"key"`
	MetadataURL  string                 `json:"-"`
	MetadataFile string                 `json:"-"`
	IDPMetadata  *saml.EntityDescriptor `json:"-"`
	CertFile     string                 `json:"-"`
	KeyFile      string                 `json:"-"`
	RootURL      string                 `json:"-"`
	Middleware   *samlsp.Middleware     `json:"-"`
}

func init() {
	petname.NonDeterministicMode()
}

// GetAllProviders returns all configured SAML providers
func GetAllProviders() ([]*Provider, error) {
	if !config.AuthSAMLEnabled.GetBool() {
		return nil, nil
	}

	providers := []*Provider{}
	exists, err := keyvalue.GetWithValue("saml_providers", &providers)
	if exists && err == nil {
		return providers, nil
	}

	rawProviders := config.AuthSAMLProviders.Get()
	if rawProviders == nil {
		return nil, nil
	}

	rawProvider, is := rawProviders.(map[string]interface{})
	if !is {
		if rawProviderInterface, ok := rawProviders.(map[interface{}]interface{}); ok {
			rawProvider = make(map[string]interface{}, len(rawProviderInterface))
			for k, v := range rawProviderInterface {
				if key, keyOK := k.(string); keyOK {
					rawProvider[key] = v
				}
			}
		} else {
			log.Criticalf("SAML configuration is in the wrong format. Please check the docs.")
			return nil, nil
		}
	}

	for key, p := range rawProvider {
		pi, is := p.(map[string]interface{})
		if !is {
			if pis, pisOK := p.(map[interface{}]interface{}); pisOK {
				pi = make(map[string]interface{}, len(pis))
				for i, s := range pis {
					if k, keyOK := i.(string); keyOK {
						pi[k] = s
					}
				}
			} else {
				log.Errorf("SAML provider %s has invalid configuration format, skipping", key)
				continue
			}
		}

		provider, err := getProviderFromMap(pi, key)
		if err != nil {
			log.Errorf("Error while getting SAML provider %s: %s", key, err)
			continue
		}
		if provider == nil {
			continue
		}

		providers = append(providers, provider)

		err = keyvalue.Put("saml_provider_"+key, provider)
		if err != nil {
			return nil, err
		}
	}

	err = keyvalue.Put("saml_providers", providers)
	return providers, err
}

// GetProvider retrieves a single SAML provider by key
func GetProvider(key string) (*Provider, error) {
	provider := &Provider{}
	exists, err := keyvalue.GetWithValue("saml_provider_"+key, provider)
	if err != nil {
		return nil, err
	}
	if !exists {
		_, err = GetAllProviders()
		if err != nil {
			return nil, err
		}
		_, err = keyvalue.GetWithValue("saml_provider_"+key, provider)
		if err != nil {
			return nil, err
		}
	}

	if provider.Middleware == nil {
		err = provider.initMiddleware()
		if err != nil {
			return nil, err
		}
	}

	return provider, nil
}

func getProviderFromMap(pi map[string]interface{}, key string) (*Provider, error) {
	optionalKeys := []string{"name", "metadataurl", "metadatafile", "certfile", "keyfile"}

	for _, configKey := range optionalKeys {
		valueFromFile := config.GetConfigValueFromFile("auth.saml.providers." + key + "." + configKey)
		if valueFromFile != "" {
			pi[configKey] = valueFromFile
		}
	}

	if _, exists := pi["name"]; !exists {
		return nil, fmt.Errorf("required key 'name' is missing in the SAML provider configuration")
	}

	hasMetadataURL := func() bool { _, ok := pi["metadataurl"]; return ok }()
	hasMetadataFile := func() bool { _, ok := pi["metadatafile"]; return ok }()
	if !hasMetadataURL && !hasMetadataFile {
		return nil, fmt.Errorf("either 'metadataurl' or 'metadatafile' is required in the SAML provider configuration")
	}

	name, is := pi["name"].(string)
	if !is {
		return nil, nil
	}

	var metadataURL, metadataFile string
	if v, ok := pi["metadataurl"]; ok {
		metadataURL, _ = v.(string)
	}
	if v, ok := pi["metadatafile"]; ok {
		metadataFile, _ = v.(string)
	}

	var certFile, keyFile string
	if v, ok := pi["certfile"]; ok {
		certFile = v.(string)
	}
	if v, ok := pi["keyfile"]; ok {
		keyFile = v.(string)
	}

	provider := &Provider{
		Name:         name,
		Key:          key,
		MetadataURL:  metadataURL,
		MetadataFile: metadataFile,
		CertFile:     certFile,
		KeyFile:      keyFile,
		RootURL:      config.ServicePublicURL.GetString(),
	}

	err := provider.initMiddleware()
	if err != nil {
		log.Errorf("Error initializing SAML provider %s: %s", key, err)
		return provider, nil
	}

	return provider, nil
}

func (p *Provider) initMiddleware() error {
	rootURL, err := url.Parse(p.RootURL)
	if err != nil {
		return fmt.Errorf("invalid root URL: %w", err)
	}

	var keyPair tls.Certificate
	if p.CertFile != "" && p.KeyFile != "" {
		keyPair, err = tls.LoadX509KeyPair(p.CertFile, p.KeyFile)
		if err != nil {
			return fmt.Errorf("error loading SAML certificate: %w", err)
		}
	} else {
		// Generate a self-signed cert for SP metadata
		keyPair, err = generateSelfSignedCert()
		if err != nil {
			return fmt.Errorf("error generating self-signed certificate: %w", err)
		}
	}

	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return fmt.Errorf("error parsing certificate: %w", err)
	}

	var idpMetadata *saml.EntityDescriptor
	if p.MetadataFile != "" {
		xmlData, readErr := os.ReadFile(p.MetadataFile)
		if readErr != nil {
			return fmt.Errorf("error reading IDP metadata file for %s: %w", p.Name, readErr)
		}
		idpMetadata, err = samlsp.ParseMetadata(xmlData)
		if err != nil {
			return fmt.Errorf("error parsing IDP metadata file for %s: %w", p.Name, err)
		}
	} else {
		idpMetadataURL, parseErr := url.Parse(p.MetadataURL)
		if parseErr != nil {
			return fmt.Errorf("invalid metadata URL: %w", parseErr)
		}
		idpMetadata, err = samlsp.FetchMetadata(
			context.Background(),
			http.DefaultClient,
			*idpMetadataURL,
		)
		if err != nil {
			return fmt.Errorf("error fetching IDP metadata for %s: %w", p.Name, err)
		}
	}

	p.IDPMetadata = idpMetadata

	rootURLStr := strings.TrimRight(p.RootURL, "/")
	acsURL := fmt.Sprintf("%s/api/v1/auth/saml/%s/acs", rootURLStr, p.Key)
	metadataURL := fmt.Sprintf("%s/api/v1/auth/saml/%s/metadata", rootURLStr, p.Key)

	samlSP, err := samlsp.New(samlsp.Options{
		EntityID:    metadataURL,
		URL:         *rootURL,
		Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
		Certificate: keyPair.Leaf,
		IDPMetadata: idpMetadata,
	})
	if err != nil {
		return fmt.Errorf("error creating SAML SP for %s: %w", p.Name, err)
	}

	// Override ACS and Metadata URLs to match our actual API routes
	parsedACSURL, err := url.Parse(acsURL)
	if err != nil {
		return fmt.Errorf("error parsing ACS URL: %w", err)
	}
	parsedMetadataURL, err := url.Parse(metadataURL)
	if err != nil {
		return fmt.Errorf("error parsing metadata URL: %w", err)
	}
	samlSP.ServiceProvider.AcsURL = *parsedACSURL
	samlSP.ServiceProvider.MetadataURL = *parsedMetadataURL

	// Allow IDP-initiated SSO so that InResponseTo validation is skipped.
	// This is needed because the SAML ACS endpoint receives a cross-origin POST
	// from the IDP, and browsers won't send the request-tracking cookie back
	// (SameSite=None requires Secure/HTTPS).
	samlSP.ServiceProvider.AllowIDPInitiated = true

	p.Middleware = samlSP
	return nil
}

// HandleMetadata returns the SP metadata XML for a given provider
func HandleMetadata(c *echo.Context) error {
	providerKey := c.Param("provider")
	provider, err := GetProvider(providerKey)
	if err != nil {
		return err
	}
	if provider == nil || provider.Middleware == nil {
		return echo.NewHTTPError(http.StatusNotFound, "SAML provider not found")
	}

	buf, err := xml.MarshalIndent(provider.Middleware.ServiceProvider.Metadata(), "", "  ")
	if err != nil {
		return err
	}

	c.Response().Header().Set("Content-Type", "application/samlmetadata+xml")
	return c.Blob(http.StatusOK, "application/samlmetadata+xml", buf)
}

// HandleACS processes the SAML Assertion Consumer Service callback
// @Summary Authenticate a user with SAML
// @Description After the SAML Identity Provider redirects back with a SAML response, this endpoint processes the assertion and logs the user in.
// @ID get-token-saml
// @tags auth
// @Accept application/x-www-form-urlencoded
// @Produce json
// @Param provider path string true "The SAML provider key"
// @Param SAMLResponse formData string true "The SAML Response"
// @Success 200 {object} auth.Token
// @Failure 500 {object} models.Message "Internal error"
// @Router /auth/saml/{provider}/acs [post]
func HandleACS(c *echo.Context) error {
	providerKey := c.Param("provider")
	provider, err := GetProvider(providerKey)
	if err != nil {
		return err
	}
	if provider == nil || provider.Middleware == nil {
		return echo.NewHTTPError(http.StatusNotFound, "SAML provider not found")
	}

	r := c.Request()
	err = r.ParseForm()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "could not parse form")
	}

	// Retrieve tracked request IDs to validate InResponseTo
	var possibleRequestIDs []string
	if provider.Middleware.ServiceProvider.AllowIDPInitiated {
		possibleRequestIDs = append(possibleRequestIDs, "")
	}
	trackedRequests := provider.Middleware.RequestTracker.GetTrackedRequests(r)
	for _, tr := range trackedRequests {
		possibleRequestIDs = append(possibleRequestIDs, tr.SAMLRequestID)
	}

	frontendURL := strings.TrimRight(config.ServicePublicURL.GetString(), "/")

	assertion, err := provider.Middleware.ServiceProvider.ParseResponse(r, possibleRequestIDs)
	if err != nil {
		if invalidErr, ok := err.(*saml.InvalidResponseError); ok {
			log.Errorf("SAML assertion parse error for provider %s: %v (private: %v)", provider.Name, invalidErr.Response, invalidErr.PrivateErr)
		} else {
			log.Errorf("SAML assertion parse error for provider %s: %v", provider.Name, err)
		}
		return c.Redirect(http.StatusFound, frontendURL+"/auth/saml/"+providerKey+"?error=saml_assertion_failed")
	}

	// Stop tracking the request now that it's been validated
	if relayState := r.FormValue("RelayState"); relayState != "" {
		_ = provider.Middleware.RequestTracker.StopTrackingRequest(
			c.Response(), c.Request(), relayState,
		)
	}

	// We require a subject NameID and treat it as the stable SAML account identifier.
	if assertion.Subject == nil || assertion.Subject.NameID == nil {
		return c.Redirect(http.StatusFound, frontendURL+"/auth/saml/"+providerKey+"?error=no_email")
	}

	email := strings.TrimSpace(assertion.Subject.NameID.Value)

	if !strings.Contains(email, "@") {
		return c.Redirect(http.StatusFound, frontendURL+"/auth/saml/"+providerKey+"?error=no_email")
	}

	name := strings.TrimSpace(getAttributeValue(
		assertion,
		"displayName",
		"urn:oid:2.16.840.1.113730.3.1.241",
		"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
	))
	if name == "" {
		givenName := strings.TrimSpace(getAttributeValue(
			assertion,
			"givenName",
			"urn:oid:2.5.4.42",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
		))
		surname := strings.TrimSpace(getAttributeValue(
			assertion,
			"sn",
			"surname",
			"urn:oid:2.5.4.4",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
		))

		switch {
		case givenName != "" && surname != "":
			name = givenName + " " + surname
		case givenName != "":
			name = givenName
		case surname != "":
			name = surname
		default:
			name = email
		}
	}

	s := db.NewSession()
	defer s.Close()

	u, err := getOrCreateUser(s, email, name, assertion.Issuer.Value, assertion.Subject.NameID.Value)
	if err != nil {
		_ = s.Rollback()
		log.Errorf("Error creating user for SAML provider %s: %v", provider.Name, err)
		return err
	}

	if u.Status == user.StatusDisabled {
		_ = s.Rollback()
		return &user.ErrAccountDisabled{UserID: u.ID}
	}
	if u.Status == user.StatusAccountLocked {
		_ = s.Rollback()
		return &user.ErrAccountLocked{UserID: u.ID}
	}

	err = s.Commit()
	if err != nil {
		_ = s.Rollback()
		return err
	}

	session, err := models.CreateSession(s, u.ID, "SAML Login", "", false)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	token, err := auth.NewUserJWTAuthtoken(u, session.ID)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	return c.Redirect(http.StatusFound, frontendURL+"/auth/saml/"+providerKey+"?token="+token)
}

// HandleLogin initiates SAML authentication by redirecting to the IDP
func HandleLogin(c *echo.Context) error {
	providerKey := c.Param("provider")
	provider, err := GetProvider(providerKey)
	if err != nil {
		return err
	}
	if provider == nil || provider.Middleware == nil {
		return echo.NewHTTPError(http.StatusNotFound, "SAML provider not found")
	}

	binding := saml.HTTPRedirectBinding
	bindingLocation := provider.Middleware.ServiceProvider.GetSSOBindingLocation(binding)
	if bindingLocation == "" {
		binding = saml.HTTPPostBinding
		bindingLocation = provider.Middleware.ServiceProvider.GetSSOBindingLocation(binding)
	}

	authReq, err := provider.Middleware.ServiceProvider.MakeAuthenticationRequest(
		bindingLocation,
		binding,
		saml.HTTPPostBinding,
	)
	if err != nil {
		return err
	}

	// Track the request ID so we can validate InResponseTo in the ACS handler
	relayState, err := provider.Middleware.RequestTracker.TrackRequest(
		c.Response(), c.Request(), authReq.ID,
	)
	if err != nil {
		return err
	}

	redirectURL, err := authReq.Redirect(relayState, &provider.Middleware.ServiceProvider)
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, redirectURL.String())
}

func getAttributeValue(assertion *saml.Assertion, names ...string) string {
	for _, stmt := range assertion.AttributeStatements {
		for _, attr := range stmt.Attributes {
			for _, name := range names {
				if attr.Name == name || attr.FriendlyName == name {
					if len(attr.Values) > 0 {
						return attr.Values[0].Value
					}
				}
			}
		}
	}
	return ""
}

func getOrCreateUser(s *xorm.Session, email, name, issuer, subject string) (u *user.User, err error) {
	u, err = user.GetUserWithEmail(s, &user.User{
		Issuer:  issuer,
		Subject: subject,
	})
	if err == nil {
		if email != u.Email {
			u.Email = email
		}
		if name != u.Name {
			u.Name = name
		}
		u, err = user.UpdateUser(s, u, false)
		return u, err
	}
	if !user.IsErrUserDoesNotExist(err) && !user.IsErrUserStatusError(err) {
		return nil, err
	}
	if user.IsErrUserStatusError(err) {
		return u, nil
	}

	// Link an existing local account when the email already exists.
	u, err = user.GetUserWithEmail(s, &user.User{
		Email:  email,
		Issuer: user.IssuerLocal,
	})
	if err == nil {
		u.Issuer = issuer
		u.Subject = subject
		u, err = user.UpdateUser(s, u, false)
		return u, err
	}
	if !user.IsErrUserDoesNotExist(err) && !user.IsErrUserStatusError(err) {
		return nil, err
	}
	if user.IsErrUserStatusError(err) {
		return u, nil
	}

	uu := &user.User{
		Email:   email,
		Name:    name,
		Status:  user.StatusActive,
		Issuer:  issuer,
		Subject: subject,
	}

	u, err = auth.CreateUserWithRandomUsername(s, uu)
	return u, err
}

// CleanupSavedSAMLProviders removes cached SAML providers
func CleanupSavedSAMLProviders() {
	_ = keyvalue.Del("saml_providers")
}

func generateSelfSignedCert() (tls.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Vikunja SAML SP"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}, nil
}
