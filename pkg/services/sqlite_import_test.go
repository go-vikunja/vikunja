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
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
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
	engine := getTestEngine()
	registry := NewServiceRegistry(engine)
	service := registry.SQLiteImport()

	// Create a test database with unique IDs to avoid conflicts with other tests
	tmpDB1 := createTestSQLiteDBForRollback(t, 5000) // Use high IDs (5000+)
	defer os.Remove(tmpDB1)

	// Get initial counts
	var initialUserCount int64
	_, err := engine.SQL("SELECT COUNT(*) FROM users").Get(&initialUserCount)
	require.NoError(t, err)

	// First import - should succeed
	opts1 := ImportOptions{
		SQLiteFile: tmpDB1,
		DryRun:     false,
		Quiet:      true,
	}

	report1, err := service.ImportFromSQLite(opts1)
	require.NoError(t, err)
	require.True(t, report1.Success)

	// Get counts after first import
	var countAfterFirst int64
	_, err = engine.SQL("SELECT COUNT(*) FROM users").Get(&countAfterFirst)
	require.NoError(t, err)
	assert.Equal(t, initialUserCount+1, countAfterFirst, "Should have one more user after first import")

	// Now try to import the same data again (will cause duplicate key violations)
	tmpDB2 := createTestSQLiteDBForRollback(t, 5000) // Same IDs - will conflict
	defer os.Remove(tmpDB2)

	opts2 := ImportOptions{
		SQLiteFile: tmpDB2,
		DryRun:     false,
		Quiet:      true,
	}

	// Import should fail due to duplicate key constraint violation
	report2, err := service.ImportFromSQLite(opts2)
	require.Error(t, err, "Import should fail due to duplicate key violation")
	assert.False(t, report2.Success, "Import should not be successful")
	assert.False(t, report2.DatabaseImported, "Database should not be marked as imported")

	// Verify database state is unchanged (transaction rolled back)
	var countAfterFailed int64
	_, err = engine.SQL("SELECT COUNT(*) FROM users").Get(&countAfterFailed)
	require.NoError(t, err)

	assert.Equal(t, countAfterFirst, countAfterFailed, "User count should be unchanged after rollback")

	// Verify the specific user from failed import wasn't duplicated
	var count int64
	count, err = engine.Where("username = ?", "testuser_rollback_5000").Count(&user.User{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "Should only have one instance of the test user (from first import)")

	// Verify error is reported
	assert.Greater(t, len(report2.Errors), 0, "Should have error messages")
	assert.Contains(t, report2.Errors[0], "failed to insert", "Error should mention insert failure")
	assert.Greater(t, report2.Duration, time.Duration(0), "Should have recorded duration")

	// Cleanup - remove the test user
	_, err = engine.Where("id = ?", 5000).Delete(&user.User{})
	require.NoError(t, err)
}

