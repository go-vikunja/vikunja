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

package cmd

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestImportDBCommand_Flags(t *testing.T) {
	// Test that all required flags are defined
	assert.NotNil(t, importDBCmd.Flags().Lookup("sqlite-file"))
	assert.NotNil(t, importDBCmd.Flags().Lookup("files-dir"))
	assert.NotNil(t, importDBCmd.Flags().Lookup("dry-run"))
	assert.NotNil(t, importDBCmd.Flags().Lookup("quiet"))
}

func TestImportDBCommand_Help(t *testing.T) {
	// Test that help text is defined
	assert.NotEmpty(t, importDBCmd.Short)
	assert.NotEmpty(t, importDBCmd.Long)
	assert.Contains(t, importDBCmd.Long, "primary migration path")
	assert.Contains(t, importDBCmd.Long, "Examples:")
}

func TestImportDBCommand_RequiredFlags(t *testing.T) {
	// Test that sqlite-file flag is defined
	flag := importDBCmd.Flags().Lookup("sqlite-file")
	assert.NotNil(t, flag)
}

// createTestSQLiteDB creates a minimal SQLite database for testing
func createTestSQLiteDB(t *testing.T) string {
	tmpFile, err := os.CreateTemp("", "vikunja-cli-test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	dbPath := tmpFile.Name()
	tmpFile.Close()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open SQLite database: %v", err)
	}
	defer db.Close()

	// Create minimal schema
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
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Insert one test user
	now := time.Now()
	_, err = db.Exec(`
		INSERT INTO users (id, username, password, email, name, created, updated, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, 9999, "cli_test", "hash", "cli@test.com", "CLI Test", now, now, 0)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	return dbPath
}
