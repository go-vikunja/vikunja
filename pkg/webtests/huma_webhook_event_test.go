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
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaWebhookEvents covers the available-webhook-events listing. The route
// is only registered when webhooks are enabled (the test config default).
func TestHumaWebhookEvents(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	t.Run("Returns the events", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/webhooks/events", "", humaTokenFor(t, &testuser1), "")
		require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

		var events []string
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &events))
		assert.ElementsMatch(t, models.GetAvailableWebhookEvents(), events)
	})
	t.Run("Unauthenticated", func(t *testing.T) {
		rec := humaRequest(t, e, http.MethodGet, "/api/v2/webhooks/events", "", "", "")
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "body: %s", rec.Body.String())
	})
}
