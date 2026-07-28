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

package websocket

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllowedOriginPatterns(t *testing.T) {
	tests := []struct {
		name        string
		corsOrigins []string
		publicURL   string
		want        []string
	}{
		{
			name:        "keeps configured origins and appends the public url",
			corsOrigins: []string{"http://localhost:*"},
			publicURL:   "https://vikunja.example.com/",
			want:        []string{"http://localhost:*", "https://vikunja.example.com"},
		},
		{
			name:        "drops the allow-all wildcard",
			corsOrigins: []string{"*"},
			publicURL:   "https://vikunja.example.com/",
			want:        []string{"https://vikunja.example.com"},
		},
		{
			name:        "drops wildcard hosts with a scheme",
			corsOrigins: []string{"https://*", "http://*"},
			publicURL:   "https://vikunja.example.com/",
			want:        []string{"https://vikunja.example.com"},
		},
		{
			name:        "keeps wildcard subdomains",
			corsOrigins: []string{"https://*.example.com"},
			publicURL:   "https://vikunja.example.com",
			want:        []string{"https://*.example.com", "https://vikunja.example.com"},
		},
		{
			name:        "strips the path from the public url",
			corsOrigins: nil,
			publicURL:   "https://example.com/vikunja/",
			want:        []string{"https://example.com"},
		},
		{
			name:        "does not duplicate the public url",
			corsOrigins: []string{"https://vikunja.example.com"},
			publicURL:   "https://vikunja.example.com/",
			want:        []string{"https://vikunja.example.com"},
		},
		{
			name:        "no public url configured",
			corsOrigins: []string{"*"},
			publicURL:   "",
			want:        []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, allowedOriginPatterns(tt.corsOrigins, tt.publicURL))
		})
	}
}

// upgradeTestServer serves UpgradeHandler with the given origin configuration.
func upgradeTestServer(t *testing.T, corsOrigins []string, publicURL string) *httptest.Server {
	t.Cleanup(config.InitDefaultConfig)
	config.CorsOrigins.Set(corsOrigins)
	config.ServicePublicURL.Set(publicURL)
	InitHub()

	e := echo.New()
	e.GET("/ws", UpgradeHandler)

	srv := httptest.NewServer(e)
	t.Cleanup(srv.Close)
	return srv
}

// handshake sends a websocket upgrade request and returns the response status code.
// An empty origin omits the Origin header entirely.
func handshake(t *testing.T, srv *httptest.Server, origin string) int {
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/ws", nil)
	require.NoError(t, err)
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", base64.StdEncoding.EncodeToString(make([]byte, 16)))
	if origin != "" {
		req.Header.Set("Origin", origin)
	}

	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = resp.Body.Close()
	})
	return resp.StatusCode
}

func TestUpgradeHandlerOriginVerification(t *testing.T) {
	t.Run("accepts a configured origin", func(t *testing.T) {
		srv := upgradeTestServer(t, []string{"https://frontend.example.com"}, "https://vikunja.example.com")
		assert.Equal(t, http.StatusSwitchingProtocols, handshake(t, srv, "https://frontend.example.com"))
	})

	t.Run("rejects a foreign origin", func(t *testing.T) {
		srv := upgradeTestServer(t, []string{"https://frontend.example.com"}, "https://vikunja.example.com")
		assert.Equal(t, http.StatusForbidden, handshake(t, srv, "https://evil.example.com"))
	})

	t.Run("cors wildcard does not accept a foreign origin", func(t *testing.T) {
		srv := upgradeTestServer(t, []string{"*"}, "https://vikunja.example.com")
		assert.Equal(t, http.StatusForbidden, handshake(t, srv, "https://evil.example.com"))
	})

	t.Run("cors wildcard still accepts the public url", func(t *testing.T) {
		srv := upgradeTestServer(t, []string{"*"}, "https://vikunja.example.com/")
		assert.Equal(t, http.StatusSwitchingProtocols, handshake(t, srv, "https://vikunja.example.com"))
	})

	t.Run("same origin connects without any matching pattern", func(t *testing.T) {
		srv := upgradeTestServer(t, []string{"*"}, "")
		assert.Equal(t, http.StatusSwitchingProtocols, handshake(t, srv, srv.URL))
	})

	t.Run("no origin header connects", func(t *testing.T) {
		srv := upgradeTestServer(t, []string{"https://frontend.example.com"}, "https://vikunja.example.com")
		assert.Equal(t, http.StatusSwitchingProtocols, handshake(t, srv, ""))
	})

	t.Run("cors disabled keeps the public url working", func(t *testing.T) {
		t.Cleanup(config.InitDefaultConfig)
		config.CorsEnable.Set(false)
		srv := upgradeTestServer(t, nil, "https://vikunja.example.com")
		assert.Equal(t, http.StatusSwitchingProtocols, handshake(t, srv, "https://vikunja.example.com"))
		assert.Equal(t, http.StatusForbidden, handshake(t, srv, "https://evil.example.com"))
	})
}
