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
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"

	"code.vikunja.io/api/pkg/models"
)

// Client is the party asking for authorization: a LegacyClient or a MetadataClient.
type Client interface {
	ID() string
	allowsRedirect(redirectURI string) bool
	// scopedOnly clients never receive a user session.
	scopedOnly() bool
	tokenTitle() string
	describe() AuthorizeClient
}

// LegacyClient is a first-party app naming itself with an unverified plain string.
type LegacyClient struct {
	id string
}

func (c LegacyClient) ID() string { return c.id }

func (c LegacyClient) allowsRedirect(redirectURI string) bool {
	return ValidateRedirectURI(redirectURI)
}

func (LegacyClient) scopedOnly() bool { return false }

func (c LegacyClient) tokenTitle() string { return c.id }

func (c LegacyClient) describe() AuthorizeClient {
	return AuthorizeClient{
		ID:   c.id,
		Kind: ClientKindLegacy,
	}
}

// MetadataClient is identified by an https URL serving its Client ID Metadata Document.
type MetadataClient struct {
	id       metadataClientID
	metadata ClientMetadata
}

func (c MetadataClient) ID() string { return string(c.id) }

func (c MetadataClient) allowsRedirect(redirectURI string) bool {
	if !safeMetadataRedirect(redirectURI) {
		return false
	}
	for _, registered := range c.metadata.RedirectURIs {
		if registered == redirectURI || sameLoopbackRedirect(registered, redirectURI) {
			return true
		}
	}
	return false
}

func (MetadataClient) scopedOnly() bool { return true }

func (c MetadataClient) tokenTitle() string { return c.metadata.Name }

func (c MetadataClient) describe() AuthorizeClient {
	return AuthorizeClient{
		ID:   c.ID(),
		Kind: ClientKindMetadataDocument,
		Name: c.metadata.Name,
		Host: c.id.host(),
		URI:  c.metadata.URI,
	}
}

// clientID is a syntactically valid client_id: a legacyClientID or a metadataClientID.
type clientID interface {
	isClientID()
}

type legacyClientID string

func (legacyClientID) isClientID() {}

// metadataClientID has passed parseClientID's URL rules.
type metadataClientID string

func (metadataClientID) isClientID() {}

func (id metadataClientID) host() string {
	u, _ := url.Parse(string(id))
	return u.Host
}

// Fits oauth_codes.client_id.
const maxClientIDURLLength = 255

var uriScheme = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9+.-]*:`)

func parseClientID(raw string) (clientID, error) {
	if !strings.Contains(raw, "://") && !uriScheme.MatchString(raw) {
		return legacyClientID(raw), nil
	}
	if !validClientIDURL(raw) {
		return nil, &models.ErrOAuthInvalidClientID{}
	}
	return metadataClientID(raw), nil
}

// draft-ietf-oauth-client-id-metadata-document §3. The id is compared as a string, so a
// non-normalized one (uppercase scheme, non-ASCII) is rejected rather than normalized.
func validClientIDURL(raw string) bool {
	if len(raw) > maxClientIDURLLength || !strings.HasPrefix(raw, "https://") || strings.Contains(raw, "#") {
		return false
	}
	for i := 0; i < len(raw); i++ {
		if raw[i] <= ' ' || raw[i] > '~' {
			return false
		}
	}
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Hostname() == "" || u.Path == "" || u.Path == "/" {
		return false
	}
	for _, segment := range strings.Split(u.Path, "/") {
		if segment == "." || segment == ".." {
			return false
		}
	}
	return true
}

func (srv *Server) resolveClient(ctx context.Context, raw string, budget fetchBudget) (Client, error) {
	id, err := parseClientID(raw)
	if err != nil {
		return nil, err
	}
	switch id := id.(type) {
	case legacyClientID:
		return LegacyClient{id: string(id)}, nil
	case metadataClientID:
		// Such a client only ever gets a scoped grant; don't fetch when none can be issued.
		resource, ok := srv.resourceURL()
		if !ok {
			return nil, &models.ErrOAuthInvalidTarget{}
		}
		// draft-ietf-oauth-client-id-metadata-document §10.1: simple string comparison.
		if issuer, _ := Issuer(); string(id) == issuer || string(id) == resource {
			return nil, &models.ErrOAuthClientIDIsServer{}
		}
		metadata, err := srv.metadata.get(ctx, id, budget)
		if err != nil {
			return nil, err
		}
		return MetadataClient{
			id:       id,
			metadata: metadata,
		}, nil
	}
	return nil, fmt.Errorf("unknown client id type %T", id)
}

// ValidateRedirectURI checks that the redirect_uri is either a Vikunja native
// app scheme (e.g. vikunja-flutter://callback) or a loopback http URL as
// recommended by RFC 8252 for native apps that cannot register a custom
// scheme. Any address in 127.0.0.0/8, the IPv6 loopback (::1, in any
// notation), and the literal hostname "localhost" are accepted; dangerous
// schemes like javascript:, data:, https://, or non-loopback http:// targets
// are rejected.
func ValidateRedirectURI(redirectURI string) bool {
	u, err := url.Parse(redirectURI)
	if err != nil || u.Scheme == "" {
		return false
	}

	if strings.HasPrefix(u.Scheme, "vikunja-") {
		return true
	}

	return u.Scheme == "http" && isLoopbackHost(u.Hostname())
}

func isLoopbackRedirect(redirectURI string) bool {
	u, err := url.Parse(redirectURI)
	return err == nil && strings.EqualFold(u.Scheme, "http") && isLoopbackHost(u.Hostname())
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// The frontend navigates to the redirect URI, so a scheme the browser runs itself would execute in Vikunja's origin.
var browserSchemes = map[string]bool{
	"about":       true,
	"blob":        true,
	"data":        true,
	"file":        true,
	"filesystem":  true,
	"ftp":         true,
	"javascript":  true,
	"vbscript":    true,
	"view-source": true,
	"ws":          true,
	"wss":         true,
}

// safeMetadataRedirect allows https, loopback http and private-use schemes (RFC 8252 §7).
func safeMetadataRedirect(redirectURI string) bool {
	u, err := url.Parse(redirectURI)
	if err != nil || u.Scheme == "" || u.User != nil || strings.Contains(redirectURI, "#") {
		return false
	}
	switch scheme := strings.ToLower(u.Scheme); scheme {
	case "https":
		return u.Hostname() != ""
	case "http":
		return isLoopbackHost(u.Hostname())
	default:
		return !browserSchemes[scheme]
	}
}

// sameLoopbackRedirect ignores only the port (RFC 8252 §7.3): native apps get one at runtime.
func sameLoopbackRedirect(registered, requested string) bool {
	regHost, regRest, ok := splitLoopbackRedirect(registered)
	if !ok {
		return false
	}
	reqHost, reqRest, ok := splitLoopbackRedirect(requested)
	return ok && regHost == reqHost && regRest == reqRest
}

func splitLoopbackRedirect(raw string) (host, rest string, ok bool) {
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || !isLoopbackHost(u.Hostname()) {
		return "", "", false
	}
	rest, ok = strings.CutPrefix(raw, "http://"+u.Host)
	return u.Hostname(), rest, ok
}
