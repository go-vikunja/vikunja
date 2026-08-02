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
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookErrorResponseBodyIsTruncated(t *testing.T) {
	const bodyMarker = "A"
	const bodySize = 5 << 20

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(bytes.Repeat([]byte(bodyMarker), bodySize))
	}))
	defer ts.Close()

	// httptest binds to loopback, which the SSRF-safe client blocks by default.
	previousAllowNonRoutable := config.OutgoingRequestsAllowNonRoutableIPs.GetBool()
	config.OutgoingRequestsAllowNonRoutableIPs.Set(true)
	previousClient := webhookClient
	webhookClient = nil

	logDir := t.TempDir()
	log.ConfigureStandardLogger(true, "file", logDir, "ERROR", "text")

	t.Cleanup(func() {
		config.OutgoingRequestsAllowNonRoutableIPs.Set(previousAllowNonRoutable)
		webhookClient = previousClient
		log.InitLogger()
	})

	w := &Webhook{ID: 42, TargetURL: ts.URL}
	err := w.sendWebhookPayload(&WebhookPayload{EventName: "task.updated"})
	require.Error(t, err)

	logged, err := os.ReadFile(filepath.Join(logDir, "standard.log"))
	require.NoError(t, err)

	assert.Contains(t, string(logged), "from webhook 42")
	loggedBodyBytes := strings.Count(string(logged), bodyMarker)
	assert.Positive(t, loggedBodyBytes, "the response body should still be logged for diagnostics")
	assert.LessOrEqual(t, loggedBodyBytes, maxWebhookErrorBodySize, "the response body must not be read and logged unbounded")
}
