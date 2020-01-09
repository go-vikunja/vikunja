// Vikunja is a todo-list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package db

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"encoding/gob"
	"fmt"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	"os"
	"strconv"
	"time"

	xrc "github.com/go-xorm/xorm-redis-cache"

	_ "github.com/go-sql-driver/mysql" // Because.
	_ "github.com/mattn/go-sqlite3"    // Because.
)

// We only want one instance of the engine, so we can reate it once and reuse it
var x *xorm.Engine

// CreateDBEngine initializes a db engine from the config
func CreateDBEngine() (engine *xorm.Engine, err error) {

	if x != nil {
		return x, nil
	}

	// If the database type is not set, this likely means we need to initialize the config first
	if config.DatabaseType.GetString() == "" {
		config.InitConfig()
	}

	// Use Mysql if set
	if config.DatabaseType.GetString() == "mysql" {
		engine, err = initMysqlEngine()
		if err != nil {
			return
		}
	} else {
		// Otherwise use sqlite
		engine, err = initSqliteEngine()
		if err != nil {
			return
		}
	}

	engine.SetMapper(core.GonicMapper{})
	engine.ShowSQL(config.LogDatabase.GetString() != "off")
	engine.SetLogger(xorm.NewSimpleLogger(log.GetLogWriter("database")))

	// Cache
	// We have to initialize the cache here to avoid import cycles
	if config.CacheEnabled.GetBool() {
		switch config.CacheType.GetString() {
		case "memory":
			cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), config.CacheMaxElementSize.GetInt())
			engine.SetDefaultCacher(cacher)
		case "redis":
			cacher := xrc.NewRedisCacher(config.RedisEnabled.GetString(), config.RedisPassword.GetString(), xrc.DEFAULT_EXPIRATION, engine.Logger())
			engine.SetDefaultCacher(cacher)
		default:
			log.Info("Did not find a valid cache type. Caching disabled. Please refer to the docs for poosible cache types.")
		}
	}

	x = engine
	return
}

// CreateTestEngine creates an instance of the db engine which lives in memory
func CreateTestEngine() (engine *xorm.Engine, err error) {

	if x != nil {
		return x, nil
	}

	if os.Getenv("VIKUNJA_TESTS_USE_CONFIG") == "1" {
		config.InitConfig()
		engine, err = CreateDBEngine()
		if err != nil {
			return nil, err
		}
	} else {
		engine, err = xorm.NewEngine("sqlite3", "file::memory:?cache=shared")
		if err != nil {
			return nil, err
		}
	}

	engine.SetMapper(core.GonicMapper{})
	engine.ShowSQL(os.Getenv("UNIT_TESTS_VERBOSE") == "1")
	engine.SetLogger(xorm.NewSimpleLogger(log.GetLogWriter("database")))
	x = engine
	return
}

// RegisterTableStructsForCache registers tables in gob encoding for redis cache
func RegisterTableStructsForCache(val interface{}) {
	gob.Register(val)
}

func initMysqlEngine() (engine *xorm.Engine, err error) {
	// We're using utf8mb here instead of just utf8 because we want to use non-BMP characters.
	// See https://stackoverflow.com/a/30074553/10924593 for more info.
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true",
		config.DatabaseUser.GetString(),
		config.DatabasePassword.GetString(),
		config.DatabaseHost.GetString(),
		config.DatabaseDatabase.GetString())
	engine, err = xorm.NewEngine("mysql", connStr)
	if err != nil {
		return
	}
	engine.SetMaxOpenConns(config.DatabaseMaxOpenConnections.GetInt())
	engine.SetMaxIdleConns(config.DatabaseMaxIdleConnections.GetInt())
	max, err := time.ParseDuration(strconv.Itoa(config.DatabaseMaxConnectionLifetime.GetInt()) + `ms`)
	if err != nil {
		return
	}
	engine.SetConnMaxLifetime(max)
	return
}

func initSqliteEngine() (engine *xorm.Engine, err error) {
	path := config.DatabasePath.GetString()
	if path == "" {
		path = "./db.db"
	}

	return xorm.NewEngine("sqlite3", path)
}
