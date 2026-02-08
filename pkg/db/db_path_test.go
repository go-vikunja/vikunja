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
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"code.vikunja.io/api/pkg/log"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	log.InitLogger()
	os.Exit(m.Run())
}

func Test_resolveDatabasePath(t *testing.T) {
	mockGetUserDataDir := func(path string) func() (string, error) {
		return func() (string, error) {
			return path, nil
		}
	}

	mockGetUserDataDirError := func() (string, error) {
		return "", fmt.Errorf("no home directory")
	}

	tests := []struct {
		name           string
		cfg            DatabasePathConfig
		getUserDataDir func() (string, error)
		expected       string
		expectError    bool
	}{
		// Memory special case
		{
			name: "memory database",
			cfg: DatabasePathConfig{
				ConfiguredPath: "memory",
				RootPath:       "/opt/vikunja",
				ExecutablePath: "/opt/vikunja",
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "memory",
		},

		// Absolute paths - THE BUG FROM ISSUE #2189
		{
			name: "absolute path should be used as-is",
			cfg: DatabasePathConfig{
				ConfiguredPath: "/var/lib/vikunja/vikunja.db",
				RootPath:       "/opt/vikunja",
				ExecutablePath: "/opt/vikunja",
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/var/lib/vikunja/vikunja.db",
		},
		{
			name: "absolute path with different rootpath still used as-is",
			cfg: DatabasePathConfig{
				ConfiguredPath: "/data/mydb.db",
				RootPath:       "/custom/path",
				ExecutablePath: "/opt/vikunja",
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/data/mydb.db",
		},

		// Relative paths with explicit rootpath
		{
			name: "relative path with explicit rootpath",
			cfg: DatabasePathConfig{
				ConfiguredPath: "vikunja.db",
				RootPath:       "/var/lib/vikunja",
				ExecutablePath: "/opt/vikunja",
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/var/lib/vikunja/vikunja.db",
		},
		{
			name: "relative subdirectory path with explicit rootpath",
			cfg: DatabasePathConfig{
				ConfiguredPath: "data/vikunja.db",
				RootPath:       "/var/lib/vikunja",
				ExecutablePath: "/opt/vikunja",
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/var/lib/vikunja/data/vikunja.db",
		},

		// Relative paths with default rootpath (uses user data dir)
		{
			name: "relative path with default rootpath uses user data dir",
			cfg: DatabasePathConfig{
				ConfiguredPath: "vikunja.db",
				RootPath:       "/opt/vikunja",
				ExecutablePath: "/opt/vikunja",
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/home/user/.local/share/vikunja/vikunja.db",
		},

		// Fallback when getUserDataDir fails
		{
			name: "falls back to rootpath when getUserDataDir fails",
			cfg: DatabasePathConfig{
				ConfiguredPath: "vikunja.db",
				RootPath:       "/opt/vikunja",
				ExecutablePath: "/opt/vikunja",
			},
			getUserDataDir: mockGetUserDataDirError,
			expected:       "/opt/vikunja/vikunja.db",
		},

		// Edge cases
		{
			name: "empty configured path with explicit rootpath",
			cfg: DatabasePathConfig{
				ConfiguredPath: "",
				RootPath:       "/var/lib/vikunja",
				ExecutablePath: "/opt/vikunja",
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/var/lib/vikunja",
		},
		{
			name: "empty configured path with default rootpath",
			cfg: DatabasePathConfig{
				ConfiguredPath: "",
				RootPath:       "/opt/vikunja",
				ExecutablePath: "/opt/vikunja",
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/home/user/.local/share/vikunja",
		},
		{
			name: "path with dots normalized",
			cfg: DatabasePathConfig{
				ConfiguredPath: "/var/lib/vikunja/../vikunja/./db.db",
				RootPath:       "/opt/vikunja",
				ExecutablePath: "/opt/vikunja",
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/var/lib/vikunja/db.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveDatabasePath(tt.cfg, tt.getUserDataDir)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_resolveDatabasePath_Integration(t *testing.T) {
	// Integration tests using the real getUserDataDir function

	t.Run("with explicitly configured rootpath", func(t *testing.T) {
		cfg := DatabasePathConfig{
			ConfiguredPath: "vikunja.db",
			RootPath:       "/custom/path",
			ExecutablePath: "/opt/vikunja",
		}

		result, err := resolveDatabasePath(cfg, getUserDataDir)
		require.NoError(t, err)

		expected := filepath.Join("/custom/path", "vikunja.db")
		assert.Equal(t, expected, result)
	})

	t.Run("with default rootpath uses user data directory", func(t *testing.T) {
		// Get the actual executable path
		execPath, err := os.Executable()
		require.NoError(t, err)
		execDir := filepath.Dir(execPath)

		cfg := DatabasePathConfig{
			ConfiguredPath: "vikunja.db",
			RootPath:       execDir,
			ExecutablePath: execDir,
		}

		result, err := resolveDatabasePath(cfg, getUserDataDir)
		require.NoError(t, err)

		// Result should contain the platform-specific user data directory
		// and not be in the executable directory
		assert.NotEqual(t, filepath.Join(execDir, "vikunja.db"), result)
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
				execDir,
				"Database should not be in executable directory",
			)
		}
	})

	t.Run("with subdirectory path", func(t *testing.T) {
		cfg := DatabasePathConfig{
			ConfiguredPath: "data/vikunja.db",
			RootPath:       "/custom/path",
			ExecutablePath: "/opt/vikunja",
		}

		result, err := resolveDatabasePath(cfg, getUserDataDir)
		require.NoError(t, err)

		expected := filepath.Join("/custom/path", "data", "vikunja.db")
		assert.Equal(t, expected, result)
	})
}

func Test_resolveDatabasePath_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	mockGetUserDataDir := func(path string) func() (string, error) {
		return func() (string, error) {
			return path, nil
		}
	}

	tests := []struct {
		name           string
		cfg            DatabasePathConfig
		getUserDataDir func() (string, error)
		expected       string
	}{
		{
			name: "windows absolute path",
			cfg: DatabasePathConfig{
				ConfiguredPath: "C:\\ProgramData\\Vikunja\\vikunja.db",
				RootPath:       "C:\\Program Files\\Vikunja",
				ExecutablePath: "C:\\Program Files\\Vikunja",
			},
			getUserDataDir: mockGetUserDataDir("C:\\Users\\test\\AppData\\Local\\Vikunja"),
			expected:       "C:\\ProgramData\\Vikunja\\vikunja.db",
		},
		{
			name: "windows relative path with explicit rootpath",
			cfg: DatabasePathConfig{
				ConfiguredPath: "vikunja.db",
				RootPath:       "C:\\ProgramData\\Vikunja",
				ExecutablePath: "C:\\Program Files\\Vikunja",
			},
			getUserDataDir: mockGetUserDataDir("C:\\Users\\test\\AppData\\Local\\Vikunja"),
			expected:       "C:\\ProgramData\\Vikunja\\vikunja.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resolveDatabasePath(tt.cfg, tt.getUserDataDir)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetUserDataDir(t *testing.T) {

	test := func() string {
		dataDir, err := getUserDataDir()
		require.NoError(t, err)
		assert.NotEmpty(t, dataDir)

		// Verify the directory was created
		info, err := os.Stat(dataDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		return dataDir
	}

	// Verify platform-specific paths
	switch runtime.GOOS {
	case "windows":
		dataDir := test()
		assert.Contains(t, dataDir, "Vikunja")
	case "darwin":
		dataDir := test()
		assert.Contains(t, dataDir, "Library")
		assert.Contains(t, dataDir, "Application Support")
		assert.Contains(t, dataDir, "Vikunja")
	default:
		originalXDGDataHome := os.Getenv("XDG_DATA_HOME")
		defer func() {
			if originalXDGDataHome != "" {
				os.Setenv("XDG_DATA_HOME", originalXDGDataHome)
			} else {
				os.Unsetenv("XDG_DATA_HOME")
			}
		}()

		t.Run("with XDG_DATA_HOME", func(t *testing.T) {
			os.Setenv("XDG_DATA_HOME", "/tmp")
			dataDir := test()
			assert.Contains(t, dataDir, filepath.Join("/tmp", "vikunja"))
		})

		t.Run("without XDG_DATA_HOME", func(t *testing.T) {
			os.Unsetenv("XDG_DATA_HOME")
			dataDir := test()
			assert.Contains(t, dataDir, "vikunja")
		})
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
	if runtime.GOOS == "windows" {
		t.Run("false positives - paths containing 'windows' but not system directories", func(t *testing.T) {
			tests := []struct {
				name string
				path string
			}{
				{
					name: "custom app with windows in path",
					path: "C:\\myapp\\windows\\data\\vikunja.db",
				},
				{
					name: "windows directory on non-C drive",
					path: "D:\\windows\\vikunja.db",
				},
				{
					name: "user directory named windows",
					path: "C:\\Users\\windows\\vikunja.db",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					assert.False(t, isSystemDirectory(tt.path))
				})
			}
		})

		t.Run("safe Windows subdirectories", func(t *testing.T) {
			assert.False(t, isSystemDirectory("C:\\Windows\\Temp\\vikunja.db"))
		})

		t.Run("actual Windows system directories", func(t *testing.T) {
			tests := []struct {
				name string
				path string
			}{
				{
					name: "Windows root",
					path: "C:\\Windows\\vikunja.db",
				},
				{
					name: "Windows root lowercase",
					path: "c:\\windows\\vikunja.db",
				},
				{
					name: "System32",
					path: "C:\\Windows\\System32\\vikunja.db",
				},
				{
					name: "System32 uppercase",
					path: "C:\\WINDOWS\\SYSTEM32\\vikunja.db",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					assert.True(t, isSystemDirectory(tt.path))
				})
			}
		})
	} else {
		t.Run("false positives - paths containing system dir names", func(t *testing.T) {
			tests := []struct {
				name string
				path string
			}{
				{
					name: "/home/bin not same as /bin",
					path: "/home/bin/vikunja.db",
				},
				{
					name: "/opt/sbin not same as /sbin",
					path: "/opt/sbin/vikunja.db",
				},
				{
					name: "/usr/local/bin is safe",
					path: "/usr/local/bin/vikunja.db",
				},
				{
					name: "/binaries not same as /bin",
					path: "/binaries/vikunja.db",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					assert.False(t, isSystemDirectory(tt.path))
				})
			}
		})

		t.Run("actual Unix system directories", func(t *testing.T) {
			tests := []struct {
				name string
				path string
			}{
				{
					name: "/bin",
					path: "/bin/vikunja.db",
				},
				{
					name: "/etc",
					path: "/etc/vikunja.db",
				},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					assert.True(t, isSystemDirectory(tt.path))
				})
			}
		})
	}
}
