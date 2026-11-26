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
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRedirectURI(t *testing.T) {
	t.Run("accepts vikunja-flutter scheme", func(t *testing.T) {
		assert.True(t, ValidateRedirectURI("vikunja-flutter://callback"))
	})
	t.Run("accepts vikunja-desktop scheme", func(t *testing.T) {
		assert.True(t, ValidateRedirectURI("vikunja-desktop://auth"))
	})
	t.Run("accepts http localhost", func(t *testing.T) {
		assert.True(t, ValidateRedirectURI("http://localhost/callback"))
	})
	t.Run("accepts http localhost with port", func(t *testing.T) {
		assert.True(t, ValidateRedirectURI("http://localhost:8080/callback"))
	})
	t.Run("accepts http 127.0.0.1", func(t *testing.T) {
		assert.True(t, ValidateRedirectURI("http://127.0.0.1:8080/callback"))
	})
	t.Run("accepts other 127.0.0.0/8 loopback addresses", func(t *testing.T) {
		assert.True(t, ValidateRedirectURI("http://127.0.0.2:1234/callback"))
	})
	t.Run("accepts http ipv6 loopback", func(t *testing.T) {
		assert.True(t, ValidateRedirectURI("http://[::1]:8080/callback"))
	})
	t.Run("accepts expanded ipv6 loopback", func(t *testing.T) {
		assert.True(t, ValidateRedirectURI("http://[0:0:0:0:0:0:0:1]:1234/callback"))
	})
	t.Run("accepts localhost case-insensitively", func(t *testing.T) {
		assert.True(t, ValidateRedirectURI("http://LocalHost:8080/callback"))
	})
	t.Run("rejects 0.0.0.0", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI("http://0.0.0.0:8080/callback"))
	})
	t.Run("rejects https scheme", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI("https://evil.com/callback"))
	})
	t.Run("rejects https localhost", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI("https://localhost/callback"))
	})
	t.Run("rejects non-loopback http", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI("http://evil.com/callback"))
	})
	t.Run("rejects http with localhost in path", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI("http://evil.com/localhost"))
	})
	t.Run("rejects http with localhost subdomain trick", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI("http://localhost.evil.com/callback"))
	})
	t.Run("rejects javascript scheme", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI("javascript:alert(1)"))
	})
	t.Run("rejects data scheme", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI("data:text/html,<script>alert(1)</script>"))
	})
	t.Run("rejects non-vikunja custom scheme", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI("myapp://callback"))
	})
	t.Run("rejects empty URI", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI(""))
	})
}

func TestParseClientID(t *testing.T) {
	t.Run("plain strings are legacy clients", func(t *testing.T) {
		for _, raw := range []string{
			"vikunja",
			"vikunja-desktop",
			"veans-cli",
			"test-mcp-client",
			"",
		} {
			id, err := parseClientID(raw)
			require.NoError(t, err, raw)
			assert.Equal(t, legacyClientID(raw), id, raw)
		}
	})

	t.Run("https URLs are metadata document clients", func(t *testing.T) {
		for _, raw := range []string{
			"https://app.example.com/oauth/client-metadata.json",
			"https://app.example.com:8443/client",
			"https://app.example.com/client?version=2",
			"https://127.0.0.1/client.json",
			"https://app.example.com/" + strings.Repeat("a", maxClientIDURLLength-len("https://app.example.com/")),
		} {
			id, err := parseClientID(raw)
			require.NoError(t, err, raw)
			assert.Equal(t, metadataClientID(raw), id, raw)
		}
	})

	t.Run("rejects URL-shaped ids that are not valid metadata document URLs", func(t *testing.T) {
		for _, raw := range []string{
			"http://app.example.com/client.json",
			"HTTPS://app.example.com/client.json",
			"ftp://app.example.com/client.json",
			"vikunja-desktop://callback",
			"urn:example:client",
			"mailto:client@example.com",
			"https://app.example.com",
			"https://app.example.com/",
			"https:///client.json",
			"https://user@app.example.com/client.json",
			"https://user:secret@app.example.com/client.json",
			"https://app.example.com/client.json#frag",
			"https://app.example.com/client.json#",
			"https://app.example.com/./client.json",
			"https://app.example.com/oauth/../client.json",
			"https://app.example.com/oauth/%2e%2e/client.json",
			"https://app.example.com/oauth/..",
			"https://app.example.com/client json",
			"https://app.exämple.com/client.json",
			"https://app.example.com/" + strings.Repeat("a", maxClientIDURLLength),
		} {
			_, err := parseClientID(raw)
			assert.True(t, models.IsErrOAuthInvalidClientID(err), "%s: %v", raw, err)
		}
	})
}

