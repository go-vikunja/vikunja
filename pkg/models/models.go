//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"encoding/gob"
	"fmt"
	_ "github.com/go-sql-driver/mysql" // Because.
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	xrc "github.com/go-xorm/xorm-redis-cache"
	_ "github.com/mattn/go-sqlite3" // Because.
	"github.com/spf13/viper"
)

var (
	x *xorm.Engine

	tables            []interface{}
	tablesWithPointer []interface{}
)

func getEngine() (*xorm.Engine, error) {
	// Use Mysql if set
	if viper.GetString("database.type") == "mysql" {
		connStr := fmt.Sprintf(
			"%s:%s@tcp(%s)/%s?charset=utf8&parseTime=true",
			viper.GetString("database.user"),
			viper.GetString("database.password"),
			viper.GetString("database.host"),
			viper.GetString("database.database"))
		e, err := xorm.NewEngine("mysql", connStr)
		e.SetMaxOpenConns(viper.GetInt("database.openconnections"))
		return e, err
	}

	// Otherwise use sqlite
	path := viper.GetString("database.path")
	if path == "" {
		path = "./db.db"
	}
	return xorm.NewEngine("sqlite3", path)
}

func init() {
	tables = append(tables,
		new(User),
		new(List),
		new(ListTask),
		new(Team),
		new(TeamMember),
		new(TeamList),
		new(TeamNamespace),
		new(Namespace),
		new(ListUser),
		new(NamespaceUser),
		new(ListTaskAssginee),
		new(Label),
		new(LabelTask),
	)

	tablesWithPointer = append(tables,
		&User{},
		&List{},
		&ListTask{},
		&Team{},
		&TeamMember{},
		&TeamList{},
		&TeamNamespace{},
		&Namespace{},
		&ListUser{},
		&NamespaceUser{},
		&ListTaskAssginee{},
		&Label{},
		&LabelTask{},
	)
}

// SetEngine sets the xorm.Engine
func SetEngine() (err error) {
	x, err = getEngine()
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %v", err)
	}

	// Cache
	if viper.GetBool("cache.enabled") {
		switch viper.GetString("cache.type") {
		case "memory":
			cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), viper.GetInt("cache.maxelementsize"))
			x.SetDefaultCacher(cacher)
			break
		case "redis":
			cacher := xrc.NewRedisCacher(viper.GetString("redis.host"), viper.GetString("redis.password"), xrc.DEFAULT_EXPIRATION, x.Logger())
			x.SetDefaultCacher(cacher)
			gob.Register(tables)
			gob.Register(tablesWithPointer) // Need to register tables with pointer as well...
			break
		default:
			fmt.Println("Did not find a valid cache type. Caching disabled. Please refer to the docs for poosible cache types.")
		}
	}

	x.SetMapper(core.GonicMapper{})

	// Sync dat shit
	if err = x.StoreEngine("InnoDB").Sync2(tables...); err != nil {
		return fmt.Errorf("sync database struct error: %v", err)
	}

	x.ShowSQL(viper.GetBool("database.showqueries"))

	return nil
}

func getLimitFromPageIndex(page int) (limit, start int) {

	// Get everything when page index is -1
	if page < 0 {
		return 0, 0
	}

	limit = viper.GetInt("service.pagecount")
	start = limit * (page - 1)
	return
}

// GetTotalCount returns the total amount of something
func GetTotalCount(counting interface{}) (count int64, err error) {
	return x.Count(counting)
}
