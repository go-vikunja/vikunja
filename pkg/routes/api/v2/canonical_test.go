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

package apiv2

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCanonicalAPIIncludesFeatureGatedRoutes(t *testing.T) {
	api, err := NewCanonicalAPI()
	require.NoError(t, err)

	paths := api.OpenAPI().Paths
	for _, path := range []string{
		"/auth/openid/{provider}/callback",
		"/backgrounds/unsplash/search",
		"/login",
		"/migration/microsoft-todo/auth",
		"/migration/todoist/auth",
		"/migration/trello/auth",
		"/projects/{project}/backgrounds/upload",
		"/projects/{project}/shares",
		"/projects/{project}/webhooks",
		"/register",
		"/shares/{share}/auth",
		"/tasks/{task}/attachments",
		"/tasks/{task}/comments",
		"/user/settings/totp",
		"/user/settings/webhooks",
		"/webhooks/events",
	} {
		assert.Contains(t, paths, path)
	}
	for path := range paths {
		assert.False(t, strings.HasPrefix(path, "/test/"), "canonical API includes %s", path)
	}
	require.Len(t, api.OpenAPI().Servers, 1)
	assert.Equal(t, GroupPrefix, api.OpenAPI().Servers[0].URL)
}
