// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package db

import (
	"encoding/gob"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	xrc "gitea.com/xorm/xorm-redis-cache"
	"xorm.io/core"
	"xorm.io/xorm"
	"xorm.io/xorm/caches"
	"xorm.io/xorm/schemas"

	_ "github.com/go-sql-driver/mysql" // Because.
	_ "github.com/lib/pq"              // Because.
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
	switch config.DatabaseType.GetString() {
	case "mysql":
		engine, err = initMysqlEngine()
		if err != nil {
			return
		}
	case "postgres":
		engine, err = initPostgresEngine()
		if err != nil {
			return
		}
	case "sqlite":
		// Otherwise use sqlite
		engine, err = initSqliteEngine()
		if err != nil {
			return
		}
	default:
		log.Fatalf("Unknown database type %s", config.DatabaseType.GetString())
	}

	engine.SetTZLocation(config.GetTimeZone()) // Vikunja's timezone
	loc, err := time.LoadLocation("GMT")       // The db data timezone
	if err != nil {
		log.Fatalf("Error parsing time zone: %s", err)
	}
	engine.SetTZDatabase(loc)
	engine.SetMapper(core.GonicMapper{})
	logger := log.NewXormLogger("")
	engine.SetLogger(logger)

	// Cache
	// We have to initialize the cache here to avoid import cycles
	if config.CacheEnabled.GetBool() {
		switch config.CacheType.GetString() {
		case "memory":
			cacher := caches.NewLRUCacher(caches.NewMemoryStore(), config.CacheMaxElementSize.GetInt())
			engine.SetDefaultCacher(cacher)
		case "redis":
			cacher := xrc.NewRedisCacher(config.RedisHost.GetString(), config.RedisPassword.GetString(), xrc.DEFAULT_EXPIRATION, engine.Logger())
			engine.SetDefaultCacher(cacher)
		default:
			log.Info("Did not find a valid cache type. Caching disabled. Please refer to the docs for poosible cache types.")
		}
	}

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

// parsePostgreSQLHostPort parses given input in various forms defined in
// https://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING
// and returns proper host and port number.
func parsePostgreSQLHostPort(info string) (string, string) {
	host, port := "127.0.0.1", "5432"
	if strings.Contains(info, ":") && !strings.HasSuffix(info, "]") {
		idx := strings.LastIndex(info, ":")
		host = info[:idx]
		port = info[idx+1:]
	} else if len(info) > 0 {
		host = info
	}
	return host, port
}

func initPostgresEngine() (engine *xorm.Engine, err error) {
	host, port := parsePostgreSQLHostPort(config.DatabaseHost.GetString())
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host,
		port,
		url.PathEscape(config.DatabaseUser.GetString()),
		url.PathEscape(config.DatabasePassword.GetString()),
		config.DatabaseDatabase.GetString(),
		config.DatabaseSslMode.GetString(),
	)

	engine, err = xorm.NewEngine("postgres", connStr)
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

	// Try opening the db file to return a better error message if that does not work
	var exists = true
	if _, err := os.Stat(path); err != nil {
		exists = !os.IsNotExist(err)
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0)
	if err != nil {
		return nil, fmt.Errorf("could not open database file [uid=%d, gid=%d]: %s", os.Getuid(), os.Getgid(), err)
	}
	_ = file.Close() // We directly close the file because we only want to check if it is writable. It will be reopened lazily later by xorm.

	if !exists {
		_ = os.Remove(path) // Remove the file to not prevent the db from creating another one
	}

	return xorm.NewEngine("sqlite3", path)
}

// WipeEverything wipes all tables and their data. Use with caution...
func WipeEverything() error {

	tables, err := x.DBMetas()
	if err != nil {
		return err
	}

	for _, t := range tables {
		if err := x.DropTables(t.Name); err != nil {
			return err
		}
	}

	return nil
}

// NewSession creates a new xorm session
func NewSession() *xorm.Session {
	return x.NewSession()
}

// Type returns the db type of the currently configured db
func Type() schemas.DBType {
	return x.Dialect().URI().DBType
}
