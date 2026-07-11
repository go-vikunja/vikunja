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

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
)

// The CalDAV ctag derives from the project's updated timestamp, so the
// listener must bump it on sub-entity changes or clients never refetch.
func TestHandleTaskUpdateLastUpdated(t *testing.T) {
	t.Run("bumps task and project updated times", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		taskBefore, err := GetTaskByIDSimple(s, 1)
		require.NoError(t, err)
		projectBefore, err := GetProjectSimpleByID(s, 1)
		require.NoError(t, err)
		_ = s.Close()

		events.TestListener(t, &TaskRelationCreatedEvent{
			Task: &taskBefore,
			Doer: &user.User{ID: 1},
		}, &HandleTaskUpdateLastUpdated{})

		s2 := db.NewSession()
		defer s2.Close()
		taskAfter, err := GetTaskByIDSimple(s2, 1)
		require.NoError(t, err)
		projectAfter, err := GetProjectSimpleByID(s2, 1)
		require.NoError(t, err)

		require.True(t, taskAfter.Updated.After(taskBefore.Updated), "task updated time must advance")
		require.True(t, projectAfter.Updated.After(projectBefore.Updated), "project updated time must advance")
	})

	t.Run("does not fail for a nonexistent task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		events.TestListener(t, &TaskRelationCreatedEvent{
			Task: &Task{ID: 99999},
			Doer: &user.User{ID: 1},
		}, &HandleTaskUpdateLastUpdated{})
	})
}