// createTestSQLiteDBForRollback creates a test SQLite database with a specific ID range
func createTestSQLiteDBForRollback(t *testing.T, baseID int) string {
	tmpFile := createEmptySQLiteDB(t)

	sqliteDB, err := sql.Open("sqlite3", tmpFile)
	require.NoError(t, err)
	defer sqliteDB.Close()

	now := time.Now()

	// Insert test user with specified ID
	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, password, email, name, created, updated, status)
		VALUES (?, ?, '$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.', ?, 'Test Rollback User', ?, ?, 0)
	`, baseID, fmt.Sprintf("testuser_rollback_%d", baseID), fmt.Sprintf("rollback_%d@example.com", baseID), now, now)
	require.NoError(t, err)

	return tmpFile
}

// TestSQLiteImportService_ProgressReporting tests progress reporting
func TestSQLiteImportService_ProgressReporting(t *testing.T) {
	// Setup
	engine := getTestEngine()
	registry := NewServiceRegistry(engine)
	service := registry.SQLiteImport()

	// Cleanup any existing data from previous tests
	_, _ = engine.Exec("DELETE FROM users WHERE id >= 6000")
	_, _ = engine.Exec("DELETE FROM projects WHERE id >= 6000")
	_, _ = engine.Exec("DELETE FROM tasks WHERE id >= 6000")

	// Create test SQLite database with enough records to trigger progress reporting
	dbFile := createTestSQLiteDBForProgress(t)
	defer os.Remove(dbFile)

	// Capture log output
	// Note: We can't easily capture log output in tests without modifying the logger,
	// so we'll verify the import completes successfully with counts
	opts := ImportOptions{
		SQLiteFile: dbFile,
		DryRun:     false,
		Quiet:      false, // Enable progress reporting
	}

	report, err := service.ImportFromSQLite(opts)

	// Verify import succeeded
	assert.NoError(t, err)
	assert.True(t, report.Success)
	assert.True(t, report.DatabaseImported)

	// Verify all entities were imported (progress was tracked)
	assert.Equal(t, int64(150), report.Counts.Users, "Should import 150 users")
	assert.Equal(t, int64(50), report.Counts.Projects, "Should import 50 projects")
	assert.Equal(t, int64(600), report.Counts.Tasks, "Should import 600 tasks")

	// Cleanup
	_, _ = engine.Exec("DELETE FROM users WHERE id >= 6000")
	_, _ = engine.Exec("DELETE FROM projects WHERE id >= 6000")
	_, _ = engine.Exec("DELETE FROM tasks WHERE id >= 6000")
}

// createTestSQLiteDBForProgress creates a test database with enough records for progress reporting
func createTestSQLiteDBForProgress(t *testing.T) string {
	tmpFile := createEmptySQLiteDB(t)

	db, err := sql.Open("sqlite3", tmpFile)
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	// Insert 150 users (will trigger progress at 100)
	for i := 6000; i < 6150; i++ {
		_, err = db.Exec(`
			INSERT INTO users (id, username, password, email, name, created, updated, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, i, fmt.Sprintf("user%d", i), "hashed", fmt.Sprintf("user%d@example.com", i),
			fmt.Sprintf("User %d", i), time.Now(), time.Now(), 0)
		if err != nil {
			t.Fatalf("Failed to insert user: %v", err)
		}
	}

	// Insert 50 projects (will trigger progress at 50)
	for i := 6000; i < 6050; i++ {
		_, err = db.Exec(`
			INSERT INTO projects (id, title, description, owner_id, created, updated, is_archived)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, i, fmt.Sprintf("Project %d", i), "Test project", 6000, time.Now(), time.Now(), false)
		if err != nil {
			t.Fatalf("Failed to insert project: %v", err)
		}
	}

	// Insert 600 tasks (will trigger progress at 500)
	for i := 6000; i < 6600; i++ {
		projectID := 6000 + (i % 50) // Distribute tasks across projects
		_, err = db.Exec(`
			INSERT INTO tasks (id, title, description, done, created_by_id, project_id, created, updated, "index")
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, i, fmt.Sprintf("Task %d", i), "Test task", false, 6000, projectID, time.Now(), time.Now(), i)
		if err != nil {
			t.Fatalf("Failed to insert task: %v", err)
		}
	}

	return tmpFile
}

// TestSQLiteImportService_FilesMigration tests file migration functionality
func TestSQLiteImportService_FilesMigration(t *testing.T) {
	// Setup
	engine := getTestEngine()
	registry := NewServiceRegistry(engine)
	service := registry.SQLiteImport()

	// Cleanup any existing data from previous tests
	_, _ = engine.Exec("DELETE FROM users WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM teams WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM projects WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM tasks WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM labels WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM files WHERE id IN (1, 2, 3)")

	// Create test SQLite database with file records
	sqliteFile := createTestSQLiteDBWithFiles(t)
	defer os.Remove(sqliteFile)

	// Create source files directory with test files
	sourceFilesDir := filepath.Join(t.TempDir(), "source_files")
	require.NoError(t, os.MkdirAll(sourceFilesDir, 0755))

	// Create test files with known content
	testFiles := map[int64]string{
		1: "This is test file 1 content",
		2: "Test file 2 with different content",
		3: "Another test file with more text",
	}

	for fileID, content := range testFiles {
		filePath := filepath.Join(sourceFilesDir, strconv.FormatInt(fileID, 10))
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))
	}

	// Create target files directory
	targetFilesDir := filepath.Join(t.TempDir(), "target_files")
	require.NoError(t, os.MkdirAll(targetFilesDir, 0755))

	// Override config for testing
	originalBasePath := config.FilesBasePath.GetString()
	config.FilesBasePath.Set(targetFilesDir)
	defer config.FilesBasePath.Set(originalBasePath)

	// Import with files
	report, err := service.ImportFromSQLite(ImportOptions{
		SQLiteFile: sqliteFile,
		FilesDir:   sourceFilesDir,
		DryRun:     false,
		Quiet:      true,
	})

	// Verify no error
	require.NoError(t, err)
	assert.True(t, report.Success)
	assert.True(t, report.DatabaseImported)
	assert.True(t, report.FilesMigrated)
	assert.Nil(t, report.FilesError)

	// Verify files were copied correctly
	for fileID, expectedContent := range testFiles {
		targetPath := filepath.Join(targetFilesDir, strconv.FormatInt(fileID, 10))

		// Check file exists
		assert.FileExists(t, targetPath, "File %d should exist", fileID)

		// Check content matches
		actualContent, err := os.ReadFile(targetPath)
		require.NoError(t, err, "Should be able to read file %d", fileID)
		assert.Equal(t, expectedContent, string(actualContent), "Content should match for file %d", fileID)
	}

	// Cleanup
	_, _ = engine.Exec("DELETE FROM files WHERE id IN (1, 2, 3)")
}

