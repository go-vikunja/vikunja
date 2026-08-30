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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
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
	return newMockOIDCServerWithAuthMethods(nil)
}

// A nil slice omits token_endpoint_auth_methods_supported.
func newMockOIDCServerWithAuthMethods(authMethods []string, tokenHandlers ...http.HandlerFunc) *httptest.Server {
	var server *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		discovery := map[string]interface{}{
			"issuer":                 server.URL,
			"authorization_endpoint": server.URL + "/auth",
			"token_endpoint":         server.URL + "/token",
			"jwks_uri":               server.URL + "/jwks",
		}
		if authMethods != nil {
			discovery["token_endpoint_auth_methods_supported"] = authMethods
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(discovery)
	})
	tokenHandler := func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "not implemented", http.StatusNotImplemented)
	}
	if len(tokenHandlers) > 0 {
		tokenHandler = tokenHandlers[0]
	}
	mux.HandleFunc("/token", tokenHandler)
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
	CleanupSavedOpenIDProviders()

	providers, err := GetAllProviders()
	require.Error(t, err)
	assert.Nil(t, providers)
	assert.True(t, IsErrDuplicateOIDCIssuer(err))

	var dupErr *ErrDuplicateOIDCIssuer
	require.ErrorAs(t, err, &dupErr)
	assert.Equal(t, server.URL, dupErr.Issuer)

	// A failed duplicate check must not leave per-provider entries behind:
	// GetProvider resolves them directly, which would keep the duplicate
	// providers usable even though the list build was refused.
	for _, key := range []string{"provider1", "provider2"} {
		exists, err := keyvalue.GetWithValue("openid_provider_"+key, &Provider{})
		require.NoError(t, err)
		assert.False(t, exists, "per-provider entry %s must not be persisted on a duplicate-issuer error", key)
	}
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

func TestGetProviderFromMapStringBooleans(t *testing.T) {
	// Regression test for #2599. When provider config is sourced from environment
	// variables or `*.file` Docker secrets, every leaf value arrives as a string.
	// The boolean fields (emailfallback, usernamefallback, forceuserinfo,
	// requireavailability) must accept stringified bools, not silently fall back
	// to zero values or reject the whole provider.
	defer CleanupSavedOpenIDProviders()

	server := newMockOIDCServer()
	defer server.Close()

	cases := []struct {
		name                    string
		emailFallback           interface{}
		usernameFallback        interface{}
		forceUserInfo           interface{}
		requireAvailability     interface{}
		wantEmailFallback       bool
		wantUsernameFallback    bool
		wantForceUserInfo       bool
		wantRequireAvailability bool
	}{
		{
			name:                    "native bool true",
			emailFallback:           true,
			usernameFallback:        true,
			forceUserInfo:           true,
			requireAvailability:     true,
			wantEmailFallback:       true,
			wantUsernameFallback:    true,
			wantForceUserInfo:       true,
			wantRequireAvailability: true,
		},
		{
			name:                    "native bool false",
			emailFallback:           false,
			usernameFallback:        false,
			forceUserInfo:           false,
			requireAvailability:     false,
			wantEmailFallback:       false,
			wantUsernameFallback:    false,
			wantForceUserInfo:       false,
			wantRequireAvailability: false,
		},
		{
			name:                    "string true",
			emailFallback:           "true",
			usernameFallback:        "true",
			forceUserInfo:           "true",
			requireAvailability:     "true",
			wantEmailFallback:       true,
			wantUsernameFallback:    true,
			wantForceUserInfo:       true,
			wantRequireAvailability: true,
		},
		{
			name:                    "string false",
			emailFallback:           "false",
			usernameFallback:        "false",
			forceUserInfo:           "false",
			requireAvailability:     "false",
			wantEmailFallback:       false,
			wantUsernameFallback:    false,
			wantForceUserInfo:       false,
			wantRequireAvailability: false,
		},
		{
			name:                    "string 1 and 0",
			emailFallback:           "1",
			usernameFallback:        "0",
			forceUserInfo:           "1",
			requireAvailability:     "0",
			wantEmailFallback:       true,
			wantUsernameFallback:    false,
			wantForceUserInfo:       true,
			wantRequireAvailability: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pi := map[string]interface{}{
				"name":                "Test Provider",
				"authurl":             server.URL,
				"clientid":            "client1",
				"clientsecret":        "secret1",
				"emailfallback":       tc.emailFallback,
				"usernamefallback":    tc.usernameFallback,
				"forceuserinfo":       tc.forceUserInfo,
				"requireavailability": tc.requireAvailability,
			}

			provider, err := getProviderFromMap(pi, "test")
			require.NoError(t, err)
			require.NotNil(t, provider)

			assert.Equal(t, tc.wantEmailFallback, provider.EmailFallback, "EmailFallback")
			assert.Equal(t, tc.wantUsernameFallback, provider.UsernameFallback, "UsernameFallback")
			assert.Equal(t, tc.wantForceUserInfo, provider.ForceUserInfo, "ForceUserInfo")
			assert.Equal(t, tc.wantRequireAvailability, provider.RequireAvailability, "RequireAvailability")
		})
	}
}

