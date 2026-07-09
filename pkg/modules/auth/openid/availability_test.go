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
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckProvidersAvailability(t *testing.T) {
	defer func() {
		config.AuthOpenIDEnabled.Set(false)
		config.AuthOpenIDProviders.Set(nil)
		CleanupSavedOpenIDProviders()
	}()

	t.Run("disabled returns nil", func(t *testing.T) {
		invalidateAvailabilityCache()
		config.AuthOpenIDEnabled.Set(false)

		assert.Nil(t, CheckProvidersAvailability(context.Background()))
	})

	t.Run("no providers returns nil", func(t *testing.T) {
		invalidateAvailabilityCache()
		config.AuthOpenIDEnabled.Set(true)
		config.AuthOpenIDProviders.Set(nil)

		assert.Nil(t, CheckProvidersAvailability(context.Background()))
	})

	t.Run("reports reachable and unreachable providers", func(t *testing.T) {
		invalidateAvailabilityCache()
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

		results := CheckProvidersAvailability(context.Background())
		require.Len(t, results, 2)

		// Results are sorted by provider key.
		assert.Equal(t, "down", results[0].Key)
		assert.Equal(t, "Down Provider", results[0].Name)
		assert.False(t, results[0].Reachable)
		assert.Equal(t, "up", results[1].Key)
		assert.Equal(t, "Up Provider", results[1].Name)
		assert.True(t, results[1].Reachable)
	})

	t.Run("yaml style config maps work", func(t *testing.T) {
		invalidateAvailabilityCache()
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

		results := CheckProvidersAvailability(context.Background())
		require.Len(t, results, 1)
		assert.Equal(t, "yaml", results[0].Key)
		assert.True(t, results[0].Reachable)
	})

	t.Run("results are cached", func(t *testing.T) {
		invalidateAvailabilityCache()

		var hits atomic.Int64
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			hits.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		}))
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

		first := CheckProvidersAvailability(context.Background())
		second := CheckProvidersAvailability(context.Background())
		assert.Equal(t, first, second)
		assert.Equal(t, int64(1), hits.Load())
	})
}
