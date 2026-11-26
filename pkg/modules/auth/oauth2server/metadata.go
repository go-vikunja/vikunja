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
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/version"
)

// ClientMetadata is what Vikunja keeps of a validated Client ID Metadata Document.
type ClientMetadata struct {
	Name         string
	URI          string
	RedirectURIs []string
}

const (
	// draft-ietf-oauth-client-id-metadata-document §10.11.
	maxMetadataDocumentBytes = 5 << 10
	metadataFetchTimeout     = 5 * time.Second
	// Outlives an authorization code, so the token exchange reuses the document the user consented to.
	minMetadataCacheTTL     = 10 * time.Minute
	defaultMetadataCacheTTL = time.Hour
	maxMetadataCacheTTL     = 24 * time.Hour
	maxClientNameRunes      = 100
	maxClientURILength      = 2048
	// The window matches maxMetadataCacheTTL, so a user also keeps at most about twice this many entries cached.
	maxMetadataFetchesPerUser = 50
	metadataFetchWindow       = 24 * time.Hour
)

// fetchBudget is charged before fetching a document that is not cached.
type fetchBudget func() error

// The token exchange re-resolves a client only after consuming a valid code, which a fetch already paid for.
func unmeteredFetch() error { return nil }

func userFetchBudget(userID int64) fetchBudget {
	return func() error {
		// Reserved before fetching so concurrent requests cannot overshoot.
		fetches, err := keyvalue.IncrByWithTTL("oauth_metadata_fetches_"+strconv.FormatInt(userID, 10), 1, metadataFetchWindow)
		if err != nil {
			return err
		}
		if fetches > maxMetadataFetchesPerUser {
			return &models.ErrOAuthClientMetadataFetchLimit{}
		}
		return nil
	}
}

// MetadataFetcher loads and caches Client ID Metadata Documents.
type MetadataFetcher struct {
	client *http.Client
}

// NewMetadataFetcher fetches with a copy of client that does not follow redirects.
// Production passes utils.NewSSRFSafeHTTPClient().
func NewMetadataFetcher(client *http.Client) *MetadataFetcher {
	c := *client
	c.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &MetadataFetcher{client: &c}
}

func (f *MetadataFetcher) get(ctx context.Context, id metadataClientID, budget fetchBudget) (ClientMetadata, error) {
	key := "oauth_client_metadata_" + utils.Sha256Hex(string(id))
	var cached ClientMetadata
	exists, err := keyvalue.GetWithValue(key, &cached)
	if err != nil {
		return ClientMetadata{}, err
	}
	if exists {
		return cached, nil
	}

	if err := budget(); err != nil {
		return ClientMetadata{}, err
	}
	body, ttl, err := f.fetch(ctx, id)
	if err != nil {
		return ClientMetadata{}, err
	}
	metadata, err := parseClientMetadata(id, body)
	if err != nil {
		return ClientMetadata{}, err
	}
	if err := keyvalue.PutWithTTL(key, metadata, ttl); err != nil {
		log.Errorf("Could not cache OAuth client metadata of %s: %s", id, err)
	}
	return metadata, nil
}

