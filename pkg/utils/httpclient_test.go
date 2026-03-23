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

package utils

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSSRFSafeHTTPClient(t *testing.T) {
	t.Run("returns a non-nil client", func(t *testing.T) {
		client := NewSSRFSafeHTTPClient()
		assert.NotNil(t, client)
	})

	t.Run("can reach a routable test server", func(t *testing.T) {
		config.OutgoingRequestsAllowNonRoutableIPs.Set("true")
		defer config.OutgoingRequestsAllowNonRoutableIPs.Set("false")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewSSRFSafeHTTPClient()
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("blocks non-routable IPs when config is false", func(t *testing.T) {
		config.OutgoingRequestsAllowNonRoutableIPs.Set("false")
		client := NewSSRFSafeHTTPClient()

		// Attempt to connect to localhost (non-routable)
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://127.0.0.1:1/test", nil)
		require.NoError(t, err)
		_, err = client.Do(req) //nolint:bodyclose
		require.Error(t, err)
	})

	t.Run("allows non-routable IPs when config is true", func(t *testing.T) {
		config.OutgoingRequestsAllowNonRoutableIPs.Set("true")
		defer config.OutgoingRequestsAllowNonRoutableIPs.Set("false")

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewSSRFSafeHTTPClient()
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
		require.NoError(t, err)
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
