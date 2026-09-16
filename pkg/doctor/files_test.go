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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setFilesConfig(t *testing.T, filesType, basePath string) {
	t.Cleanup(config.ResetForTests)

	config.FilesType.Set(filesType)
	config.FilesBasePath.Set(basePath)
}

func TestCheckFiles_DoesNotCreateBasePath(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "files")
	setFilesConfig(t, "local", basePath)

	group := CheckFiles()

	_, err := os.Stat(basePath)
	require.ErrorIs(t, err, os.ErrNotExist, "doctor must not create the directory it reports on")

	require.Len(t, group.Results, 2)
	assert.Equal(t, "Path", group.Results[0].Name)
	assert.Equal(t, basePath, group.Results[0].Value)
	assert.Equal(t, "Directory exists", group.Results[1].Name)
	assert.False(t, group.Results[1].Passed)
}

func TestCheckFiles_ExistingBasePath(t *testing.T) {
	basePath := t.TempDir()
	setFilesConfig(t, "local", basePath)

	group := CheckFiles()

	require.NotEmpty(t, group.Results)

	var names []string
	var writable *CheckResult
	for i, result := range group.Results {
		assert.True(t, result.Passed, "check %q failed: %s", result.Name, result.Error)
		// "Ownership match" only appears as non-root under an active user namespace (rootless Docker).
		if result.Name == "Ownership match" {
			continue
		}
		names = append(names, result.Name)
		if result.Name == "Writable" {
			writable = &group.Results[i]
		}
	}

	assert.Equal(t, []string{
		"Path",
		"Directory exists",
		"Directory permissions",
		"Directory owner",
		"Writable",
		"Disk space",
		"Stored files",
	}, names)

	require.NotNil(t, writable)
	assert.True(t, writable.Passed, "writability probe failed: %s", writable.Error)
	assert.Equal(t, "yes", writable.Value)

	entries, err := os.ReadDir(basePath)
	require.NoError(t, err)
	assert.Empty(t, entries, "the writability probe must clean up after itself")
}

func TestCheckFiles_BasePathIsAFile(t *testing.T) {
	basePath := filepath.Join(t.TempDir(), "files")
	require.NoError(t, os.WriteFile(basePath, []byte("not a directory"), 0600))
	setFilesConfig(t, "local", basePath)

	group := CheckFiles()

	require.Len(t, group.Results, 2)
	assert.Equal(t, "Path", group.Results[0].Name)
	assert.Equal(t, "Directory exists", group.Results[1].Name)
	assert.False(t, group.Results[1].Passed)
	assert.Contains(t, group.Results[1].Error, "exists but is not a directory")
}
