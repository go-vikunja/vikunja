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

package services

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// TestSQLiteImportService_InvalidFile tests error handling for invalid SQLite files
func TestSQLiteImportService_InvalidFile(t *testing.T) {
	engine := getTestEngine()
	registry := NewServiceRegistry(engine)
	service := registry.SQLiteImport()

	opts := ImportOptions{
		SQLiteFile: "/non/existent/file.db",
		DryRun:     true,
		Quiet:      true,
	}

	report, err := service.ImportFromSQLite(opts)
	require.Error(t, err)
	assert.False(t, report.Success)
	assert.Contains(t, err.Error(), "cannot access SQLite file")
}

// TestSQLiteImportService_DryRun tests dry-run mode
func TestSQLiteImportService_DryRun(t *testing.T) {
	engine := getTestEngine()
	registry := NewServiceRegistry(engine)
	service := registry.SQLiteImport()

	// Create a temporary SQLite database with test data
	tmpDB := createTestSQLiteDB(t)
	defer os.Remove(tmpDB)

	// Get initial user count
	var initialCount int64
	_, err := engine.SQL("SELECT COUNT(*) FROM users").Get(&initialCount)
	require.NoError(t, err)

	opts := ImportOptions{
		SQLiteFile: tmpDB,
		DryRun:     true,
		Quiet:      true,
	}

	report, err := service.ImportFromSQLite(opts)
	require.NoError(t, err)
	assert.True(t, report.Success)
	assert.False(t, report.DatabaseImported, "Database should not be imported in dry-run mode")

	// Verify no additional data was inserted
	var finalCount int64
	_, err = engine.SQL("SELECT COUNT(*) FROM users").Get(&finalCount)
	require.NoError(t, err)
	assert.Equal(t, initialCount, finalCount, "No users should be inserted in dry-run mode")
}

// TestSQLiteImportService_EmptyDatabase tests importing from an empty database
func TestSQLiteImportService_EmptyDatabase(t *testing.T) {
	engine := getTestEngine()
	registry := NewServiceRegistry(engine)
	service := registry.SQLiteImport()

	// Create an empty SQLite database
	tmpDB := createEmptySQLiteDB(t)
	defer os.Remove(tmpDB)

	opts := ImportOptions{
		SQLiteFile: tmpDB,
		DryRun:     false,
		Quiet:      true,
	}

	report, err := service.ImportFromSQLite(opts)
	require.NoError(t, err)
	assert.True(t, report.Success)
	assert.Equal(t, int64(0), report.Counts.Users)
	assert.Equal(t, int64(0), report.Counts.Projects)
	assert.Equal(t, int64(0), report.Counts.Tasks)
}

// TestSQLiteImportService_BasicImport tests importing basic data
func TestSQLiteImportService_BasicImport(t *testing.T) {
	engine := getTestEngine()
	registry := NewServiceRegistry(engine)
	service := registry.SQLiteImport()

	// Create a test SQLite database with sample data
	tmpDB := createTestSQLiteDB(t)
	defer os.Remove(tmpDB)

	// Get initial user count (test fixtures may already have users)
	var initialCount int64
	_, err := engine.SQL("SELECT COUNT(*) FROM users").Get(&initialCount)
	require.NoError(t, err)

	opts := ImportOptions{
		SQLiteFile: tmpDB,
		DryRun:     false,
		Quiet:      true,
	}

	report, err := service.ImportFromSQLite(opts)
	require.NoError(t, err)
	assert.True(t, report.Success)
	assert.True(t, report.DatabaseImported)
	assert.Greater(t, report.Counts.Users, int64(0))
	assert.Greater(t, report.Duration, time.Duration(0))

	// Verify data was actually imported
	var importedUser user.User
	exists, err := engine.Where("username = ?", "testuser_import").Get(&importedUser)
	require.NoError(t, err)
	assert.True(t, exists, "Test user should be imported")
	assert.Equal(t, "testuser_import", importedUser.Username)
	assert.Equal(t, "test_import@example.com", importedUser.Email)
}

// TestSQLiteImportService_TransactionRollback tests transaction rollback on error
func TestSQLiteImportService_TransactionRollback(t *testing.T) {
	t.Skip("TODO: Implement in T003 (Transaction Management)")
	// This test will verify that database state is unchanged after a failed import
}

// TestSQLiteImportService_ProgressReporting tests progress reporting
func TestSQLiteImportService_ProgressReporting(t *testing.T) {
	t.Skip("TODO: Implement in T005 (Progress Reporting)")
	// This test will verify that progress is reported correctly
}

// Helper functions

