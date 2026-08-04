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
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
)

func TestDeleteExpiredProjects(t *testing.T) {
	// Projects 44, 45 and 46 were soft-deleted at this time in the fixtures
	deletedAt := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)

	t.Run("older than the retention period", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)

		deleteExpiredProjects(deletedAt.Add(ProjectDeleteRetention + 24*time.Hour))

		db.AssertMissing(t, "projects", map[string]interface{}{"id": 44})
		db.AssertMissing(t, "projects", map[string]interface{}{"id": 45})
		db.AssertMissing(t, "projects", map[string]interface{}{"id": 46})
	})

	t.Run("newer than the retention period", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		files.InitTestFileFixtures(t)

		deleteExpiredProjects(deletedAt.Add(ProjectDeleteRetention - 24*time.Hour))

		db.AssertExists(t, "projects", map[string]interface{}{"id": 44}, false)
	})
}
