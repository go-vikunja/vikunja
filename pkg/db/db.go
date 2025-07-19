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

package db

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"

	"xorm.io/builder"
	"xorm.io/xorm"
	"xorm.io/xorm/names"
	"xorm.io/xorm/schemas"

	_ "github.com/go-sql-driver/mysql" // Because.
	_ "github.com/lib/pq"              // Because.
	_ "github.com/mattn/go-sqlite3"    // Because.
)

var (
	// We only want one instance of the engine, so we can create it once and reuse it
	x *xorm.Engine
	// paradedbInstalled marks whether the paradedb extension is available
	// and can be used for full text search.
	paradedbInstalled bool
)

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
	engine.SetMapper(names.GonicMapper{})
	logger := log.NewXormLogger(config.LogEnabled.GetBool(), config.LogDatabase.GetString(), config.LogDatabaseLevel.GetString(), config.LogFormat.GetString())
	engine.SetLogger(logger)

	x = engine
	return
}

func initMysqlEngine() (engine *xorm.Engine, err error) {
	// We're using utf8mb here instead of just utf8 because we want to use non-BMP characters.
	// See https://stackoverflow.com/a/30074553/10924593 for more info.
	host := fmt.Sprintf("tcp(%s)", config.DatabaseHost.GetString())
	if config.DatabaseHost.GetString()[0] == '/' { // looks like a unix socket
		host = fmt.Sprintf("unix(%s)", config.DatabaseHost.GetString())
	}

	connStr := fmt.Sprintf(
		"%s:%s@%s/%s?charset=utf8mb4&parseTime=true&tls=%s",
		config.DatabaseUser.GetString(),
		config.DatabasePassword.GetString(),
		host,
		config.DatabaseDatabase.GetString(),
		config.DatabaseTLS.GetString())
	engine, err = xorm.NewEngine("mysql", connStr)
	if err != nil {
		return
	}
	engine.SetMaxOpenConns(config.DatabaseMaxOpenConnections.GetInt())
	engine.SetMaxIdleConns(config.DatabaseMaxIdleConnections.GetInt())
	maxLifetime, err := time.ParseDuration(strconv.Itoa(config.DatabaseMaxConnectionLifetime.GetInt()) + `ms`)
	if err != nil {
		return
	}
	engine.SetConnMaxLifetime(maxLifetime)
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

// Copied and adopted from https://github.com/go-gitea/gitea/blob/f337c32e868381c6d2d948221aca0c59f8420c13/modules/setting/database.go#L176-L186
func getPostgreSQLConnectionString(dbHost, dbUser, dbPasswd, dbName, dbSslMode, dbSslCert, dbSslKey, dbSslRootCert string) (connStr string) {
	dbParam := "?"
	if strings.Contains(dbName, dbParam) {
		dbParam = "&"
	}
	host, port := parsePostgreSQLHostPort(dbHost)
	if host[0] == '/' { // looks like a unix socket
		connStr = fmt.Sprintf("postgres://%s:%s@:%s/%s%ssslmode=%s&sslcert=%s&sslkey=%s&sslrootcert=%s&host=%s",
			url.PathEscape(dbUser), url.PathEscape(dbPasswd), port, dbName, dbParam, dbSslMode, dbSslCert, dbSslKey, dbSslRootCert, host)
	} else {
		connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s%ssslmode=%s&sslcert=%s&sslkey=%s&sslrootcert=%s",
			url.PathEscape(dbUser), url.PathEscape(dbPasswd), host, port, dbName, dbParam, dbSslMode, dbSslCert, dbSslKey, dbSslRootCert)
	}
	return connStr
}

func initPostgresEngine() (engine *xorm.Engine, err error) {
	connStr := getPostgreSQLConnectionString(
		config.DatabaseHost.GetString(),
		config.DatabaseUser.GetString(),
		config.DatabasePassword.GetString(),
		config.DatabaseDatabase.GetString(),
		config.DatabaseSslMode.GetString(),
		config.DatabaseSslCert.GetString(),
		config.DatabaseSslKey.GetString(),
		config.DatabaseSslRootCert.GetString(),
	)

	engine, err = xorm.NewEngine("postgres", connStr)
	if err != nil {
		return
	}
	engine.SetSchema(config.DatabaseSchema.GetString())
	engine.SetMaxOpenConns(config.DatabaseMaxOpenConnections.GetInt())
	engine.SetMaxIdleConns(config.DatabaseMaxIdleConnections.GetInt())
	maxLifetime, err := time.ParseDuration(strconv.Itoa(config.DatabaseMaxConnectionLifetime.GetInt()) + `ms`)
	if err != nil {
		return
	}
	engine.SetConnMaxLifetime(maxLifetime)

	checkParadeDB(engine)
	return
}

func initSqliteEngine() (engine *xorm.Engine, err error) {
	path := config.DatabasePath.GetString()

	if path != "memory" && !filepath.IsAbs(path) {
		path = filepath.Join(config.ServiceRootpath.GetString(), path)
	}

	if path == "memory" {
		return xorm.NewEngine("sqlite3", "file::memory:?cache=shared")
	}

	// Try opening the db file to return a better error message if that does not work
	var exists = true
	if _, err := os.Stat(path); err != nil {
		exists = !os.IsNotExist(err)
	}
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0)
	if err != nil {
		return nil, fmt.Errorf("could not open database file [uid=%d, gid=%d]: %w", os.Getuid(), os.Getgid(), err)
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

func GetDialect() string {
	dialect := config.DatabaseType.GetString()
	if dialect == "sqlite" {
		dialect = builder.SQLITE
	}

	return dialect
}

func checkParadeDB(engine *xorm.Engine) {
	if engine.Dialect().URI().DBType != schemas.POSTGRES {
		return
	}

	exists := false
	if _, err := engine.SQL("SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname='pg_search')").Get(&exists); err != nil {
		log.Errorf("could not check for paradedb extension: %v", err)
		return
	}

	if !exists {
		return
	}

	paradedbInstalled = true
	log.Debug("ParadeDB extension detected, using @@@ search operator")
}

func CreateParadeDBIndexes() error {
	if !paradedbInstalled {
		return nil
	}
	// ParadeDB only allows one bm25 index per table, so we create a single index covering both fields
	// Use optimized configuration with fast fields and field boosting for better performance
	indexSQL := `CREATE INDEX IF NOT EXISTS idx_tasks_paradedb ON tasks USING bm25 (id, title, description, project_id, done) 
	WITH (
		key_field='id',
		text_fields='{
			"title": {"fast": true, "record": "freq"}, 
			"description": {"fast": true, "record": "freq"}
		}',
		numeric_fields='{
			"project_id": {"fast": true}
		}',
		boolean_fields='{
			"done": {"fast": true}
		}'
	)`
	if _, err := x.Exec(indexSQL); err != nil {
		return fmt.Errorf("could not ensure paradedb task index: %w", err)
	}

	// Create ParadeDB index for projects table
	projectIndexSQL := `CREATE INDEX IF NOT EXISTS idx_projects_paradedb ON projects USING bm25 (id, title, description, identifier) 
	WITH (
		key_field='id',
		text_fields='{
			"title": {"fast": true, "record": "freq"}, 
			"description": {"fast": true, "record": "freq"},
			"identifier": {"fast": true, "record": "freq"}
		}'
	)`
	if _, err := x.Exec(projectIndexSQL); err != nil {
		return fmt.Errorf("could not ensure paradedb project index: %w", err)
	}

	return nil
}
