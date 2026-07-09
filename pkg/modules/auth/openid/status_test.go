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
	"net"
	"net/http"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProvidersStatus(t *testing.T) {
	defer func() {
		config.AuthOpenIDEnabled.Set(false)
		config.AuthOpenIDProviders.Set(nil)
		CleanupSavedOpenIDProviders()
	}()

	t.Run("disabled returns nil", func(t *testing.T) {
		CleanupSavedOpenIDProviders()
		config.AuthOpenIDEnabled.Set(false)

		assert.Nil(t, GetProvidersStatus())
	})

	t.Run("no providers returns nil", func(t *testing.T) {
		CleanupSavedOpenIDProviders()
		config.AuthOpenIDEnabled.Set(true)
		config.AuthOpenIDProviders.Set(nil)

		assert.Nil(t, GetProvidersStatus())
	})

	t.Run("reports provider availability", func(t *testing.T) {
		CleanupSavedOpenIDProviders()
		server := newMockOIDCServer()
		defer server.Close()

		config.AuthOpenIDEnabled.Set(true)
		config.AuthOpenIDProviders.Set(map[string]interface{}{
			"up": map[string]interface{}{
				"name":         "Up Provider",
				"authurl":      server.URL,
				"clientid":     "client1",
				"clientsecret": "secret1",
			},
			"down": map[string]interface{}{
				"name":         "Down Provider",
				"authurl":      "http://127.0.0.1:1",
				"clientid":     "client2",
				"clientsecret": "secret2",
			},
		})

		statuses := GetProvidersStatus()
		require.Len(t, statuses, 2)

		// Results are sorted by provider key.
		assert.Equal(t, "down", statuses[0].Key)
		assert.False(t, statuses[0].Available)
		assert.Equal(t, "up", statuses[1].Key)
		assert.True(t, statuses[1].Available)
	})

	t.Run("yaml style config maps work", func(t *testing.T) {
		CleanupSavedOpenIDProviders()
		server := newMockOIDCServer()
		defer server.Close()

		config.AuthOpenIDEnabled.Set(true)
		config.AuthOpenIDProviders.Set(map[interface{}]interface{}{
			"yaml": map[interface{}]interface{}{
				"name":         "Yaml Provider",
				"authurl":      server.URL,
				"clientid":     "client1",
				"clientsecret": "secret1",
			},
		})

		statuses := GetProvidersStatus()
		require.Len(t, statuses, 1)
		assert.Equal(t, "yaml", statuses[0].Key)
		assert.True(t, statuses[0].Available)
	})
}

func TestInitializeUnavailableProviders(t *testing.T) {
	defer func() {
		config.AuthOpenIDEnabled.Set(false)
		config.AuthOpenIDProviders.Set(nil)
		CleanupSavedOpenIDProviders()
	}()

	CleanupSavedOpenIDProviders()

	// Reserve a port, then release it to simulate a provider that is down
	// while Vikunja starts.
	var lc net.ListenConfig
	listener, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := listener.Addr().String()
	providerURL := "http://" + addr
	require.NoError(t, listener.Close())

	config.AuthOpenIDEnabled.Set(true)
	config.AuthOpenIDProviders.Set(map[string]interface{}{
		"flaky": map[string]interface{}{
			"name":         "Flaky Provider",
			"authurl":      providerURL,
			"clientid":     "client1",
			"clientsecret": "secret1",
		},
	})

	providers, err := GetAllProviders()
	require.NoError(t, err)
	assert.Empty(t, providers, "the provider must fail initialization while its server is down")

	// Retrying while the provider is still down must not make it available.
	initializeUnavailableProviders()
	statuses := GetProvidersStatus()
	require.Len(t, statuses, 1)
	assert.False(t, statuses[0].Available)

	// The provider comes back on the same address.
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"issuer":                 providerURL,
			"authorization_endpoint": providerURL + "/auth",
			"token_endpoint":         providerURL + "/token",
			"jwks_uri":               providerURL + "/jwks",
		})
	})
	for range 20 {
		listener, err = lc.Listen(t.Context(), "tcp", addr)
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	require.NoError(t, err, "could not re-listen on %s", addr)
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: time.Second}
	go func() { _ = srv.Serve(listener) }()
	defer srv.Close()

	initializeUnavailableProviders()

	providers, err = GetAllProviders()
	require.NoError(t, err)
	require.Len(t, providers, 1)
	assert.Equal(t, "flaky", providers[0].Key)

	statuses = GetProvidersStatus()
	require.Len(t, statuses, 1)
	assert.True(t, statuses[0].Available)
}