// TestSQLiteImportService_FilesMigration_MissingFiles tests graceful handling of missing files
func TestSQLiteImportService_FilesMigration_MissingFiles(t *testing.T) {
	// Setup
	engine := getTestEngine()
	registry := NewServiceRegistry(engine)
	service := registry.SQLiteImport()

	// Cleanup any existing data from previous tests
	_, _ = engine.Exec("DELETE FROM users WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM teams WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM projects WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM tasks WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM labels WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM files WHERE id IN (1, 2, 3)")

	// Create test SQLite database with file records
	sqliteFile := createTestSQLiteDBWithFiles(t)
	defer os.Remove(sqliteFile)

	// Create source files directory but don't create all files (simulate missing files)
	sourceFilesDir := filepath.Join(t.TempDir(), "source_files_missing")
	require.NoError(t, os.MkdirAll(sourceFilesDir, 0755))

	// Only create file 1, leave files 2 and 3 missing
	filePath := filepath.Join(sourceFilesDir, "1")
	require.NoError(t, os.WriteFile(filePath, []byte("Test file 1"), 0644))

	// Create target files directory
	targetFilesDir := filepath.Join(t.TempDir(), "target_files_missing")
	require.NoError(t, os.MkdirAll(targetFilesDir, 0755))

	// Override config for testing
	originalBasePath := config.FilesBasePath.GetString()
	config.FilesBasePath.Set(targetFilesDir)
	defer config.FilesBasePath.Set(originalBasePath)

	// Import with missing files - should not fail
	report, err := service.ImportFromSQLite(ImportOptions{
		SQLiteFile: sqliteFile,
		FilesDir:   sourceFilesDir,
		DryRun:     false,
		Quiet:      true,
	})

	// Verify no error - missing files should be reported but not block import
	require.NoError(t, err)
	assert.True(t, report.Success)
	assert.True(t, report.DatabaseImported)

	// Verify file 1 was copied
	targetPath1 := filepath.Join(targetFilesDir, "1")
	assert.FileExists(t, targetPath1, "File 1 should exist")

	// Verify files 2 and 3 don't exist (they were missing)
	targetPath2 := filepath.Join(targetFilesDir, "2")
	targetPath3 := filepath.Join(targetFilesDir, "3")
	assert.NoFileExists(t, targetPath2, "File 2 should not exist (was missing)")
	assert.NoFileExists(t, targetPath3, "File 3 should not exist (was missing)")

	// Cleanup
	_, _ = engine.Exec("DELETE FROM files WHERE id IN (1, 2, 3)")
}

// TestSQLiteImportService_FilesMigration_NoFilesDir tests import without files directory
func TestSQLiteImportService_FilesMigration_NoFilesDir(t *testing.T) {
	// Setup
	engine := getTestEngine()
	registry := NewServiceRegistry(engine)
	service := registry.SQLiteImport()

	// Cleanup any existing data from previous tests
	_, _ = engine.Exec("DELETE FROM users WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM teams WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM projects WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM tasks WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM labels WHERE id >= 1000")
	_, _ = engine.Exec("DELETE FROM files WHERE id IN (1, 2, 3)")

	// Create test SQLite database
	sqliteFile := createTestSQLiteDBWithFiles(t)
	defer os.Remove(sqliteFile)

	// Import without files directory - should succeed, just skip files
	report, err := service.ImportFromSQLite(ImportOptions{
		SQLiteFile: sqliteFile,
		FilesDir:   "", // No files directory
		DryRun:     false,
		Quiet:      true,
	})

	// Verify success
	require.NoError(t, err)
	assert.True(t, report.Success)
	assert.True(t, report.DatabaseImported)
	// FilesMigrated should be false since no files dir was provided
	assert.False(t, report.FilesMigrated)

	// Cleanup
	_, _ = engine.Exec("DELETE FROM files WHERE id IN (1, 2, 3)")
}

// Helper functions

// createTestSQLiteDBWithFiles creates a test SQLite database with file records
func createTestSQLiteDBWithFiles(t *testing.T) string {
	// Use the existing helper to create a full test database
	tmpFile := createTestSQLiteDB(t)

	sqliteDB, err := sql.Open("sqlite3", tmpFile)
	require.NoError(t, err)
	defer sqliteDB.Close()

	now := time.Now().Unix()

	// Insert test file records (using user ID 1000 from createTestSQLiteDB)
	_, err = sqliteDB.Exec(`
		INSERT INTO files (id, name, mime, size, created_by_id, created)
		VALUES 
			(1, 'test_file_1.txt', 'text/plain', 28, 1000, ?),
			(2, 'test_file_2.txt', 'text/plain', 35, 1000, ?),
			(3, 'test_file_3.txt', 'text/plain', 36, 1000, ?)
	`, now, now, now)
	require.NoError(t, err)

	return tmpFile
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

		CREATE TABLE files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			mime TEXT,
			size INTEGER NOT NULL,
			created_by_id INTEGER NOT NULL,
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
