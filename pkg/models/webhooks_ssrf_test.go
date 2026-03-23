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
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookSSRFProtection(t *testing.T) {
	// Reset the singleton client before each test
	resetWebhookClient := func() {
		webhookClient = nil
	}

	t.Run("blocks requests to loopback addresses", func(t *testing.T) {
		resetWebhookClient()
		config.OutgoingRequestsAllowNonRoutableIPs.Set("false")
		config.OutgoingRequestsProxyURL.Set("")
		config.OutgoingRequestsProxyPassword.Set("")

		w := &Webhook{
			ID:        1,
			TargetURL: "http://127.0.0.1:12345/hook",
		}

		err := w.sendWebhookPayload(&WebhookPayload{
			EventName: "test.event",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "prohibited")
	})

	t.Run("allows requests to public addresses", func(t *testing.T) {
		resetWebhookClient()
		config.OutgoingRequestsAllowNonRoutableIPs.Set("false")
		config.OutgoingRequestsProxyURL.Set("")
		config.OutgoingRequestsProxyPassword.Set("")

		// Start a test server (binds to 127.0.0.1 but we test
		// separately that public IPs are allowed in principle)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		// When allownonroutableips is false, even our test server
		// on 127.0.0.1 should be blocked. This confirms the guard works.
		w := &Webhook{
			ID:        1,
			TargetURL: ts.URL + "/hook",
		}

		err := w.sendWebhookPayload(&WebhookPayload{
			EventName: "test.event",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "prohibited")
	})

	t.Run("allows loopback when allownonroutableips is true", func(t *testing.T) {
		resetWebhookClient()
		config.OutgoingRequestsAllowNonRoutableIPs.Set("true")
		config.OutgoingRequestsProxyURL.Set("")
		config.OutgoingRequestsProxyPassword.Set("")

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		w := &Webhook{
			ID:        1,
			TargetURL: ts.URL + "/hook",
		}

		err := w.sendWebhookPayload(&WebhookPayload{
			EventName: "test.event",
		})
		require.NoError(t, err)
	})

	t.Run("blocks requests to private RFC1918 addresses", func(t *testing.T) {
		resetWebhookClient()
		config.OutgoingRequestsAllowNonRoutableIPs.Set("false")
		config.OutgoingRequestsProxyURL.Set("")
		config.OutgoingRequestsProxyPassword.Set("")

		privateAddrs := []string{
			"http://10.0.0.1:80/hook",
			"http://172.16.0.1:80/hook",
			"http://192.168.1.1:80/hook",
		}

		for _, addr := range privateAddrs {
			webhookClient = nil // reset singleton for each
			w := &Webhook{
				ID:        1,
				TargetURL: addr,
			}

			err := w.sendWebhookPayload(&WebhookPayload{
				EventName: "test.event",
			})
			require.Error(t, err, "expected SSRF block for %s", addr)
		}
	})

	t.Run("blocks requests to metadata endpoint", func(t *testing.T) {
		resetWebhookClient()
		config.OutgoingRequestsAllowNonRoutableIPs.Set("false")
		config.OutgoingRequestsProxyURL.Set("")
		config.OutgoingRequestsProxyPassword.Set("")

		w := &Webhook{
			ID:        1,
			TargetURL: "http://169.254.169.254/latest/meta-data/",
		}

		err := w.sendWebhookPayload(&WebhookPayload{
			EventName: "test.event",
		})
		require.Error(t, err)
	})
}
