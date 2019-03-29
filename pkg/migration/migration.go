//   Vikunja is a todo-list application to facilitate your life.
//   Copyright 2019 Vikunja and contributors. All rights reserved.
//
//   This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License
//   along with this program.  If not, see <https://www.gnu.org/licenses/>.

package migration

import (
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"github.com/go-xorm/xorm"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/viper"
	"os"
	"sort"
	"src.techknowlogick.com/xormigrate"
)

// You can get the id string for new migrations by running `date +%Y%m%d%H%M%S` on a unix system.

var migrations []*xormigrate.Migration

// A helper function because we need a migration in various places which we can't really solve with an init() function.
func initMigration(x *xorm.Engine) *xormigrate.Xormigrate {
	// Get our own xorm engine if we don't have one
	if x == nil {
		var err error
		x, err = db.CreateDBEngine()
		if err != nil {
			log.Log.Criticalf("Could not connect to db: %v", err.Error())
			return nil
		}
	}

	// Because init() does not guarantee the order in which these are added to the slice,
	// we need to sort them to ensure that they are in order
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].ID < migrations[j].ID
	})

	m := xormigrate.New(x, migrations)
	m.NewLogger(log.GetLogWriter("database"))
	m.InitSchema(initSchema)
	return m
}

// Migrate runs all migrations
func Migrate(x *xorm.Engine) {
	m := initMigration(x)
	err := m.Migrate()
	if err != nil {
		log.Log.Fatalf("Migration failed: %v", err)
	}
	log.Log.Info("Ran all migrations successfully.")
}

// ListMigrations pretty-prints a list with all migrations.
func ListMigrations() {
	x, err := db.CreateDBEngine()
	if err != nil {
		log.Log.Fatalf("Could not connect to db: %v", err.Error())
	}
	ms := []*xormigrate.Migration{}
	err = x.Find(&ms)
	if err != nil {
		log.Log.Fatalf("Error getting migration table: %v", err.Error())
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Description"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor})

	for _, m := range ms {
		table.Append([]string{m.ID, m.Description})
	}
	table.Render()
}

// Rollback rolls back all migrations until a certain point.
func Rollback(migrationID string) {
	m := initMigration(nil)
	err := m.RollbackTo(migrationID)
	if err != nil {
		log.Log.Fatalf("Could not rollback: %v", err)
	}
	log.Log.Info("Rolled back successfully.")
}

// Deletes a column from a table. All arguments are strings, to let them be standalone and not depending on any struct.
func dropTableColum(x *xorm.Engine, tableName, col string) error {

	switch viper.GetString("database.type") {
	case "sqlite":
		log.Log.Warning("Unable to drop columns in SQLite")
	case "mysql":
		_, err := x.Exec("ALTER TABLE " + tableName + " DROP COLUMN " + col)
		if err != nil {
			return err
		}
	default:
		log.Log.Fatal("Unknown db.")
	}
	return nil
}

func initSchema(tx *xorm.Engine) error {
	return tx.Sync2(
		models.GetTables()...,
	)
}