func TestMetadataClientAllowsRedirect(t *testing.T) {
	client := MetadataClient{
		id: "https://app.example.com/client.json",
		metadata: ClientMetadata{
			Name: "App",
			RedirectURIs: []string{
				"https://app.example.com/callback",
				"http://127.0.0.1:3000/callback",
				"http://localhost/cb",
				"http://[::1]:1234/cb",
				"cursor://anysphere.cursor-retrieval/oauth/callback",
				"javascript:alert(1)//",
				"data:text/html,hi",
				"http://evil.example/cb",
			},
		},
	}

	for _, uri := range []string{
		"https://app.example.com/callback",
		"http://127.0.0.1:3000/callback",
		"http://127.0.0.1:54321/callback",
		"http://127.0.0.1/callback",
		"http://localhost/cb",
		"http://localhost:8080/cb",
		"http://[::1]:9/cb",
		"cursor://anysphere.cursor-retrieval/oauth/callback",
	} {
		assert.True(t, client.allowsRedirect(uri), uri)
	}

	for _, uri := range []string{
		"https://app.example.com/callback2",
		"https://app.example.com/callback/",
		"https://app.example.com/Callback",
		"https://app.example.com/callback?x=1",
		"https://app.example.com/callback#x",
		"https://app.example.com:443/callback",
		"https://APP.example.com/callback",
		"http://app.example.com/callback",
		"https://user@app.example.com/callback",
		"https://evil.example.com/callback",
		"https://app.example.com.evil.example/callback",
		"https://127.0.0.1:3000/callback",
		"http://127.0.0.1:3000/callback2",
		"http://127.0.0.1:3000/callback/",
		"http://127.0.0.1:3000/callback?x=1",
		"http://127.0.0.2:3000/callback",
		"http://localhost:3000/callback",
		"http://user@127.0.0.1:3000/callback",
		"http://127.0.0.1:3000@evil.example/callback",
		"http://localhost/cb2",
		"http://[::1]:9/cb/",
		"cursor://anysphere.cursor-retrieval/oauth/callback2",
		"javascript:alert(1)//",
		"data:text/html,hi",
		"http://evil.example/cb",
		"",
	} {
		assert.False(t, client.allowsRedirect(uri), uri)
	}
}

func TestLegacyClientKeepsRedirectRules(t *testing.T) {
	client := LegacyClient{id: "vikunja"}
	assert.True(t, client.allowsRedirect("vikunja-flutter://callback"))
	assert.True(t, client.allowsRedirect("http://127.0.0.1:1/callback"))
	assert.False(t, client.allowsRedirect("https://app.example.com/callback"))
	assert.False(t, client.allowsRedirect("cursor://anysphere.cursor-retrieval/oauth/callback"))
}

func TestIsLoopbackRedirect(t *testing.T) {
	for _, uri := range []string{
		"http://127.0.0.1:3000/callback",
		"http://127.1.2.3/callback",
		"http://localhost:8080/callback",
		"HTTP://LOCALHOST/callback",
		"http://[::1]:3000/callback",
		"http://[0:0:0:0:0:0:0:1]/callback",
		"http://[::ffff:127.0.0.1]:3000/callback",
	} {
		assert.True(t, isLoopbackRedirect(uri), uri)
	}
	for _, uri := range []string{
		"https://127.0.0.1/callback",
		"https://app.example.com/callback",
		"http://app.example.com/callback",
		"http://127.0.0.1.example.com/callback",
		"vikunja-flutter://callback",
		"cursor://localhost/callback",
		"::not a url",
	} {
		assert.False(t, isLoopbackRedirect(uri), uri)
	}
}
