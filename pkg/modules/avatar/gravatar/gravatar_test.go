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

package gravatar

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	log.InitLogger()
	config.InitDefaultConfig()

	os.Exit(m.Run())
}

func TestGetAvatarSSRF(t *testing.T) {
	// httptest servers listen on 127.0.0.1, so a gravatar base URL pointing at one is exactly the
	// non-routable destination the SSRF-safe client is supposed to refuse.
	setup := func(allowNonRoutable string) (*user.User, func()) {
		keyvalue.InitStorage()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("not-really-an-image"))
		}))

		oldBaseURL := config.AvatarGravatarBaseURL.GetString()
		config.AvatarGravatarBaseURL.Set(server.URL)
		config.OutgoingRequestsAllowNonRoutableIPs.Set(allowNonRoutable)

		return &user.User{ID: 42, Username: "gravatar-ssrf-" + allowNonRoutable, Email: "test@example.com"}, func() {
			server.Close()
			config.AvatarGravatarBaseURL.Set(oldBaseURL)
			config.OutgoingRequestsAllowNonRoutableIPs.Set("false")
		}
	}

	t.Run("refuses a non-routable gravatar host", func(t *testing.T) {
		u, cleanup := setup("false")
		defer cleanup()

		_, _, err := (&Provider{}).GetAvatar(u, 64)
		require.Error(t, err)
	})

	t.Run("allows a non-routable gravatar host when configured", func(t *testing.T) {
		u, cleanup := setup("true")
		defer cleanup()

		content, mimeType, err := (&Provider{}).GetAvatar(u, 64)
		require.NoError(t, err)
		assert.Equal(t, "not-really-an-image", string(content))
		assert.Equal(t, "image/png", mimeType)
	})
}
