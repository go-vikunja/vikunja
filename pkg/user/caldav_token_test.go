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

package user

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"github.com/stretchr/testify/require"
)

func TestDeleteCaldavTokenByID(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	owner := &User{ID: 1}
	token, err := GenerateNewCaldavToken(owner)
	require.NoError(t, err)

	require.NoError(t, DeleteCaldavTokenByID(&User{ID: 2}, token.ID))
	db.AssertExists(t, "user_tokens", map[string]interface{}{
		"id":      token.ID,
		"user_id": owner.ID,
	}, false)

	require.NoError(t, DeleteCaldavTokenByID(owner, token.ID))
	db.AssertMissing(t, "user_tokens", map[string]interface{}{
		"id": token.ID,
	})
}
