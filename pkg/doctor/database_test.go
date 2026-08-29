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

package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setDatabaseConfig(t *testing.T, dbType, path string) {
	t.Cleanup(config.ResetForTests)

	config.DatabaseType.Set(dbType)
	config.DatabasePath.Set(path)
}

func TestCheckDatabase_DoesNotCreateSqliteFile(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "vikunja.db")
	setDatabaseConfig(t, "sqlite", dbPath)

	group := CheckDatabase()

	_, err := os.Stat(dbPath)
	require.ErrorIs(t, err, os.ErrNotExist, "doctor must not create the database it reports on")
	entries, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	assert.Empty(t, entries)

	assert.Equal(t, "Database (sqlite)", group.Name)
	// A second result would mean CheckDatabase connected anyway, creating the file.
	require.Len(t, group.Results, 1)
	assert.Equal(t, "Database file", group.Results[0].Name)
	assert.False(t, group.Results[0].Passed)
	assert.Contains(t, group.Results[0].Error, dbPath)
}

func TestCheckDatabase_MemoryPath(t *testing.T) {
	tempDir := t.TempDir()
	// initSqliteEngine backs the ephemeral database with os.MkdirTemp, so redirecting
	// the temp dir makes "nothing was created" observable.
	t.Setenv("TMPDIR", tempDir)
	t.Setenv("TMP", tempDir)
	t.Setenv("TEMP", tempDir)
	setDatabaseConfig(t, "sqlite", db.DatabasePathMemory)

	group := CheckDatabase()

	entries, err := os.ReadDir(tempDir)
	require.NoError(t, err)
	assert.Empty(t, entries, "doctor must not spin up the ephemeral database")

	assert.Equal(t, "Database (sqlite)", group.Name)
	require.Len(t, group.Results, 1)
	assert.Equal(t, "Database file", group.Results[0].Name)
	assert.True(t, group.Results[0].Passed)
	assert.Equal(t, "memory (ephemeral, nothing to verify)", group.Results[0].Value)
}

func TestCheckDatabase_SqlitePathIsADirectory(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "vikunja.db")
	require.NoError(t, os.Mkdir(dbPath, 0700))
	setDatabaseConfig(t, "sqlite", dbPath)

	group := CheckDatabase()

	assert.Equal(t, "Database (sqlite)", group.Name)
	require.Len(t, group.Results, 1)
	assert.Equal(t, "Database file", group.Results[0].Name)
	assert.False(t, group.Results[0].Passed)
	assert.Equal(t, dbPath+" exists but is not a file", group.Results[0].Error)
}
