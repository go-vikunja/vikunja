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

package vikunjafile

import (
	"os"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/require"
)

func TestVikunjaFileMigrator_Migrate(t *testing.T) {
	t.Run("migrate successfully", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		m := &FileMigrator{}
		u := &user.User{ID: 1}

		f, err := os.Open(config.ServiceRootpath.GetString() + "/pkg/modules/migration/vikunja-file/export.zip")
		if err != nil {
			t.Fatalf("Could not open file: %s", err)
		}
		defer f.Close()
		s, err := f.Stat()
		if err != nil {
			t.Fatalf("Could not stat file: %s", err)
		}

		err = m.Migrate(u, f, s.Size())
		require.NoError(t, err)
		db.AssertExists(t, "projects", map[string]interface{}{
			"title":    "test project",
			"owner_id": u.ID,
		}, false)
		db.AssertExists(t, "projects", map[string]interface{}{
			"title":    "Inbox",
			"owner_id": u.ID,
		}, false)
		db.AssertExists(t, "tasks", map[string]interface{}{
			"title":         "some other task",
			"created_by_id": u.ID,
		}, false)
		db.AssertExists(t, "task_comments", map[string]interface{}{
			"comment":   "This is a comment",
			"author_id": u.ID,
		}, false)
		db.AssertExists(t, "files", map[string]interface{}{
			"name":          "grant-whitty-546453-unsplash.jpg",
			"created_by_id": u.ID,
		}, false)
		db.AssertExists(t, "labels", map[string]interface{}{
			"title":         "test",
			"created_by_id": u.ID,
		}, false)
		db.AssertExists(t, "buckets", map[string]interface{}{
			"title":         "Test Bucket",
			"created_by_id": u.ID,
		}, false)
	})
	t.Run("should not accept an old import", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		m := &FileMigrator{}
		u := &user.User{ID: 1}

		f, err := os.Open(config.ServiceRootpath.GetString() + "/pkg/modules/migration/vikunja-file/export_pre_0.21.0.zip")
		if err != nil {
			t.Fatalf("Could not open file: %s", err)
		}
		defer f.Close()
		s, err := f.Stat()
		if err != nil {
			t.Fatalf("Could not stat file: %s", err)
		}

		err = m.Migrate(u, f, s.Size())
		require.Error(t, err)
		require.ErrorContainsf(t, err, "export was created with an older version", "Invalid error message")
	})
}
