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

package openid

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAllProvidersTypeSafety(t *testing.T) {
	// Clean up any existing providers
	defer func() {
		CleanupSavedOpenIDProviders()
	}()

	t.Run("should handle []interface{} without panic", func(t *testing.T) {
		// Setup config with OpenID enabled
		config.AuthOpenIDEnabled.Set(true)

		// Mock the config value to be []interface{} which causes the original panic
		configValue := []interface{}{
			map[string]interface{}{
				"name":         "test-provider",
				"authurl":      "https://example.com/auth",
				"clientid":     "test-client",
				"clientsecret": "test-secret",
			},
		}
		config.AuthOpenIDProviders.Set(configValue)

		// Clear keyvalue cache to force reading from config
		_ = keyvalue.Del("openid_providers")

		// This should not panic, but should handle gracefully and return empty
		providers, err := GetAllProviders()

		// Should return empty providers since the config format is invalid
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(providers) != 0 {
			t.Errorf("Expected empty providers list, got: %d", len(providers))
		}
	})

	t.Run("should handle other invalid types without panic", func(t *testing.T) {
		// Setup config with OpenID enabled
		config.AuthOpenIDEnabled.Set(true)

		// Mock the config value to be a string (another invalid type)
		configValue := "invalid-config"
		config.AuthOpenIDProviders.Set(configValue)

		// Clear keyvalue cache to force reading from config
		_ = keyvalue.Del("openid_providers")

		// This should not panic
		providers, err := GetAllProviders()

		// Should return empty providers since the config format is invalid
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if len(providers) != 0 {
			t.Errorf("Expected empty providers list, got: %d", len(providers))
		}
	})
}

// newMockOIDCServer creates a test HTTP server that serves a valid OIDC discovery document.
// The issuer in the discovery document matches the server's URL.
func newMockOIDCServer() *httptest.Server {
	var server *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		discovery := map[string]interface{}{
			"issuer":                 server.URL,
			"authorization_endpoint": server.URL + "/auth",
			"token_endpoint":         server.URL + "/token",
			"jwks_uri":               server.URL + "/jwks",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(discovery)
	})
	server = httptest.NewServer(mux)
	return server
}

func TestDuplicateIssuersDetected(t *testing.T) {
	defer CleanupSavedOpenIDProviders()

	// Create a single mock server — both providers will use the same issuer
	server := newMockOIDCServer()
	defer server.Close()

	config.AuthOpenIDEnabled.Set(true)
	config.AuthOpenIDProviders.Set(map[string]interface{}{
		"provider1": map[string]interface{}{
			"name":         "Provider One",
			"authurl":      server.URL,
			"clientid":     "client1",
			"clientsecret": "secret1",
		},
		"provider2": map[string]interface{}{
			"name":         "Provider Two",
			"authurl":      server.URL,
			"clientid":     "client2",
			"clientsecret": "secret2",
		},
	})
	_ = keyvalue.Del("openid_providers")

	providers, err := GetAllProviders()
	require.Error(t, err)
	assert.Nil(t, providers)
	assert.True(t, IsErrDuplicateOIDCIssuer(err))

	var dupErr *ErrDuplicateOIDCIssuer
	require.ErrorAs(t, err, &dupErr)
	assert.Equal(t, server.URL, dupErr.Issuer)
}

func TestUniqueIssuersAllowed(t *testing.T) {
	defer CleanupSavedOpenIDProviders()

	// Create two separate mock servers — different issuers
	server1 := newMockOIDCServer()
	defer server1.Close()
	server2 := newMockOIDCServer()
	defer server2.Close()

	config.AuthOpenIDEnabled.Set(true)
	config.AuthOpenIDProviders.Set(map[string]interface{}{
		"provider1": map[string]interface{}{
			"name":         "Provider One",
			"authurl":      server1.URL,
			"clientid":     "client1",
			"clientsecret": "secret1",
		},
		"provider2": map[string]interface{}{
			"name":         "Provider Two",
			"authurl":      server2.URL,
			"clientid":     "client2",
			"clientsecret": "secret2",
		},
	})
	_ = keyvalue.Del("openid_providers")

	providers, err := GetAllProviders()
	require.NoError(t, err)
	assert.Len(t, providers, 2)
}

func TestFailedDiscoverySkippedInIssuerCheck(t *testing.T) {
	defer CleanupSavedOpenIDProviders()

	// One valid server, one unreachable
	server := newMockOIDCServer()
	defer server.Close()

	config.AuthOpenIDEnabled.Set(true)
	config.AuthOpenIDProviders.Set(map[string]interface{}{
		"valid": map[string]interface{}{
			"name":         "Valid Provider",
			"authurl":      server.URL,
			"clientid":     "client1",
			"clientsecret": "secret1",
		},
		"broken": map[string]interface{}{
			"name":         "Broken Provider",
			"authurl":      "http://127.0.0.1:1",
			"clientid":     "client2",
			"clientsecret": "secret2",
		},
	})
	_ = keyvalue.Del("openid_providers")

	// The broken provider will fail discovery and be skipped.
	// The valid provider should load successfully.
	providers, err := GetAllProviders()
	require.NoError(t, err)
	assert.Len(t, providers, 1)
	assert.Equal(t, "Valid Provider", providers[0].Name)
}
