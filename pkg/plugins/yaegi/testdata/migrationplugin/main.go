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

package main

import (
	"code.vikunja.io/api/pkg/plugins"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type MigrationTestPlugin struct{}

// Interpreted types reach xorm as anonymous reflect structs with no methods, so
// TableName() is invisible and the table name has to be passed via Table().
type pluginData struct {
	ID   int64  `xorm:"pk autoincr"`
	Name string `xorm:"varchar(250)"`
}

func (p *MigrationTestPlugin) Name() string    { return "migration-test" }
func (p *MigrationTestPlugin) Version() string { return "1.0.0" }
func (p *MigrationTestPlugin) Init() error     { return nil }
func (p *MigrationTestPlugin) Shutdown() error { return nil }

func (p *MigrationTestPlugin) Migrations() []*xormigrate.Migration {
	return []*xormigrate.Migration{
		{
			ID:          "20260101000000-create-plugin-migration-test",
			Description: "Create the plugin migration test table",
			Migrate: func(tx *xorm.Engine) error {
				return tx.Table("plugin_migration_test").Sync2(&pluginData{})
			},
			Rollback: func(tx *xorm.Engine) error {
				return tx.DropTables("plugin_migration_test")
			},
		},
	}
}

func NewPlugin() plugins.Plugin { return &MigrationTestPlugin{} }

func NewMigrationPlugin() plugins.MigrationPlugin { return &MigrationTestPlugin{} }
