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

package mcp

import (
	"code.vikunja.io/api/pkg/modules/auth/oauth2server"

	"github.com/modelcontextprotocol/go-sdk/oauthex"
)

// ResourceMetadataPath is the RFC 9728 well-known path of protected resource metadata.
const ResourceMetadataPath = "/.well-known/oauth-protected-resource"

// ResourceMetadataPaths are the paths the metadata is served at: the root, where
// clients probe last, and the MCP endpoint's path inserted after it (RFC 9728 §3.1).
// A subpath deployment also gets the issuer path inserted, for a proxy that forwards
// the host's well-known URLs here unchanged.
func ResourceMetadataPaths() []string {
	paths := []string{
		ResourceMetadataPath,
		ResourceMetadataPath + RoutePrefix,
	}
	if issuerPath := oauth2server.IssuerPath(); issuerPath != "" && issuerPath != "/" {
		paths = append(paths, ResourceMetadataPath+issuerPath+RoutePrefix)
	}
	return paths
}

// ResourceMetadata is the RFC 9728 document of the MCP endpoint; there is none without an issuer.
func ResourceMetadata() (*oauthex.ProtectedResourceMetadata, bool) {
	issuer, ok := oauth2server.Issuer()
	if !ok {
		return nil, false
	}
	scopes := oauth2server.SupportedScopes()
	scopesSupported := make([]string, 0, len(scopes))
	for _, s := range scopes {
		scopesSupported = append(scopesSupported, string(s))
	}
	return &oauthex.ProtectedResourceMetadata{
		Resource:               issuer + RoutePrefix,
		AuthorizationServers:   []string{issuer},
		ScopesSupported:        scopesSupported,
		BearerMethodsSupported: []string{"header"},
		ResourceName:           "Vikunja",
	}, true
}
