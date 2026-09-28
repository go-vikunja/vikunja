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

	"github.com/stretchr/testify/require"
)

func TestDeleteOrphanedLabelTasks20260928140648(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)
	// The full model structs, not a partial one: with VIKUNJA_TESTS_USE_CONFIG
	// the engine is the db every other test package shares, and a partial sync
	// would leave those tables missing columns for all of them.
	require.NoError(t, x.Sync2(&models.Label{}, &models.LabelTask{}))

	keptLabel := &models.Label{Title: "kept 20260928140648", CreatedByID: 1}
	_, err = x.Insert(keptLabel)
	require.NoError(t, err)
	goneLabel := &models.Label{Title: "gone 20260928140648", CreatedByID: 1}
	_, err = x.Insert(goneLabel)
	require.NoError(t, err)

	kept := &models.LabelTask{TaskID: 1, LabelID: keptLabel.ID}
	orphan := &models.LabelTask{TaskID: 1, LabelID: goneLabel.ID}
	_, err = x.Insert(kept, orphan)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := x.In("id", kept.ID, orphan.ID).Delete(&models.LabelTask{})
		require.NoError(t, err)
		_, err = x.In("id", keptLabel.ID, goneLabel.ID).Delete(&models.Label{})
		require.NoError(t, err)
	})

	_, err = x.ID(goneLabel.ID).Delete(&models.Label{})
	require.NoError(t, err)

	require.NoError(t, deleteOrphanedLabelTasks20260928140648(x))

	has, err := x.ID(orphan.ID).Exist(&models.LabelTask{})
	require.NoError(t, err)
	require.False(t, has, "orphaned label_tasks row should be gone")

	has, err = x.ID(kept.ID).Exist(&models.LabelTask{})
	require.NoError(t, err)
	require.True(t, has, "label_tasks row of an existing label must stay")

	// Idempotent once there is nothing left to clean.
	require.NoError(t, deleteOrphanedLabelTasks20260928140648(x))
}
