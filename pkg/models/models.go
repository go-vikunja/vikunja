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
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"encoding/gob"
	_ "github.com/go-sql-driver/mysql" // Because.
	"github.com/go-xorm/xorm"
	xrc "github.com/go-xorm/xorm-redis-cache"
	_ "github.com/mattn/go-sqlite3" // Because.
	"github.com/spf13/viper"
)

var (
	x *xorm.Engine
)

// GetTables returns all structs which are also a table.
func GetTables() []interface{} {
	return []interface{}{
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
	}
}

// SetEngine sets the xorm.Engine
func SetEngine() (err error) {
	x, err = db.CreateDBEngine()
	if err != nil {
		log.Log.Criticalf("Could not connect to db: %v", err.Error())
		return
	}

	// Cache
	// We have to initialize the cache here to avoid import cycles
	if viper.GetBool("cache.enabled") {
		switch viper.GetString("cache.type") {
		case "memory":
			cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), viper.GetInt("cache.maxelementsize"))
			x.SetDefaultCacher(cacher)
		case "redis":
			cacher := xrc.NewRedisCacher(viper.GetString("redis.host"), viper.GetString("redis.password"), xrc.DEFAULT_EXPIRATION, x.Logger())
			x.SetDefaultCacher(cacher)
			gob.Register(GetTables())
		default:
			log.Log.Info("Did not find a valid cache type. Caching disabled. Please refer to the docs for poosible cache types.")
		}
	}

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
