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
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetAvailabilityTestState() {
	invalidateAvailabilityCache()
	CleanupSavedOpenIDProviders()
}

func TestProbeProvidersAvailability(t *testing.T) {
	defer func() {
		config.AuthOpenIDEnabled.Set(false)
		config.AuthOpenIDProviders.Set(nil)
		resetAvailabilityTestState()
	}()

	t.Run("disabled returns nil", func(t *testing.T) {
		resetAvailabilityTestState()
		config.AuthOpenIDEnabled.Set(false)

		assert.Nil(t, ProbeProvidersAvailability(context.Background()))
	})

	t.Run("no providers returns nil", func(t *testing.T) {
		resetAvailabilityTestState()
		config.AuthOpenIDEnabled.Set(true)
		config.AuthOpenIDProviders.Set(nil)

		assert.Nil(t, ProbeProvidersAvailability(context.Background()))
	})

	t.Run("reports registration and reachability", func(t *testing.T) {
		resetAvailabilityTestState()
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

		results := ProbeProvidersAvailability(context.Background())
		require.Len(t, results, 2)

		// Results are sorted by provider key.
		assert.Equal(t, "down", results[0].Key)
		assert.Equal(t, "Down Provider", results[0].Name)
		assert.False(t, results[0].Registered)
		assert.False(t, results[0].Reachable)
		assert.Equal(t, "up", results[1].Key)
		assert.Equal(t, "Up Provider", results[1].Name)
		assert.True(t, results[1].Registered)
		assert.True(t, results[1].Reachable)
	})

	t.Run("yaml style config maps work", func(t *testing.T) {
		resetAvailabilityTestState()
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

		results := ProbeProvidersAvailability(context.Background())
		require.Len(t, results, 1)
		assert.Equal(t, "yaml", results[0].Key)
		assert.True(t, results[0].Registered)
		assert.True(t, results[0].Reachable)
	})
}

func TestCheckProvidersAvailability(t *testing.T) {
	defer func() {
		config.AuthOpenIDEnabled.Set(false)
		config.AuthOpenIDProviders.Set(nil)
		resetAvailabilityTestState()
	}()

	resetAvailabilityTestState()

	var hits atomic.Int64
	var server *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"issuer":                 server.URL,
			"authorization_endpoint": server.URL + "/auth",
			"token_endpoint":         server.URL + "/token",
			"jwks_uri":               server.URL + "/jwks",
		})
	})
	server = httptest.NewServer(mux)
	defer server.Close()

	config.AuthOpenIDEnabled.Set(true)
	config.AuthOpenIDProviders.Set(map[string]interface{}{
		"cached": map[string]interface{}{
			"name":         "Cached Provider",
			"authurl":      server.URL,
			"clientid":     "client1",
			"clientsecret": "secret1",
		},
	})

	// The first call cannot serve anything yet, it only kicks off the
	// background refresh.
	assert.Nil(t, CheckProvidersAvailability(context.Background()))

	require.Eventually(t, func() bool {
		return CheckProvidersAvailability(context.Background()) != nil
	}, 10*time.Second, 50*time.Millisecond)

	results := CheckProvidersAvailability(context.Background())
	require.Len(t, results, 1)
	assert.Equal(t, "cached", results[0].Key)
	assert.True(t, results[0].Registered)
	assert.True(t, results[0].Reachable)

	// Further calls within the cache TTL serve the cached results without
	// hitting the provider again.
	probed := hits.Load()
	for i := 0; i < 3; i++ {
		require.Len(t, CheckProvidersAvailability(context.Background()), 1)
	}
	assert.Equal(t, probed, hits.Load())
}

func TestRegisterMissingProviders(t *testing.T) {
	defer func() {
		config.AuthOpenIDEnabled.Set(false)
		config.AuthOpenIDProviders.Set(nil)
		resetAvailabilityTestState()
	}()

	resetAvailabilityTestState()

	// Reserve a port, then release it to simulate a provider that is down
	// while Vikunja starts.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
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
	assert.Empty(t, providers, "the provider must fail registration while its server is down")

	// Re-registration while the provider is still down must not register it.
	registerMissingProviders()
	providers, err = GetAllProviders()
	require.NoError(t, err)
	assert.Empty(t, providers)

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
		listener, err = net.Listen("tcp", addr)
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	require.NoError(t, err, "could not re-listen on %s", addr)
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(listener) }()
	defer srv.Close()

	registerMissingProviders()

	providers, err = GetAllProviders()
	require.NoError(t, err)
	require.Len(t, providers, 1)
	assert.Equal(t, "flaky", providers[0].Key)

	results := ProbeProvidersAvailability(context.Background())
	require.Len(t, results, 1)
	assert.True(t, results[0].Registered)
	assert.True(t, results[0].Reachable)
}