func (f *MetadataFetcher) fetch(ctx context.Context, id metadataClientID) (body []byte, ttl time.Duration, err error) {
	ctx, cancel := context.WithTimeout(ctx, metadataFetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, string(id), nil)
	if err != nil {
		return nil, 0, &models.ErrOAuthClientMetadataUnavailable{}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Vikunja/"+version.Version)

	resp, err := f.client.Do(req)
	if err != nil {
		log.Debugf("Could not fetch OAuth client metadata from %s: %s", id, err)
		return nil, 0, &models.ErrOAuthClientMetadataUnavailable{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Debugf("Fetching OAuth client metadata from %s returned status %d", id, resp.StatusCode)
		return nil, 0, &models.ErrOAuthClientMetadataUnavailable{}
	}
	if !isJSONMediaType(resp.Header.Get("Content-Type")) {
		return nil, 0, &models.ErrOAuthInvalidClientMetadata{Reason: "it is not served as JSON"}
	}

	body, err = io.ReadAll(io.LimitReader(resp.Body, maxMetadataDocumentBytes+1))
	if err != nil {
		log.Debugf("Could not read OAuth client metadata from %s: %s", id, err)
		return nil, 0, &models.ErrOAuthClientMetadataUnavailable{}
	}
	if len(body) > maxMetadataDocumentBytes {
		return nil, 0, &models.ErrOAuthInvalidClientMetadata{Reason: "it is larger than 5 KiB"}
	}

	return body, metadataCacheTTL(resp.Header.Get("Cache-Control")), nil
}

func isJSONMediaType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	return err == nil && (mediaType == "application/json" || strings.HasPrefix(mediaType, "application/") && strings.HasSuffix(mediaType, "+json"))
}

// no-cache and no-store get the minimum too: re-fetching on every request would let a client make Vikunja its traffic source.
func metadataCacheTTL(cacheControl string) time.Duration {
	ttl := defaultMetadataCacheTTL
	for _, directive := range strings.Split(cacheControl, ",") {
		name, value, _ := strings.Cut(strings.TrimSpace(directive), "=")
		switch strings.ToLower(name) {
		case "no-cache", "no-store":
			return minMetadataCacheTTL
		case "max-age":
			seconds, err := strconv.ParseInt(strings.Trim(value, `"`), 10, 64)
			if err == nil && seconds >= 0 {
				ttl = time.Duration(min(seconds, int64(maxMetadataCacheTTL/time.Second))) * time.Second
			}
		}
	}
	return min(max(ttl, minMetadataCacheTTL), maxMetadataCacheTTL)
}

type metadataDocument struct {
	ClientID                          string          `json:"client_id"`
	ClientName                        string          `json:"client_name"`
	ClientURI                         string          `json:"client_uri"`
	RedirectURIs                      []string        `json:"redirect_uris"`
	GrantTypes                        []string        `json:"grant_types"`
	ResponseTypes                     []string        `json:"response_types"`
	TokenEndpointAuthMethod           string          `json:"token_endpoint_auth_method"`
	TokenEndpointAuthMethodsSupported []string        `json:"token_endpoint_auth_methods_supported"`
	ClientSecret                      json.RawMessage `json:"client_secret"`
	ClientSecretExpiresAt             json.RawMessage `json:"client_secret_expires_at"`
	JWKS                              json.RawMessage `json:"jwks"`
	JWKSURI                           json.RawMessage `json:"jwks_uri"`
}

func invalidMetadata(reason string) error {
	return &models.ErrOAuthInvalidClientMetadata{Reason: reason}
}

func parseClientMetadata(id metadataClientID, body []byte) (ClientMetadata, error) {
	var doc metadataDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return ClientMetadata{}, invalidMetadata("it is not a JSON object with the registered metadata types")
	}

	if doc.ClientID != string(id) {
		return ClientMetadata{}, invalidMetadata("its client_id does not match the URL it was fetched from")
	}
	if doc.ClientSecret != nil || doc.ClientSecretExpiresAt != nil {
		return ClientMetadata{}, invalidMetadata("it contains a client secret")
	}
	if doc.JWKS != nil && doc.JWKSURI != nil {
		return ClientMetadata{}, invalidMetadata("it contains both jwks and jwks_uri")
	}
	if !publicClientAuth(doc.TokenEndpointAuthMethod, doc.TokenEndpointAuthMethodsSupported) {
		return ClientMetadata{}, invalidMetadata("it requires client authentication, but only public clients (token endpoint auth method none) are supported")
	}
	if doc.GrantTypes != nil && !slices.Contains(doc.GrantTypes, "authorization_code") {
		return ClientMetadata{}, invalidMetadata("its grant_types do not include authorization_code")
	}
	if doc.ResponseTypes != nil && !slices.Contains(doc.ResponseTypes, "code") {
		return ClientMetadata{}, invalidMetadata("its response_types do not include code")
	}

	name := cleanClientName(doc.ClientName)
	if name == "" {
		return ClientMetadata{}, invalidMetadata("client_name is missing")
	}
	if len(doc.RedirectURIs) == 0 {
		return ClientMetadata{}, invalidMetadata("redirect_uris is missing")
	}

	return ClientMetadata{
		Name:         name,
		URI:          informativeURI(doc.ClientURI),
		RedirectURIs: doc.RedirectURIs,
	}, nil
}

// draft §5: the plural list wins and needs "none"; shared-secret methods must not appear at all.
func publicClientAuth(method string, supported []string) bool {
	for _, m := range append([]string{method}, supported...) {
		if strings.HasPrefix(m, "client_secret_") {
			return false
		}
	}
	if supported != nil {
		return slices.Contains(supported, "none") && (method == "" || slices.Contains(supported, method))
	}
	return method == "" || method == "none"
}

// Drops control and format characters (bidi overrides could disguise the name) and bounds it for display.
func cleanClientName(raw string) string {
	name := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return ' '
		}
		return r
	}, raw)
	name = strings.Join(strings.Fields(name), " ")
	if runes := []rune(name); len(runes) > maxClientNameRunes {
		name = strings.TrimSpace(string(runes[:maxClientNameRunes]))
	}
	return name
}

// An unusable client_uri is dropped, not fatal: it is informative only.
func informativeURI(raw string) string {
	if len(raw) > maxClientURILength || !strings.HasPrefix(raw, "https://") {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Hostname() == "" {
		return ""
	}
	return raw
}
