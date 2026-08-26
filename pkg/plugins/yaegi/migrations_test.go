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

package yaegi

import (
	"testing"

	_ "github.com/mattn/go-sqlite3" // Needed to open the in-memory engine the migration runs against.
	"xorm.io/xorm"
)

const migrationPluginDir = "testdata/migrationplugin"

func TestLoadPluginWithMigrations(t *testing.T) {
	loaded, err := LoadPluginFull(migrationPluginDir)
	if err != nil {
		t.Fatalf("LoadPluginFull failed: %v", err)
	}

	if loaded.Migration == nil {
		t.Fatal("Migration is nil — typed factory NewMigrationPlugin not found")
	}

	migrations := loaded.Migration.Migrations()
	if len(migrations) != 1 {
		t.Fatalf("expected 1 migration, got %d", len(migrations))
	}
	if migrations[0].ID != "20260101000000-create-plugin-migration-test" {
		t.Errorf("unexpected migration id %q", migrations[0].ID)
	}

	// The interpreted MigrateFunc has to survive the interface boundary and run
	// against a real engine, not just be non-nil.
	engine, err := xorm.NewEngine("sqlite3", "file::memory:")
	if err != nil {
		t.Fatalf("could not create test engine: %v", err)
	}
	defer engine.Close()

	if err := migrations[0].Migrate(engine); err != nil {
		t.Fatalf("running the plugin migration failed: %v", err)
	}

	exists, err := engine.IsTableExist("plugin_migration_test")
	if err != nil {
		t.Fatalf("checking for the migrated table failed: %v", err)
	}
	if !exists {
		t.Error("plugin migration ran but did not create its table")
	}
}
