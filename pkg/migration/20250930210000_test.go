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
	"code.vikunja.io/api/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestBackfillLegacyTaskCreators(t *testing.T) {
	db.LoadAndAssertFixtures(t)

	engine := db.GetEngine()
	session := engine.NewSession()
	defer session.Close()

	legacyTasks := []*models.Task{
		{
			Title:       "legacy project 1",
			ProjectID:   1,
			CreatedByID: 0,
			Index:       995,
			Created:     time.Now(),
			Updated:     time.Now(),
		},
		{
			Title:       "legacy project 2",
			ProjectID:   2,
			CreatedByID: 0,
			Index:       996,
			Created:     time.Now(),
			Updated:     time.Now(),
		},
		{
			Title:       "legacy project missing owner",
			ProjectID:   999,
			CreatedByID: 0,
			Index:       997,
			Created:     time.Now(),
			Updated:     time.Now(),
		},
	}

	for _, task := range legacyTasks {
		_, err := session.Insert(task)
		require.NoError(t, err)
	}

	err := backfillLegacyTaskCreators(engine)
	require.NoError(t, err)

	type result struct {
		CreatedByID int64 `xorm:"created_by_id"`
	}

	for i, task := range legacyTasks {
		var row result
		has, err := engine.Table("tasks").Where("id = ?", task.ID).Get(&row)
		require.NoError(t, err)
		require.True(t, has)

		expected := int64(1)
		if i == 1 {
			expected = 3
		}
		require.Equal(t, expected, row.CreatedByID)
	}
}
