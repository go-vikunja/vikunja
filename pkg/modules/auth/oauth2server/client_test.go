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

package oauth2server

import (
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestValidateRedirectURI(t *testing.T) {
	t.Run("accepts vikunja-flutter scheme", func(t *testing.T) {
		assert.True(t, ValidateRedirectURI(authorizeRequest{RedirectURI: "vikunja-flutter://callback"}, nil))
	})
	t.Run("accepts vikunja-desktop scheme", func(t *testing.T) {
		assert.True(t, ValidateRedirectURI(authorizeRequest{RedirectURI: "vikunja-desktop://auth"}, nil))
	})
	t.Run("rejects https scheme", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI(authorizeRequest{RedirectURI: "https://evil.com/callback"}, nil))
	})
	t.Run("rejects http scheme", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI(authorizeRequest{RedirectURI: "http://localhost/callback"}, nil))
	})
	t.Run("rejects javascript scheme", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI(authorizeRequest{RedirectURI: "javascript:alert(1)"}, nil))
	})
	t.Run("rejects data scheme", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI(authorizeRequest{RedirectURI: "data:text/html,<script>alert(1)</script>"}, nil))
	})
	t.Run("rejects non-vikunja custom scheme", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI(authorizeRequest{RedirectURI: "myapp://callback"}, nil))
	})
	t.Run("rejects empty URI", func(t *testing.T) {
		assert.False(t, ValidateRedirectURI(authorizeRequest{RedirectURI: ""}, nil))
	})

	t.Run("accepts redirect from client allowed list", func(t *testing.T) {
		client := &models.OAuthClient{
			ClientID:     "test-client",
			ClientName:   "Test App",
			RedirectURIs: "https://myapp.com/callback,https://myapp.com/callback2",
		}
		assert.True(t, ValidateRedirectURI(authorizeRequest{RedirectURI: "https://myapp.com/callback"}, client))
	})

	t.Run("accepts redirect from client allowed list with multiple URIs", func(t *testing.T) {
		client := &models.OAuthClient{
			ClientID:     "test-client",
			ClientName:   "Test App",
			RedirectURIs: "https://myapp.com/callback,https://other.com/auth",
		}
		assert.True(t, ValidateRedirectURI(authorizeRequest{RedirectURI: "https://other.com/auth"}, client))
	})

	t.Run("rejects redirect not in client allowed list", func(t *testing.T) {
		client := &models.OAuthClient{
			ClientID:     "test-client",
			ClientName:   "Test App",
			RedirectURIs: "https://myapp.com/callback",
		}
		assert.False(t, ValidateRedirectURI(authorizeRequest{RedirectURI: "https://evil.com/callback"}, client))
	})

	t.Run("client allowed list takes precedence over scheme check", func(t *testing.T) {
		client := &models.OAuthClient{
			ClientID:     "test-client",
			ClientName:   "Test App",
			RedirectURIs: "http://localhost/callback",
		}
		assert.True(t, ValidateRedirectURI(authorizeRequest{RedirectURI: "http://localhost/callback"}, client))
	})
}
