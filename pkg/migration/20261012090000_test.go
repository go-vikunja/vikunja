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

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
)

func TestDeleteOrphanedReactions20261012090000(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	// Full model structs: a partial sync would strip columns from tables other test packages share.
	require.NoError(t, x.Sync2(&user.User{}, &models.Reaction{}))

	keptUser := &user.User{Username: "kept20261012090000"}
	_, err = x.Insert(keptUser)
	require.NoError(t, err)
	goneUser := &user.User{Username: "gone20261012090000"}
	_, err = x.Insert(goneUser)
	require.NoError(t, err)

	kept := &models.Reaction{UserID: keptUser.ID, EntityID: 1, EntityKind: models.ReactionKindTask, Value: "👋"}
	orphan := &models.Reaction{UserID: goneUser.ID, EntityID: 1, EntityKind: models.ReactionKindTask, Value: "👋"}
	_, err = x.Insert(kept, orphan)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := x.In("id", kept.ID, orphan.ID).Delete(&models.Reaction{})
		require.NoError(t, err)
		_, err = x.In("id", keptUser.ID, goneUser.ID).Delete(&user.User{})
		require.NoError(t, err)
	})

	_, err = x.ID(goneUser.ID).Delete(&user.User{})
	require.NoError(t, err)

	require.NoError(t, deleteOrphanedReactions20261012090000(x))

	has, err := x.ID(orphan.ID).Exist(&models.Reaction{})
	require.NoError(t, err)
	require.False(t, has, "orphaned reaction should be gone")

	has, err = x.ID(kept.ID).Exist(&models.Reaction{})
	require.NoError(t, err)
	require.True(t, has, "reaction of an existing user must stay")

	// Idempotent once there is nothing left to clean.
	require.NoError(t, deleteOrphanedReactions20261012090000(x))
}
