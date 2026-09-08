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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCancel(t *testing.T) {
	t.Run("stops the running job but leaves the claim to it", func(t *testing.T) {
		clearMigrationStatus(t)
		u := &user.User{ID: 1}

		status, err := ClaimMigration(&testMigrator{name: "vikunja-file"}, u)
		require.NoError(t, err)

		ctx, done := StartRun(status.ID)
		defer done()

		require.NoError(t, Cancel(u))

		require.Error(t, ctx.Err(), "the running job must have been asked to stop")

		// The job holds the claim until it exits, so a new import must not start yet.
		_, err = ClaimMigration(&testMigrator{name: "vikunja-file"}, u)
		assertIsAlreadyRunning(t, err, "vikunja-file")

		require.NoError(t, FinishMigration(status))

		_, err = ClaimMigration(&testMigrator{name: "vikunja-file"}, u)
		require.NoError(t, err)
	})

	t.Run("refuses to release a claim held by another instance", func(t *testing.T) {
		clearMigrationStatus(t)
		u := &user.User{ID: 1}

		_, err := ClaimMigration(&testMigrator{name: "ticktick"}, u)
		require.NoError(t, err)

		err = Cancel(u)
		var notHere *ErrMigrationNotCancellableHere
		require.ErrorAs(t, err, &notHere)

		_, err = ClaimMigration(&testMigrator{name: "ticktick"}, u)
		assertIsAlreadyRunning(t, err, "ticktick")
	})

	t.Run("releases a legacy row wedging the user", func(t *testing.T) {
		clearMigrationStatus(t)
		u := &user.User{ID: 1}

		s := db.NewSession()
		_, err := s.Insert(&Status{
			UserID:       u.ID,
			MigratorName: "trello",
			StartedAt:    time.Now(),
		})
		require.NoError(t, err)
		require.NoError(t, s.Commit())
		require.NoError(t, s.Close())

		require.NoError(t, Cancel(u))

		_, err = ClaimMigration(&testMigrator{name: "trello"}, u)
		require.NoError(t, err)
	})

	t.Run("reports nothing to cancel", func(t *testing.T) {
		clearMigrationStatus(t)

		err := Cancel(&user.User{ID: 1})
		var noneRunning *ErrNoMigrationRunning
		assert.ErrorAs(t, err, &noneRunning)
	})
}

// A run that got redelivered must not have its cancel func removed by the first run finishing.
func TestStartRunDoesNotClobberAConcurrentRun(t *testing.T) {
	clearMigrationStatus(t)
	u := &user.User{ID: 1}

	status, err := ClaimMigration(&testMigrator{name: "vikunja-file"}, u)
	require.NoError(t, err)

	_, doneFirst := StartRun(status.ID)
	ctxSecond, doneSecond := StartRun(status.ID)
	defer doneSecond()

	doneFirst()

	require.NoError(t, Cancel(u))
	require.Error(t, ctxSecond.Err(), "the second run must still be cancellable")
}
