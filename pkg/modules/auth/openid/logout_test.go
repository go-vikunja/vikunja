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
	"net/url"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/keyvalue"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newMockOIDCServerWithEndSession serves a discovery document that includes an
// end_session_endpoint, exercising the RP-Initiated Logout discovery path.
func newMockOIDCServerWithEndSession() *httptest.Server {
	var server *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		discovery := map[string]interface{}{
			"issuer":                 server.URL,
			"authorization_endpoint": server.URL + "/auth",
			"token_endpoint":         server.URL + "/token",
			"jwks_uri":               server.URL + "/jwks",
			"end_session_endpoint":   server.URL + "/logout",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(discovery)
	})
	server = httptest.NewServer(mux)
	return server
}

func TestBuildEndSessionURLAssembly(t *testing.T) {
	t.Run("all params", func(t *testing.T) {
		got, err := buildEndSessionURL("https://op.example.com/logout", "my-client", "the-id-token", "https://vikunja.example.com/")
		require.NoError(t, err)

		u, err := url.Parse(got)
		require.NoError(t, err)
		q := u.Query()
		assert.Equal(t, "https", u.Scheme)
		assert.Equal(t, "op.example.com", u.Host)
		assert.Equal(t, "/logout", u.Path)
		assert.Equal(t, "the-id-token", q.Get("id_token_hint"))
		assert.Equal(t, "https://vikunja.example.com/", q.Get("post_logout_redirect_uri"))
		assert.Equal(t, "my-client", q.Get("client_id"))
	})

	t.Run("preserves existing endpoint query params", func(t *testing.T) {
		got, err := buildEndSessionURL("https://op.example.com/logout?foo=bar", "my-client", "the-id-token", "https://vikunja.example.com/")
		require.NoError(t, err)

		u, err := url.Parse(got)
		require.NoError(t, err)
		q := u.Query()
		assert.Equal(t, "bar", q.Get("foo"))
		assert.Equal(t, "the-id-token", q.Get("id_token_hint"))
	})

	t.Run("omits id_token_hint when no token", func(t *testing.T) {
		got, err := buildEndSessionURL("https://op.example.com/logout", "my-client", "", "https://vikunja.example.com/")
		require.NoError(t, err)

		u, err := url.Parse(got)
		require.NoError(t, err)
		q := u.Query()
		assert.False(t, q.Has("id_token_hint"))
		assert.Equal(t, "https://vikunja.example.com/", q.Get("post_logout_redirect_uri"))
		assert.Equal(t, "my-client", q.Get("client_id"))
	})

	t.Run("empty endpoint returns empty", func(t *testing.T) {
		got, err := buildEndSessionURL("", "my-client", "the-id-token", "https://vikunja.example.com/")
		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestBuildEndSessionURLFromDiscovery(t *testing.T) {
	defer CleanupSavedOpenIDProviders()

	server := newMockOIDCServerWithEndSession()
	defer server.Close()

	config.AuthOpenIDEnabled.Set(true)
	config.ServicePublicURL.Set("https://vikunja.example.com/")
	config.AuthOpenIDProviders.Set(map[string]interface{}{
		"provider1": map[string]interface{}{
			"name":         "Provider One",
			"authurl":      server.URL,
			"clientid":     "client1",
			"clientsecret": "secret1",
		},
	})
	_ = keyvalue.Del("openid_providers")
	_ = keyvalue.Del("openid_provider_provider1")

	got, err := BuildEndSessionURL("provider1", &models.SessionOIDCData{
		IDToken:     "raw-id-token",
		ProviderKey: "provider1",
	})
	require.NoError(t, err)

	u, err := url.Parse(got)
	require.NoError(t, err)
	q := u.Query()
	assert.Equal(t, server.URL+"/logout", u.Scheme+"://"+u.Host+u.Path)
	assert.Equal(t, "raw-id-token", q.Get("id_token_hint"))
	assert.Equal(t, "https://vikunja.example.com/", q.Get("post_logout_redirect_uri"))
	assert.Equal(t, "client1", q.Get("client_id"))
}

func TestEndSessionEndpointFallsBackToStaticLogoutURL(t *testing.T) {
	defer CleanupSavedOpenIDProviders()

	// This mock server publishes no end_session_endpoint, so the provider must
	// fall back to the statically configured logouturl.
	server := newMockOIDCServer()
	defer server.Close()

	config.AuthOpenIDEnabled.Set(true)
	config.AuthOpenIDProviders.Set(map[string]interface{}{
		"provider1": map[string]interface{}{
			"name":         "Provider One",
			"authurl":      server.URL,
			"clientid":     "client1",
			"clientsecret": "secret1",
			"logouturl":    "https://op.example.com/static-logout",
		},
	})
	_ = keyvalue.Del("openid_providers")
	_ = keyvalue.Del("openid_provider_provider1")

	provider, err := GetProvider("provider1")
	require.NoError(t, err)
	require.NotNil(t, provider)

	assert.Equal(t, "https://op.example.com/static-logout", provider.EndSessionEndpoint())
}