// createEmptySQLiteDB creates an empty SQLite database with minimal test schema
// NOTE: This duplicates schema definitions for testing purposes. In production, the actual
// schema is managed through migrations (pkg/migration/). This test schema includes only
// the essential tables needed for import testing.
func createEmptySQLiteDB(t *testing.T) string {
	tmpFile := filepath.Join(t.TempDir(), "empty_test.db")

	sqliteDB, err := sql.Open("sqlite3", tmpFile)
	require.NoError(t, err)
	defer sqliteDB.Close()

	// Create minimal test schema - matches core structure from models
	// This is intentionally simplified; real schema is richer and managed by migrations
	schema := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password TEXT,
			email TEXT,
			name TEXT,
			created DATETIME NOT NULL,
			updated DATETIME NOT NULL,
			status INTEGER DEFAULT 0,
			avatar_provider TEXT,
			language TEXT,
			timezone TEXT,
			week_start INTEGER,
			default_project_id INTEGER,
			overdue_tasks_reminders_time TEXT DEFAULT '09:00',
			overdue_tasks_reminders_enabled INTEGER DEFAULT 1
		);

		CREATE TABLE teams (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			created DATETIME NOT NULL,
			updated DATETIME NOT NULL,
			created_by_id INTEGER NOT NULL
		);

		CREATE TABLE team_members (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			team_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			admin INTEGER DEFAULT 0,
			created DATETIME NOT NULL
		);

		CREATE TABLE projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			owner_id INTEGER NOT NULL,
			identifier TEXT,
			hex_color TEXT,
			is_archived INTEGER DEFAULT 0,
			background_information TEXT,
			created DATETIME NOT NULL,
			updated DATETIME NOT NULL,
			parent_project_id INTEGER,
			position REAL
		);

		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			done INTEGER DEFAULT 0,
			done_at DATETIME,
			due_date DATETIME,
			created_by_id INTEGER NOT NULL,
			project_id INTEGER NOT NULL,
			repeat_after INTEGER,
			repeat_mode INTEGER DEFAULT 0,
			priority INTEGER,
			start_date DATETIME,
			end_date DATETIME,
			hex_color TEXT,
			percent_done REAL,
			identifier TEXT,
			"index" INTEGER DEFAULT 0,
			uid TEXT,
			cover_image_attachment_id INTEGER,
			created DATETIME NOT NULL,
			updated DATETIME NOT NULL,
			bucket_id INTEGER,
			position REAL,
			reminder_dates TEXT
		);

		CREATE TABLE labels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			hex_color TEXT,
			created_by_id INTEGER NOT NULL,
			created DATETIME NOT NULL,
			updated DATETIME NOT NULL
		);

		CREATE TABLE task_labels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			label_id INTEGER NOT NULL,
			created DATETIME NOT NULL
		);
	`

	_, err = sqliteDB.Exec(schema)
	require.NoError(t, err, "Failed to create test SQLite schema")

	return tmpFile
}

// createTestSQLiteDB creates a SQLite database with test data
func createTestSQLiteDB(t *testing.T) string {
	tmpFile := createEmptySQLiteDB(t)

	sqliteDB, err := sql.Open("sqlite3", tmpFile)
	require.NoError(t, err)
	defer sqliteDB.Close()

	now := time.Now()

	// Use high IDs (1000+) to avoid conflicts with test fixtures
	// Insert test user
	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, password, email, name, created, updated, status)
		VALUES (1000, 'testuser_import', '$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.', 'test_import@example.com', 'Test Import User', ?, ?, 0)
	`, now, now)
	require.NoError(t, err)

	// Insert test team
	_, err = sqliteDB.Exec(`
		INSERT INTO teams (id, name, description, created, updated, created_by_id)
		VALUES (1000, 'Test Import Team', 'A test team for import', ?, ?, 1000)
	`, now, now)
	require.NoError(t, err)

	// Insert test project
	_, err = sqliteDB.Exec(`
		INSERT INTO projects (id, title, description, owner_id, identifier, created, updated)
		VALUES (1000, 'Test Import Project', 'A test project for import', 1000, 'TESTIMPORT', ?, ?)
	`, now, now)
	require.NoError(t, err)

	// Insert test task
	_, err = sqliteDB.Exec(`
		INSERT INTO tasks (id, title, description, created_by_id, project_id, created, updated)
		VALUES (1000, 'Test Import Task', 'A test task for import', 1000, 1000, ?, ?)
	`, now, now)
	require.NoError(t, err)

	// Insert test label
	_, err = sqliteDB.Exec(`
		INSERT INTO labels (id, title, description, hex_color, created_by_id, created, updated)
		VALUES (1000, 'Test Import Label', 'A test label for import', 'FF0000', 1000, ?, ?)
	`, now, now)
	require.NoError(t, err)

	return tmpFile
}

// getTestEngine returns the test database engine from the services package
func getTestEngine() *xorm.Engine {
	// Use the main test database from the services package
	return testEngine
}

// cleanDatabase is no longer needed - removed in favor of accepting test fixtures
// The test database already contains fixtures which is fine for our tests.
// We verify import results by checking for our specific test data, not by requiring
// an empty database.
