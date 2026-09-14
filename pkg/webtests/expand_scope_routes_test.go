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
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expandScopeRoutes is keyed by echo path template, so a path parameter rename
// silently disables the expansion scope check (GHSA-9rg3-v78m-26q8) on that
// route. Fail loudly when a key no longer names a registered route.
func TestExpandScopeRoutesAreRegistered(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)

	registered := make(map[string]bool)
	for _, r := range e.Router().Routes() {
		if r.Method == http.MethodGet {
			registered[r.Path] = true
		}
	}
	require.NotEmpty(t, registered, "echo router should have registered routes")

	paths := models.ExpandScopeRoutes()
	require.NotEmpty(t, paths)
	for _, path := range paths {
		assert.Truef(t, registered[path],
			"expandScopeRoutes lists %s, which is not a registered GET route", path)
	}
}
