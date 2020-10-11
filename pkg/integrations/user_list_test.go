// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package integrations

import (
	"net/http"
	"testing"

	apiv1 "code.vikunja.io/api/pkg/routes/api/v1"
	"github.com/stretchr/testify/assert"
)

func TestUserList(t *testing.T) {
	t.Run("Normal test", func(t *testing.T) {
		rec, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserList, &testuser1, "", nil, nil)
		assert.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `user1`)
		assert.Contains(t, rec.Body.String(), `user2`)
		assert.Contains(t, rec.Body.String(), `user3`)
		assert.Contains(t, rec.Body.String(), `user4`)
		assert.Contains(t, rec.Body.String(), `user5`)
	})
	t.Run("Search for user3", func(t *testing.T) {
		rec, err := newTestRequestWithUser(t, http.MethodPost, apiv1.UserList, &testuser1, "", map[string][]string{"s": {"user3"}}, nil)
		assert.NoError(t, err)
		assert.Contains(t, rec.Body.String(), `user3`)
		assert.NotContains(t, rec.Body.String(), `user1`)
		assert.NotContains(t, rec.Body.String(), `user2`)
		assert.NotContains(t, rec.Body.String(), `user4`)
		assert.NotContains(t, rec.Body.String(), `user5`)
	})
}
