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
	"runtime"
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

// DatabasePathConfig holds configuration for database path resolution.
// This struct allows the path resolution logic to be tested independently
// of the global config package.
type DatabasePathConfig struct {
	ConfiguredPath string // The database.path config value
	RootPath       string // The service.rootpath config value
	ExecutablePath string // Directory of the executable binary
}

// resolveDatabasePath resolves a database path configuration to an absolute path.
//
// Resolution rules:
//  1. If ConfiguredPath is "memory", returns "memory" (special case for in-memory DB)
//  2. If ConfiguredPath is already absolute, returns it as-is (cleaned)
//  3. If ConfiguredPath is relative:
//     a. If RootPath differs from ExecutablePath (explicitly configured),
//     joins with RootPath
//     b. Otherwise, joins with platform-specific user data directory
//
// The getUserDataDir parameter allows injecting a mock for testing.
func resolveDatabasePath(cfg DatabasePathConfig, getUserDataDir func() (string, error)) (string, error) {
	if cfg.ConfiguredPath == "memory" {
		return "memory", nil
	}

	var path string

	switch {
	case filepath.IsAbs(cfg.ConfiguredPath):
		path = filepath.Clean(cfg.ConfiguredPath)
	case cfg.RootPath != cfg.ExecutablePath:
		path = filepath.Join(cfg.RootPath, cfg.ConfiguredPath)
	default:
		dataDir, err := getUserDataDir()
		if err != nil {
			log.Warningf("Could not get user data directory, falling back to rootpath: %v", err)
			path = filepath.Join(cfg.RootPath, cfg.ConfiguredPath)
		} else {
			path = filepath.Join(dataDir, cfg.ConfiguredPath)
		}
	}

	return filepath.Abs(path)
}

func initSqliteEngine() (engine *xorm.Engine, err error) {
	rootPath := config.ServiceRootpath.GetString()

	executablePath := rootPath
	if execPath, err := os.Executable(); err == nil {
		executablePath = filepath.Dir(execPath)
	}

	cfg := DatabasePathConfig{
		ConfiguredPath: config.DatabasePath.GetString(),
		RootPath:       rootPath,
		ExecutablePath: executablePath,
	}

	path, err := resolveDatabasePath(cfg, getUserDataDir)
	if err != nil {
		return nil, fmt.Errorf("could not resolve database path: %w", err)
	}

	if path == "memory" {
		return xorm.NewEngine("sqlite3", "file::memory:?cache=shared")
	}

	// Log the resolved database path
	log.Infof("Using SQLite database at: %s", path)

	// Warn if the database is in a potentially problematic location
	if isSystemDirectory(path) {
		log.Warningf("Database path (%s) appears to be in a system directory. This may cause issues. Please use an absolute path or configure the database path to a user data directory.", path)
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

// getUserDataDir returns the platform-appropriate directory for application data
func getUserDataDir() (string, error) {
	var dataDir string

	switch runtime.GOOS {
	case "windows":
		// On Windows, use %LOCALAPPDATA%\Vikunja
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			// Fallback to %USERPROFILE%\AppData\Local if LOCALAPPDATA is not set
			userProfile := os.Getenv("USERPROFILE")
			if userProfile == "" {
				return "", fmt.Errorf("neither LOCALAPPDATA nor USERPROFILE environment variables are set")
			}
			localAppData = filepath.Join(userProfile, "AppData", "Local")
		}
		dataDir = filepath.Join(localAppData, "Vikunja")
	case "darwin":
		// On macOS, use ~/Library/Application Support/Vikunja
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataDir = filepath.Join(home, "Library", "Application Support", "Vikunja")
	default:
		// On Linux and other Unix-like systems, use XDG_DATA_HOME or ~/.local/share/vikunja
		xdgDataHome := os.Getenv("XDG_DATA_HOME")
		if xdgDataHome != "" {
			dataDir = filepath.Join(xdgDataHome, "vikunja")
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			dataDir = filepath.Join(home, ".local", "share", "vikunja")
		}
	}

	// Ensure the directory exists
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return "", fmt.Errorf("could not create data directory %s: %w", dataDir, err)
	}

	return dataDir, nil
}

// isSystemDirectory checks if a path appears to be in a system directory
// where users should not typically store application data
func isSystemDirectory(path string) bool {
	// Clean and normalize the path
	path = filepath.Clean(path)
	lowerPath := strings.ToLower(path)

	// Windows system directories
	if runtime.GOOS == "windows" {
		// Convert to absolute path if possible for more accurate checking
		absPath := lowerPath
		if abs, err := filepath.Abs(path); err == nil {
			absPath = strings.ToLower(filepath.Clean(abs))
		}

		// Check common Windows system directories using prefix matching
		// This prevents false positives like C:\myapp\windows\data
		windowsSystemPrefixes := []string{
			"c:\\windows\\system32",
			"c:\\windows\\syswow64",
			"c:\\windows\\winsxs",
			"c:\\windows\\servicing",
		}

		for _, prefix := range windowsSystemPrefixes {
			if strings.HasPrefix(absPath, prefix) {
				return true
			}
		}

		// Also check for direct C:\Windows (not subdirectories like C:\myapp\windows)
		// by ensuring it starts with the drive and windows directory
		if absPath == "c:\\windows" || strings.HasPrefix(absPath, "c:\\windows\\") {
			// Exclude some safe subdirectories under C:\Windows
			safeDirs := []string{
				"c:\\windows\\temp",
			}
			for _, safeDir := range safeDirs {
				if strings.HasPrefix(absPath, safeDir) {
					return false
				}
			}
			return true
		}
	}

	// Unix-like system directories - use prefix matching
	systemDirs := []string{
		"/bin", "/sbin", "/usr/bin", "/usr/sbin",
		"/etc", "/sys", "/proc", "/dev",
	}
	for _, sysDir := range systemDirs {
		// Ensure we match exact directory boundaries
		if lowerPath == sysDir || strings.HasPrefix(lowerPath, sysDir+"/") {
			return true
		}
	}

	return false
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

// NewSession creates a new xorm session with an active transaction.
// The caller must call s.Commit() on success or s.Rollback() on error.
// s.Close() will auto-rollback any uncommitted transaction.
func NewSession() *xorm.Session {
	s := x.NewSession()
	_ = s.Begin()
	return s
}

// Type returns the db type of the currently configured db
func Type() schemas.DBType {
	return x.Dialect().URI().DBType
}

func GetDialect() string {
	switch config.DatabaseType.GetString() {
	case "mysql":
		return builder.MYSQL
	case "postgres":
		return builder.POSTGRES
	default:
		return builder.SQLITE
	}
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
