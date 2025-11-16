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

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestGetThreadID(t *testing.T) {
	// Save original config value
	originalPublicURL := config.ServicePublicURL.GetString()
	defer func() {
		config.ServicePublicURL.Set(originalPublicURL)
	}()

	t.Run("default domain when no public URL", func(t *testing.T) {
		config.ServicePublicURL.Set("")
		threadID := getThreadID(123)
		assert.Equal(t, "<task-123@vikunja>", threadID)
	})

	t.Run("simple domain without port", func(t *testing.T) {
		config.ServicePublicURL.Set("https://vikunja.example.com")
		threadID := getThreadID(456)
		assert.Equal(t, "<task-456@vikunja.example.com>", threadID)
	})

	t.Run("domain with standard HTTPS port", func(t *testing.T) {
		config.ServicePublicURL.Set("https://vikunja.example.com:443")
		threadID := getThreadID(789)
		// Should strip port to create valid RFC 5322 domain
		assert.Equal(t, "<task-789@vikunja.example.com>", threadID)
	})

	t.Run("domain with non-standard port", func(t *testing.T) {
		config.ServicePublicURL.Set("http://localhost:8080")
		threadID := getThreadID(999)
		// Should strip port to create valid RFC 5322 domain
		assert.Equal(t, "<task-999@localhost>", threadID)
	})

	t.Run("domain with port 3456", func(t *testing.T) {
		config.ServicePublicURL.Set("http://vikunja.local:3456")
		threadID := getThreadID(111)
		// Should strip port to create valid RFC 5322 domain
		assert.Equal(t, "<task-111@vikunja.local>", threadID)
	})

	t.Run("IP address with port", func(t *testing.T) {
		config.ServicePublicURL.Set("http://192.168.1.100:8080")
		threadID := getThreadID(222)
		// Should strip port to create valid RFC 5322 domain
		assert.Equal(t, "<task-222@192.168.1.100>", threadID)
	})

	t.Run("invalid URL falls back to default", func(t *testing.T) {
		config.ServicePublicURL.Set("not a valid url")
		threadID := getThreadID(333)
		assert.Equal(t, "<task-333@vikunja>", threadID)
	})

	t.Run("URL with path", func(t *testing.T) {
		config.ServicePublicURL.Set("https://example.com:9000/vikunja")
		threadID := getThreadID(444)
		// Should use hostname without port
		assert.Equal(t, "<task-444@example.com>", threadID)
	})
}
