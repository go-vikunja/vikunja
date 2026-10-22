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

	"code.vikunja.io/api/pkg/config"
	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserShow(t *testing.T) {
	t.Run("Normal test", func(t *testing.T) {
		rec, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserShow, &testuser1, "", nil, nil)
		require.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `"id":1`)
		assert.Contains(t, rec.Body.String(), `"username":"user1"`)
		assert.NotContains(t, rec.Body.String(), `"email":""`)
	})
}

func TestUserShowPendingEmail(t *testing.T) {
	// One env for the whole test: setupTestEnv reloads the fixtures, which would
	// undo the pending email change between the two requests.
	e, err := setupTestEnv()
	require.NoError(t, err)

	config.MailerEnabled.Set(true)
	defer config.MailerEnabled.Set(false)

	show := func() string {
		c, rec := createRequest(e, http.MethodGet, "", nil, nil)
		addUserTokenToContext(t, &testuser1, c)
		require.NoError(t, apiv1.UserShow(c))
		return rec.Body.String()
	}

	assert.NotContains(t, show(), "pending_email")

	c, _ := createRequest(e, http.MethodPost, `{"new_email":"pending@example.com","password":"12345678"}`, nil, nil)
	addUserTokenToContext(t, &testuser1, c)
	require.NoError(t, apiv1.UpdateUserEmail(c))

	assert.Contains(t, show(), `"pending_email":"pending@example.com"`)
}
