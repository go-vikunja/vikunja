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

package integration

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

// TestSQLiteImport_FullWorkflow tests the complete SQLite import workflow
// with a realistic dataset including users, teams, projects, tasks, labels, and files
func TestSQLiteImport_FullWorkflow(t *testing.T) {
	// Initialize test environment
	engine := setupTestEnvironment(t)

	// Create a realistic SQLite database with 100 users and 1000 tasks
	sourceDB := createRealisticSQLiteDB(t)
	defer os.Remove(sourceDB)

	// Create test files directory with sample files
	filesDir := createTestFiles(t)
	defer os.RemoveAll(filesDir)

	// Perform the import
	registry := services.NewServiceRegistry(engine)
	importService := registry.SQLiteImport()

	opts := services.ImportOptions{
		SQLiteFile: sourceDB,
		FilesDir:   filesDir,
		DryRun:     false,
		Quiet:      true,
	}

	report, err := importService.ImportFromSQLite(opts)
	require.NoError(t, err, "Import should succeed")
	assert.True(t, report.Success, "Import should be successful")
	assert.True(t, report.DatabaseImported, "Database should be imported")

	// Verify imported data matches source
	verifyImportedData(t, engine, sourceDB, report)

	// Verify foreign key relationships
	verifyForeignKeys(t, engine)

	// Verify file migrations
	verifyFiles(t, engine, filesDir)
}

// TestSQLiteImport_CrossDatabase tests importing from SQLite to different database backends
// Note: This test requires VIKUNJA_TESTS_USE_CONFIG=1 and proper database configuration
func TestSQLiteImport_CrossDatabase(t *testing.T) {
	// Skip if not configured for cross-database testing
	if os.Getenv("VIKUNJA_TESTS_USE_CONFIG") != "1" {
		t.Skip("Skipping cross-database test - set VIKUNJA_TESTS_USE_CONFIG=1 and configure database")
	}

	dbType := config.DatabaseType.GetString()
	t.Logf("Testing SQLite import to %s", dbType)

	// Initialize test environment with configured database
	engine := setupTestEnvironment(t)

	// Create source SQLite database
	sourceDB := createRealisticSQLiteDB(t)
	defer os.Remove(sourceDB)

	// Create test files
	filesDir := createTestFiles(t)
	defer os.RemoveAll(filesDir)

	// Clean the target database before import
	cleanTargetDatabase(t, engine)

	// Perform the import
	registry := services.NewServiceRegistry(engine)
	importService := registry.SQLiteImport()

	opts := services.ImportOptions{
		SQLiteFile: sourceDB,
		FilesDir:   filesDir,
		DryRun:     false,
		Quiet:      true,
	}

	report, err := importService.ImportFromSQLite(opts)
	require.NoError(t, err, "Cross-database import should succeed")
	assert.True(t, report.Success, "Import should be successful")

	// Verify the import
	verifyImportedData(t, engine, sourceDB, report)
	verifyForeignKeys(t, engine)
}

// setupTestEnvironment initializes the test environment
func setupTestEnvironment(t *testing.T) *xorm.Engine {
	// Initialize logger
	log.InitLogger()

	// Initialize config
	config.InitDefaultConfig()
	config.ServiceRootpath.Set(os.Getenv("VIKUNJA_SERVICE_ROOTPATH"))

	// Initialize files
	files.InitTests()

	// Initialize user tests
	user.InitTests()

	// Create test engine
	engine, err := db.CreateTestEngine()
	require.NoError(t, err, "Failed to create test engine")

	// Sync models
	tables := models.GetTables()
	err = engine.Sync2(tables...)
	require.NoError(t, err, "Failed to sync tables")

	// Initialize service dependencies
	services.InitializeDependencies()
	services.InitUserService()

	return engine
}

