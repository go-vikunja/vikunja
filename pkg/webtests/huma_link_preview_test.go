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
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHumaLinkPreview(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	token := humaTokenFor(t, &testuser1)
	request := func(target, auth string) int {
		return humaRequest(t, e, http.MethodGet, "/api/v2/link-preview?url="+url.QueryEscape(target), "", auth, "").Code
	}

	t.Run("requires authentication", func(t *testing.T) {
		assert.Equal(t, http.StatusUnauthorized, request("https://example.com", ""))
	})
	t.Run("rejects invalid URLs", func(t *testing.T) {
		for _, target := range []string{"", "ftp://example.com", "relative/path", "https://user:secret@example.com"} {
			assert.Equal(t, http.StatusUnprocessableEntity, request(target, token), target)
		}
	})
	t.Run("rejects this instance", func(t *testing.T) {
		previous := config.ServicePublicURL.GetString()
		config.ServicePublicURL.Set("https://tasks.example.com/")
		t.Cleanup(func() { config.ServicePublicURL.Set(previous) })
		assert.Equal(t, http.StatusUnprocessableEntity, request("https://tasks.example.com/tasks/1", token))
	})
	t.Run("blocks private destinations", func(t *testing.T) {
		previous := config.OutgoingRequestsAllowNonRoutableIPs.GetBool()
		config.OutgoingRequestsAllowNonRoutableIPs.Set(false)
		t.Cleanup(func() { config.OutgoingRequestsAllowNonRoutableIPs.Set(previous) })
		assert.Equal(t, http.StatusBadGateway, request("http://127.0.0.1:1/", token))
	})
	t.Run("shares cached metadata across authenticated users", func(t *testing.T) {
		var requests atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			requests.Add(1)
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><head><meta property="og:title" content="Example title"></head></html>`))
		}))
		defer server.Close()
		previous := config.OutgoingRequestsAllowNonRoutableIPs.GetBool()
		config.OutgoingRequestsAllowNonRoutableIPs.Set(true)
		t.Cleanup(func() { config.OutgoingRequestsAllowNonRoutableIPs.Set(previous) })
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/link-preview?url="+url.QueryEscape(server.URL), "", token, "")
		assert.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		assert.Contains(t, rec.Body.String(), `"title":"Example title"`)
		assert.Equal(t, http.StatusOK, request(server.URL, humaTokenFor(t, &testuser2)))
		assert.Equal(t, http.StatusUnauthorized, request(server.URL, ""))
		assert.EqualValues(t, 1, requests.Load())
	})
}
