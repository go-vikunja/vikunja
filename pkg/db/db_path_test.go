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
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveDatabasePath(t *testing.T) {
	// Save original config values
	originalRootpath := config.ServiceRootpath.GetString()
	defer config.ServiceRootpath.Set(originalRootpath)

	t.Run("with explicitly configured rootpath", func(t *testing.T) {
		// Set a rootpath that is different from the executable location
		config.ServiceRootpath.Set("/custom/path")

		result := resolveDatabasePath("vikunja.db")
		expected := filepath.Join("/custom/path", "vikunja.db")

		assert.Equal(t, expected, result)
	})

	t.Run("with default rootpath uses user data directory", func(t *testing.T) {
		// Get the actual executable path and set rootpath to it
		execPath, err := os.Executable()
		require.NoError(t, err)
		config.ServiceRootpath.Set(filepath.Dir(execPath))

		result := resolveDatabasePath("vikunja.db")

		// Result should contain the platform-specific user data directory
		// and not be in the executable directory
		assert.NotEqual(t, filepath.Join(filepath.Dir(execPath), "vikunja.db"), result)
		assert.Contains(t, result, "vikunja.db")

		// Verify it's using a platform-appropriate path
		switch runtime.GOOS {
		case "windows":
			// Should be in %LOCALAPPDATA%\Vikunja or %USERPROFILE%\AppData\Local\Vikunja
			assert.Contains(t, result, "Vikunja")
		case "darwin":
			// Should be in ~/Library/Application Support/Vikunja
			assert.Contains(t, result, "Library")
			assert.Contains(t, result, "Application Support")
		default:
			// Should be in ~/.local/share/vikunja or $XDG_DATA_HOME/vikunja
			assert.NotEqual(t,
				filepath.Dir(result),
				filepath.Dir(execPath),
				"Database should not be in executable directory",
			)
		}
	})

	t.Run("with subdirectory path", func(t *testing.T) {
		config.ServiceRootpath.Set("/custom/path")

		result := resolveDatabasePath("data/vikunja.db")
		expected := filepath.Join("/custom/path", "data", "vikunja.db")

		assert.Equal(t, expected, result)
	})
}

func TestGetUserDataDir(t *testing.T) {
	dataDir, err := getUserDataDir()
	require.NoError(t, err)
	assert.NotEmpty(t, dataDir)

	// Verify the directory was created
	info, err := os.Stat(dataDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify platform-specific paths
	switch runtime.GOOS {
	case "windows":
		assert.Contains(t, dataDir, "Vikunja")
	case "darwin":
		assert.Contains(t, dataDir, "Library")
		assert.Contains(t, dataDir, "Application Support")
		assert.Contains(t, dataDir, "Vikunja")
	default:
		// Linux or other Unix-like
		home, _ := os.UserHomeDir()
		xdgDataHome := os.Getenv("XDG_DATA_HOME")
		if xdgDataHome != "" {
			assert.Contains(t, dataDir, "vikunja")
		} else {
			assert.Contains(t, dataDir, filepath.Join(home, ".local", "share", "vikunja"))
		}
	}
}

func TestIsSystemDirectory(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Windows system directories
		{
			name:     "Windows System32",
			path:     "C:\\Windows\\System32\\vikunja.db",
			expected: runtime.GOOS == "windows",
		},
		{
			name:     "Windows SysWOW64",
			path:     "C:\\Windows\\SysWOW64\\vikunja.db",
			expected: runtime.GOOS == "windows",
		},
		{
			name:     "Windows root",
			path:     "C:\\Windows\\vikunja.db",
			expected: runtime.GOOS == "windows",
		},
		{
			name:     "Windows System32 lowercase",
			path:     "c:\\windows\\system32\\vikunja.db",
			expected: runtime.GOOS == "windows",
		},
		// Unix-like system directories
		{
			name:     "/bin",
			path:     "/bin/vikunja.db",
			expected: runtime.GOOS != "windows",
		},
		{
			name:     "/sbin",
			path:     "/sbin/vikunja.db",
			expected: runtime.GOOS != "windows",
		},
		{
			name:     "/usr/bin",
			path:     "/usr/bin/vikunja.db",
			expected: runtime.GOOS != "windows",
		},
		{
			name:     "/etc",
			path:     "/etc/vikunja.db",
			expected: runtime.GOOS != "windows",
		},
		// Non-system directories
		{
			name:     "user home directory (Unix)",
			path:     "/home/user/vikunja.db",
			expected: false,
		},
		{
			name:     "user profile directory (Windows)",
			path:     "C:\\Users\\user\\vikunja.db",
			expected: false,
		},
		{
			name:     "custom directory",
			path:     "/opt/vikunja/vikunja.db",
			expected: false,
		},
		{
			name:     "relative path",
			path:     "./vikunja.db",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSystemDirectory(tt.path)
			assert.Equal(t, tt.expected, result,
				"Expected isSystemDirectory(%s) to be %v on %s",
				tt.path, tt.expected, runtime.GOOS)
		})
	}
}

func TestIsSystemDirectory_EdgeCases(t *testing.T) {
	// Test paths that contain system directory names but aren't actually in them
	if runtime.GOOS == "windows" {
		assert.False(t, isSystemDirectory("C:\\myapp\\windows\\data\\vikunja.db"),
			"Should not match if 'windows' is part of a custom path")
	} else {
		assert.False(t, isSystemDirectory("/home/bin/vikunja.db"),
			"Should not match /home/bin (different from /bin)")
		assert.False(t, isSystemDirectory("/opt/sbin/vikunja.db"),
			"Should not match /opt/sbin (different from /sbin)")
	}
}