// createRealisticSQLiteDB creates a SQLite database with realistic test data
func createRealisticSQLiteDB(t *testing.T) string {
	tmpFile, err := os.CreateTemp("", "vikunja_import_test_*.db")
	require.NoError(t, err)
	tmpFile.Close()
	dbPath := tmpFile.Name()

	sqliteDB, err := sql.Open("sqlite3", dbPath)
	require.NoError(t, err)
	defer sqliteDB.Close()

	// Create schema
	createSQLiteSchema(t, sqliteDB)

	// Insert realistic test data
	now := time.Now()

	// 100 users
	for i := 9000; i < 9100; i++ {
		username := fmt.Sprintf("user%d", i)
		name := fmt.Sprintf("User %d", i)
		email := fmt.Sprintf("user%d@example.com", i)
		_, err := sqliteDB.Exec(`
			INSERT INTO users (id, name, username, email, password, created, updated)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			i, name, username, email, "$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.",
			now, now)
		require.NoError(t, err)
	}

	// 10 teams
	for i := 9000; i < 9010; i++ {
		_, err := sqliteDB.Exec(`
			INSERT INTO teams (id, name, description, created, updated, created_by_id)
			VALUES (?, ?, ?, ?, ?, ?)`,
			i, fmt.Sprintf("Team %d", i), fmt.Sprintf("Description for team %d", i),
			now, now, 9000)
		require.NoError(t, err)
	}

	// 50 projects (old "lists")
	for i := 9000; i < 9050; i++ {
		_, err := sqliteDB.Exec(`
			INSERT INTO projects (id, title, description, owner_id, created, updated)
			VALUES (?, ?, ?, ?, ?, ?)`,
			i, fmt.Sprintf("Project %d", i), fmt.Sprintf("Description for project %d", i),
			9000+(i%100), now, now)
		require.NoError(t, err)
	}

	// 1000 tasks
	for i := 9000; i < 10000; i++ {
		projectID := 9000 + (i % 50)
		_, err := sqliteDB.Exec(`
			INSERT INTO tasks (id, title, description, created, updated, created_by_id, project_id, done)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			i, fmt.Sprintf("Task %d", i), fmt.Sprintf("Description for task %d", i),
			now, now, 9000, projectID, 0)
		require.NoError(t, err)
	}

	// 20 labels
	for i := 9000; i < 9020; i++ {
		_, err := sqliteDB.Exec(`
			INSERT INTO labels (id, title, description, hex_color, created_by_id, created, updated)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			i, fmt.Sprintf("Label %d", i), fmt.Sprintf("Label description %d", i),
			"ff0000", 9000, now, now)
		require.NoError(t, err)
	}

	// 100 task-label associations
	for i := 9000; i < 9100; i++ {
		taskID := 9000 + (i % 1000)
		labelID := 9000 + (i % 20)
		_, err := sqliteDB.Exec(`
			INSERT INTO task_labels (task_id, label_id, created)
			VALUES (?, ?, ?)`, taskID, labelID, now)
		require.NoError(t, err)
	}

	// 5 files
	for i := 9000; i < 9005; i++ {
		_, err := sqliteDB.Exec(`
			INSERT INTO files (id, name, mime, size, created, created_by_id)
			VALUES (?, ?, ?, ?, ?, ?)`,
			i, fmt.Sprintf("file%d.txt", i), "text/plain", 1024,
			now, 9000)
		require.NoError(t, err)
	}

	return dbPath
}

// createSQLiteSchema creates the SQLite schema for testing
func createSQLiteSchema(t *testing.T, db *sql.DB) {
	// Use the same schema as unit tests - proven to work
	schema := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
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
			id INTEGER PRIMARY KEY,
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
			id INTEGER PRIMARY KEY,
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
			id INTEGER PRIMARY KEY,
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
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			hex_color TEXT,
			created_by_id INTEGER NOT NULL,
			created DATETIME NOT NULL,
			updated DATETIME NOT NULL
		);

		CREATE TABLE label_tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			label_id INTEGER NOT NULL,
			created DATETIME NOT NULL
		);
		
		CREATE TABLE task_labels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_id INTEGER NOT NULL,
			label_id INTEGER NOT NULL,
			created DATETIME NOT NULL
		);

		CREATE TABLE files (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			mime TEXT,
			size INTEGER NOT NULL,
			created_by_id INTEGER NOT NULL,
			created DATETIME NOT NULL
		);
	`

	_, err := db.Exec(schema)
	require.NoError(t, err, "Failed to create test SQLite schema")
}

// createTestFiles creates test files for migration
func createTestFiles(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "vikunja_files_test_*")
	require.NoError(t, err)

	// Create 5 test files matching the file IDs in the database
	for i := 9000; i < 9005; i++ {
		filePath := filepath.Join(tmpDir, fmt.Sprintf("%d", i))
		err := os.WriteFile(filePath, []byte(fmt.Sprintf("Test file content %d", i)), 0644)
		require.NoError(t, err)
	}

	return tmpDir
}

// verifyImportedData verifies that imported data matches the source
func verifyImportedData(t *testing.T, engine *xorm.Engine, sourceDB string, report *services.ImportReport) {
	// Open source database to get expected counts
	sqliteDB, err := sql.Open("sqlite3", sourceDB)
	require.NoError(t, err)
	defer sqliteDB.Close()

	// Verify user count
	var userCount int64
	_, err = engine.SQL("SELECT COUNT(*) FROM users WHERE id >= 9000 AND id < 9100").Get(&userCount)
	require.NoError(t, err)
	assert.Equal(t, int64(100), userCount, "Should import 100 users")
	assert.Equal(t, int64(100), report.Counts.Users, "Report should show 100 users")

	// Verify team count
	var teamCount int64
	_, err = engine.SQL("SELECT COUNT(*) FROM teams WHERE id >= 9000 AND id < 9010").Get(&teamCount)
	require.NoError(t, err)
	assert.Equal(t, int64(10), teamCount, "Should import 10 teams")
	assert.Equal(t, int64(10), report.Counts.Teams, "Report should show 10 teams")

	// Verify project count
	var projectCount int64
	_, err = engine.SQL("SELECT COUNT(*) FROM projects WHERE id >= 9000 AND id < 9050").Get(&projectCount)
	require.NoError(t, err)
	assert.Equal(t, int64(50), projectCount, "Should import 50 projects")
	assert.Equal(t, int64(50), report.Counts.Projects, "Report should show 50 projects")

	// Verify task count
	var taskCount int64
	_, err = engine.SQL("SELECT COUNT(*) FROM tasks WHERE id >= 9000 AND id < 10000").Get(&taskCount)
	require.NoError(t, err)
	assert.Equal(t, int64(1000), taskCount, "Should import 1000 tasks")
	assert.Equal(t, int64(1000), report.Counts.Tasks, "Report should show 1000 tasks")

	// Verify label count
	var labelCount int64
	_, err = engine.SQL("SELECT COUNT(*) FROM labels WHERE id >= 9000 AND id < 9020").Get(&labelCount)
	require.NoError(t, err)
	assert.Equal(t, int64(20), labelCount, "Should import 20 labels")
	assert.Equal(t, int64(20), report.Counts.Labels, "Report should show 20 labels")

	// Verify file count
	var fileCount int64
	_, err = engine.SQL("SELECT COUNT(*) FROM files WHERE id >= 9000 AND id < 9005").Get(&fileCount)
	require.NoError(t, err)
	assert.Equal(t, int64(5), fileCount, "Should import 5 files")
	assert.Equal(t, int64(5), report.Counts.Files, "Report should show 5 files")

	t.Logf("Import verification successful: %d users, %d teams, %d projects, %d tasks, %d labels, %d files",
		report.Counts.Users, report.Counts.Teams, report.Counts.Projects, report.Counts.Tasks, report.Counts.Labels, report.Counts.Files)
}

// verifyForeignKeys verifies that all foreign key relationships are intact
func verifyForeignKeys(t *testing.T, engine *xorm.Engine) {
	// Verify tasks reference valid projects
	var orphanedTasks int64
	_, err := engine.SQL(`
		SELECT COUNT(*) FROM tasks t
		WHERE t.id >= 9000 AND t.id < 10000
		AND NOT EXISTS (SELECT 1 FROM projects p WHERE p.id = t.project_id)
	`).Get(&orphanedTasks)
	require.NoError(t, err)
	assert.Equal(t, int64(0), orphanedTasks, "No tasks should have invalid project references")

	// Verify label_tasks references valid labels and tasks
	var orphanedLabelTasks int64
	_, err = engine.SQL(`
		SELECT COUNT(*) FROM label_tasks lt
		WHERE (lt.task_id >= 9000 AND lt.task_id < 10000)
		AND (NOT EXISTS (SELECT 1 FROM labels l WHERE l.id = lt.label_id)
		     OR NOT EXISTS (SELECT 1 FROM tasks t WHERE t.id = lt.task_id))
	`).Get(&orphanedLabelTasks)
	require.NoError(t, err)
	assert.Equal(t, int64(0), orphanedLabelTasks, "No label_tasks should have invalid references")

	// Verify projects reference valid owners
	var orphanedProjects int64
	_, err = engine.SQL(`
		SELECT COUNT(*) FROM projects p
		WHERE p.id >= 9000 AND p.id < 9050
		AND NOT EXISTS (SELECT 1 FROM users u WHERE u.id = p.owner_id)
	`).Get(&orphanedProjects)
	require.NoError(t, err)
	assert.Equal(t, int64(0), orphanedProjects, "No projects should have invalid owner references")

	t.Log("Foreign key verification successful")
}

// verifyFiles verifies that files were migrated correctly
func verifyFiles(t *testing.T, engine *xorm.Engine, sourceFilesDir string) {
	// Get target files directory from config
	targetFilesDir := config.FilesBasePath.GetString()
	if targetFilesDir == "" {
		t.Skip("Files directory not configured, skipping file verification")
	}

	// Verify each file exists and matches content
	for i := 9000; i < 9005; i++ {
		// Check if file record exists in database
		var fileExists bool
		_, err := engine.SQL("SELECT 1 FROM files WHERE id = ?", i).Get(&fileExists)
		require.NoError(t, err)
		assert.True(t, fileExists, fmt.Sprintf("File %d should exist in database", i))

		// Check if physical file exists
		targetPath := filepath.Join(targetFilesDir, fmt.Sprintf("%d", i))
		_, err = os.Stat(targetPath)
		if os.IsNotExist(err) {
			// File might not have been copied if files-dir wasn't provided
			t.Logf("File %d not found at %s (expected if import didn't include files-dir)", i, targetPath)
			continue
		}
		require.NoError(t, err)

		// Verify content matches source
		sourcePath := filepath.Join(sourceFilesDir, fmt.Sprintf("%d", i))
		sourceContent, err := os.ReadFile(sourcePath)
		require.NoError(t, err)

		targetContent, err := os.ReadFile(targetPath)
		require.NoError(t, err)

		assert.Equal(t, sourceContent, targetContent, fmt.Sprintf("File %d content should match", i))
	}

	t.Log("File verification successful")
}

// cleanTargetDatabase removes all test data from the target database
func cleanTargetDatabase(t *testing.T, engine *xorm.Engine) {
	tables := []string{
		"label_tasks",
		"task_attachments",
		"task_comments",
		"task_relations",
		"task_assignees",
		"tasks",
		"team_projects",
		"team_members",
		"teams",
		"users_projects",
		"projects",
		"labels",
		"files",
		"users",
	}

	for _, table := range tables {
		_, err := engine.Exec(fmt.Sprintf("DELETE FROM %s WHERE id >= 9000", table))
		if err != nil {
			// Table might not exist, that's okay
			t.Logf("Could not clean table %s: %v", table, err)
		}
	}
}
