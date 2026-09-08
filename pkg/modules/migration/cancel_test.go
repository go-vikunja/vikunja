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

package migration

import (
	"testing"

	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCancel(t *testing.T) {
	t.Run("stops the running job and frees the slot", func(t *testing.T) {
		clearMigrationStatus(t)
		u := &user.User{ID: 1}

		status, err := ClaimMigration(&testMigrator{name: "vikunja-file"}, u)
		require.NoError(t, err)

		ctx, done := StartRun(t.Context(), status.ID)
		defer done()

		require.NoError(t, Cancel(u))

		require.Error(t, ctx.Err(), "the running job must have been asked to stop")

		// The claim is free again, so the user can start another import.
		_, err = ClaimMigration(&testMigrator{name: "vikunja-file"}, u)
		require.NoError(t, err)
	})

	t.Run("releases a claim whose job is gone", func(t *testing.T) {
		clearMigrationStatus(t)
		u := &user.User{ID: 1}

		_, err := ClaimMigration(&testMigrator{name: "ticktick"}, u)
		require.NoError(t, err)

		require.NoError(t, Cancel(u))

		_, err = ClaimMigration(&testMigrator{name: "ticktick"}, u)
		require.NoError(t, err)
	})

	t.Run("reports nothing to cancel", func(t *testing.T) {
		clearMigrationStatus(t)

		err := Cancel(&user.User{ID: 1})
		var noneRunning *ErrNoMigrationRunning
		assert.ErrorAs(t, err, &noneRunning)
	})
}
