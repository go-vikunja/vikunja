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
	"os"
	"sort"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// You can get the id string for new migrations by running `date +%Y%m%d%H%M%S` on a unix system.

var migrations []*xormigrate.Migration

// AddPluginMigrations adds migrations provided by plugins to the global list.
func AddPluginMigrations(ms []*xormigrate.Migration) {
	migrations = append(migrations, ms...)
}

// A helper function because we need a migration in various places which we can't really solve with an init() function.
func initMigration(x *xorm.Engine) *xormigrate.Xormigrate {
	// Get our own xorm engine if we don't have one
	if x == nil {
		var err error
		x, err = db.CreateDBEngine()
		if err != nil {
			log.Fatalf("Could not connect to db: %v", err.Error())
			return nil
		}
	}

	// Because init() does not guarantee the order in which these are added to the slice,
	// we need to sort them to ensure that they are in order
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	m := xormigrate.New(x, migrations)
	logger := log.NewXormLogger(config.LogEnabled.GetBool(), config.LogEvents.GetString(), config.LogEventsLevel.GetString(), config.LogFormat.GetString())
	m.SetLogger(logger)
	m.InitSchema(initSchema)
	return m
}

// Migrate runs all migrations
func Migrate(x *xorm.Engine) {
	log.Info("Running migrations…")
	m := initMigration(x)
	err := m.Migrate()
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Info("Ran all migrations successfully.")
}

// ListMigrations pretty-prints a list with all migrations.
func ListMigrations() {
	x, err := db.CreateDBEngine()
	if err != nil {
		log.Fatalf("Could not connect to db: %v", err.Error())
	}
	ms := []*xormigrate.Migration{}
	err = x.Find(&ms)
	if err != nil {
		log.Fatalf("Error getting migration table: %v", err.Error())
	}

	table := tablewriter.NewTable(
		os.Stdout,
		tablewriter.WithHeader([]string{"ID", "Description"}),
		tablewriter.WithAlignment(tw.Alignment{tw.AlignLeft}),
	)

	for _, m := range ms {
		_ = table.Append([]string{m.ID, m.Description})
	}
	_ = table.Render()
}

// Rollback rolls back all migrations until a certain point.
func Rollback(migrationID string) {
	m := initMigration(nil)
	err := m.RollbackTo(migrationID)
	if err != nil {
		log.Fatalf("Could not rollback: %v", err)
	}
	log.Info("Rolled back successfully.")
}

// MigrateTo executes all migrations up to a certain point
func MigrateTo(migrationID string, x *xorm.Engine) error {
	m := initMigration(x)
	return m.MigrateTo(migrationID)
}

// Deletes a column from a table. All arguments are strings, to let them be standalone and not depending on any struct.
func dropTableColum(x *xorm.Engine, tableName, col string) error {

	switch config.DatabaseType.GetString() {
	case "sqlite":
		log.Warning("Unable to drop columns in SQLite")
	case "mysql":
		_, err := x.Exec("ALTER TABLE " + tableName + " DROP COLUMN " + col)
		if err != nil {
			return err
		}
	case "postgres":
		_, err := x.Exec("ALTER TABLE " + tableName + " DROP COLUMN " + col)
		if err != nil {
			return err
		}
	default:
		log.Fatal("Unknown db.")
	}
	return nil
}

// Modifies a column definition
func modifyColumn(x *xorm.Engine, tableName, col, newDefinition string) error {
	switch config.DatabaseType.GetString() {
	case "sqlite":
		log.Warning("Unable to modify columns in SQLite")
	case "mysql":
		_, err := x.Exec("ALTER TABLE " + tableName + " MODIFY COLUMN " + col + " " + newDefinition)
		if err != nil {
			return err
		}
	case "postgres":
		_, err := x.Exec("ALTER TABLE " + tableName + " ALTER COLUMN " + col + " " + newDefinition)
		if err != nil {
			return err
		}
	default:
		log.Fatal("Unknown db.")
	}
	return nil
}

func renameTable(x *xorm.Engine, oldName, newName string) error {
	switch config.DatabaseType.GetString() {
	case "sqlite":
		_, err := x.Exec("ALTER TABLE `" + oldName + "` RENAME TO `" + newName + "`")
		if err != nil {
			return err
		}
	case "mysql":
		_, err := x.Exec("RENAME TABLE `" + oldName + "` TO `" + newName + "`")
		if err != nil {
			return err
		}
	case "postgres":
		_, err := x.Exec("ALTER TABLE `" + oldName + "` RENAME TO `" + newName + "`")
		if err != nil {
			return err
		}
	default:
		log.Fatal("Unknown db.")
	}
	return nil
}

// Checks if a column exists in a table
func columnExists(x *xorm.Engine, tableName, columnName string) (bool, error) {
	switch config.DatabaseType.GetString() {
	case "sqlite":
		results, err := x.Query("PRAGMA table_info(" + tableName + ")")
		if err != nil {
			return false, err
		}

		for _, row := range results {
			if name, ok := row["name"]; ok && string(name) == columnName {
				return true, nil
			}
		}
		return false, nil
	case "mysql":
		results, err := x.Query("SHOW COLUMNS FROM `" + tableName + "` LIKE '" + columnName + "'")
		if err != nil {
			return false, err
		}
		return len(results) > 0, nil
	case "postgres":
		results, err := x.Query("SELECT column_name FROM information_schema.columns WHERE table_name = '" + tableName + "' AND column_name = '" + columnName + "'")
		if err != nil {
			return false, err
		}
		return len(results) > 0, nil
	default:
		log.Fatal("Unknown db.")
		return false, nil
	}
}

func renameColumn(x *xorm.Engine, tableName, oldColumn, newColumn string) error {
	// Check if old column exists
	exists, err := columnExists(x, tableName, oldColumn)
	if err != nil {
		return err
	}
	if !exists {
		log.Debugf("Column %s in table %s does not exist, skipping rename", oldColumn, tableName)
		return nil
	}

	// Check if new column already exists
	newExists, err := columnExists(x, tableName, newColumn)
	if err != nil {
		return err
	}
	if newExists {
		log.Debugf("Column %s in table %s already exists, skipping rename", newColumn, tableName)
		return nil
	}

	switch config.DatabaseType.GetString() {
	case "sqlite":
		_, err := x.Exec("ALTER TABLE \"" + tableName + "\" RENAME COLUMN \"" + oldColumn + "\" TO \"" + newColumn + "\"")
		if err != nil {
			return err
		}
	case "mysql":
		_, err := x.Exec("ALTER TABLE `" + tableName + "` CHANGE `" + oldColumn + "` `" + newColumn + "` BIGINT NOT NULL DEFAULT 0")
		if err != nil {
			return err
		}
	case "postgres":
		_, err := x.Exec("ALTER TABLE \"" + tableName + "\" RENAME COLUMN \"" + oldColumn + "\" TO \"" + newColumn + "\"")
		if err != nil {
			return err
		}
	default:
		log.Fatal("Unknown db.")
	}
	return nil
}

func initSchema(tx *xorm.Engine) error {
	schemeBeans := []interface{}{}
	schemeBeans = append(schemeBeans, models.GetTables()...)
	schemeBeans = append(schemeBeans, files.GetTables()...)
	schemeBeans = append(schemeBeans, migration.GetTables()...)
	schemeBeans = append(schemeBeans, user.GetTables()...)
	schemeBeans = append(schemeBeans, notifications.GetTables()...)
	return tx.Sync2(schemeBeans...)
}
