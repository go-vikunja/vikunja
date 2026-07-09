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

package webtests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/health"
	"code.vikunja.io/api/pkg/modules/auth/openid"
	"code.vikunja.io/api/pkg/routes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthcheck(t *testing.T) {
	t.Run("function", func(t *testing.T) {
		_, err := setupTestEnv()
		require.NoError(t, err)
		require.NoError(t, health.Check())
	})

	t.Run("route", func(t *testing.T) {
		rec, err := newTestRequest(t, http.MethodGet, routes.HealthcheckHandler, ``, nil, nil)
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), "OK")
	})
}

func TestHealthcheckV2OpenIDProviders(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"issuer":                 server.URL,
			"authorization_endpoint": server.URL + "/auth",
			"token_endpoint":         server.URL + "/token",
			"jwks_uri":               server.URL + "/jwks",
		})
	}))
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
	openid.CleanupSavedOpenIDProviders()
	defer func() {
		config.AuthOpenIDEnabled.Set(false)
		config.AuthOpenIDProviders.Set(nil)
		openid.CleanupSavedOpenIDProviders()
	}()

	var body struct {
		Status          string `json:"status"`
		OpenIDProviders []struct {
			Key        string `json:"key"`
			Name       string `json:"name"`
			Registered bool   `json:"registered"`
			Reachable  bool   `json:"reachable"`
		} `json:"openid_providers"`
	}

	// Provider results are probed in the background, so the first request only
	// kicks off the refresh and omits them; poll until they show up.
	require.Eventually(t, func() bool {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/health", "", "", "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		return len(body.OpenIDProviders) > 0
	}, 30*time.Second, 100*time.Millisecond)

	assert.Equal(t, "degraded", body.Status)
	require.Len(t, body.OpenIDProviders, 2)
	assert.Equal(t, "down", body.OpenIDProviders[0].Key)
	assert.False(t, body.OpenIDProviders[0].Registered)
	assert.False(t, body.OpenIDProviders[0].Reachable)
	assert.Equal(t, "up", body.OpenIDProviders[1].Key)
	assert.Equal(t, "Up Provider", body.OpenIDProviders[1].Name)
	assert.True(t, body.OpenIDProviders[1].Registered)
	assert.True(t, body.OpenIDProviders[1].Reachable)
}