func TestCleanupRemovesPerProviderEntries(t *testing.T) {
	defer func() {
		config.AuthOpenIDEnabled.Set(false)
		config.AuthOpenIDProviders.Set(nil)
		CleanupSavedOpenIDProviders()
	}()

	config.AuthOpenIDEnabled.Set(true)
	config.AuthOpenIDProviders.Set(map[string]interface{}{
		"stale": map[string]interface{}{
			"name":         "Stale Provider",
			"authurl":      "http://127.0.0.1:1",
			"clientid":     "client1",
			"clientsecret": "secret1",
		},
	})
	require.NoError(t, keyvalue.Put("openid_provider_stale", &Provider{Name: "Stale Provider"}))

	CleanupSavedOpenIDProviders()

	exists, err := keyvalue.GetWithValue("openid_provider_stale", &Provider{})
	require.NoError(t, err)
	assert.False(t, exists, "cleanup must remove per-provider entries, they take precedence over a rebuilt provider list")
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

func TestTokenEndpointAuthStyleFromDiscovery(t *testing.T) {
	cases := []struct {
		name        string
		authMethods []string
		want        oauth2.AuthStyle
	}{
		{
			name:        "basic only",
			authMethods: []string{"client_secret_basic"},
			want:        oauth2.AuthStyleInHeader,
		},
		{
			name:        "post only",
			authMethods: []string{"client_secret_post"},
			want:        oauth2.AuthStyleInParams,
		},
		{
			name:        "both advertised prefers basic",
			authMethods: []string{"client_secret_post", "client_secret_basic"},
			want:        oauth2.AuthStyleInHeader,
		},
		{
			name:        "neither advertised falls back to autodetect",
			authMethods: []string{"private_key_jwt", "none"},
			want:        oauth2.AuthStyleAutoDetect,
		},
		{
			name:        "empty list falls back to autodetect",
			authMethods: []string{},
			want:        oauth2.AuthStyleAutoDetect,
		},
		{
			name:        "field missing falls back to autodetect",
			authMethods: nil,
			want:        oauth2.AuthStyleAutoDetect,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer CleanupSavedOpenIDProviders()

			server := newMockOIDCServerWithAuthMethods(tc.authMethods)
			defer server.Close()

			provider, err := getProviderFromMap(map[string]interface{}{
				"name":         "Test Provider",
				"authurl":      server.URL,
				"clientid":     "client1",
				"clientsecret": "secret1",
			}, "test")
			require.NoError(t, err)
			require.NotNil(t, provider)

			assert.Equal(t, tc.want, provider.Oauth2Config.Endpoint.AuthStyle)
		})
	}

	t.Run("basic only exchanges once with header credentials", func(t *testing.T) {
		requestCount := 0
		var authorization, clientSecret string
		var basicParseFormErr error
		server := newMockOIDCServerWithAuthMethods([]string{authMethodBasic}, func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			authorization = r.Header.Get("Authorization")
			basicParseFormErr = r.ParseForm()
			clientSecret = r.Form.Get("client_secret")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_client","error_description":"distinctive wrong secret"}`))
		})
		defer server.Close()

		provider, err := getProviderFromMap(map[string]interface{}{
			"name":         "Test Provider",
			"authurl":      server.URL,
			"clientid":     "client1",
			"clientsecret": "wrong-secret",
		}, "test")
		require.NoError(t, err)

		_, err = provider.Oauth2Config.Exchange(context.Background(), "authorization-code")
		require.NoError(t, basicParseFormErr)
		require.Error(t, err)
		var retrieveErr *oauth2.RetrieveError
		require.ErrorAs(t, err, &retrieveErr)
		assert.Equal(t, "invalid_client", retrieveErr.ErrorCode)
		assert.Equal(t, "distinctive wrong secret", retrieveErr.ErrorDescription)
		assert.Equal(t, 1, requestCount)
		assert.Equal(t, "Basic Y2xpZW50MTp3cm9uZy1zZWNyZXQ=", authorization)
		assert.Empty(t, clientSecret)
	})

	t.Run("post only exchanges once with form credentials", func(t *testing.T) {
		requestCount := 0
		var authorization, clientID, clientSecret string
		var postParseFormErr error
		server := newMockOIDCServerWithAuthMethods([]string{authMethodPost}, func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			authorization = r.Header.Get("Authorization")
			postParseFormErr = r.ParseForm()
			clientID = r.Form.Get("client_id")
			clientSecret = r.Form.Get("client_secret")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"invalid_client"}`))
		})
		defer server.Close()

		provider, err := getProviderFromMap(map[string]interface{}{
			"name":         "Test Provider",
			"authurl":      server.URL,
			"clientid":     "client1",
			"clientsecret": "secret1",
		}, "test")
		require.NoError(t, err)

		_, err = provider.Oauth2Config.Exchange(context.Background(), "authorization-code")
		require.NoError(t, postParseFormErr)
		require.Error(t, err)
		assert.Equal(t, 1, requestCount)
		assert.Empty(t, authorization)
		assert.Equal(t, "client1", clientID)
		assert.Equal(t, "secret1", clientSecret)
	})
}
