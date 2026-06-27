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

	"code.vikunja.io/api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHumaInfo covers the public instance-info endpoint. It needs no auth and
// always reports the running version.
func TestHumaInfo(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/info", "", "", "")
	require.Equal(t, http.StatusOK, rec.Code, "body: %s", rec.Body.String())

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Contains(t, body, "version")
	assert.Contains(t, body, "auth")
	assert.Contains(t, body, "available_migrators")

	require.Contains(t, body, "concurrent_writes")
	assert.Equal(t, config.DatabaseType.GetString() != "sqlite", body["concurrent_writes"])
}
