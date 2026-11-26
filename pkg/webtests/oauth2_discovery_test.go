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

	"code.vikunja.io/api/pkg/modules/auth/oauth2server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthAuthorizeIssuer(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	for name, tc := range map[string]struct {
		version string
		params  map[string]string
	}{
		"v2 scoped":   {version: "v2", params: map[string]string{"scope": "mcp:read"}},
		"v2 unscoped": {version: "v2"},
		"v1 scoped":   {version: "v1", params: map[string]string{"scope": "mcp:read"}},
	} {
		t.Run(name, func(t *testing.T) {
			res := scopedAuthorize(t, e, tc.version, tc.params)
			require.Equal(t, http.StatusOK, res.code, string(res.body))
			var authorized oauth2server.AuthorizeResponse
			require.NoError(t, json.Unmarshal(res.body, &authorized))
			assert.Equal(t, "https://localhost", authorized.Issuer)
		})
	}

	t.Run("omitted without a public URL", func(t *testing.T) {
		withoutPublicURL(t)
		res := scopedAuthorize(t, e, "v2", nil)
		require.Equal(t, http.StatusOK, res.code, string(res.body))
		assert.NotContains(t, string(res.body), `"iss"`)
	})
}
