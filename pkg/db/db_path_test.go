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
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"

	"github.com/spf13/viper"
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
		{
			name: "memory database",
			cfg: DatabasePathConfig{
				ConfiguredPath:     "memory",
				RootPath:           "/opt/vikunja",
				RootPathConfigured: false,
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "memory",
		},

		{
			name: "absolute path should be used as-is",
			cfg: DatabasePathConfig{
				ConfiguredPath:     "/var/lib/vikunja/vikunja.db",
				RootPath:           "/opt/vikunja",
				RootPathConfigured: false,
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/var/lib/vikunja/vikunja.db",
		},
		{
			name: "absolute path with different rootpath still used as-is",
			cfg: DatabasePathConfig{
				ConfiguredPath:     "/data/mydb.db",
				RootPath:           "/custom/path",
				RootPathConfigured: true,
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/data/mydb.db",
		},

		{
			name: "relative path with explicit rootpath",
			cfg: DatabasePathConfig{
				ConfiguredPath:     "vikunja.db",
				RootPath:           "/var/lib/vikunja",
				RootPathConfigured: true,
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/var/lib/vikunja/vikunja.db",
		},
		{
			name: "relative subdirectory path with explicit rootpath",
			cfg: DatabasePathConfig{
				ConfiguredPath:     "data/vikunja.db",
				RootPath:           "/var/lib/vikunja",
				RootPathConfigured: true,
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/var/lib/vikunja/data/vikunja.db",
		},

		{
			name: "relative path with unconfigured rootpath uses user data dir",
			cfg: DatabasePathConfig{
				ConfiguredPath:     "vikunja.db",
				RootPath:           "/opt/vikunja",
				RootPathConfigured: false,
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/home/user/.local/share/vikunja/vikunja.db",
		},

		{
			name: "falls back to rootpath when getUserDataDir fails",
			cfg: DatabasePathConfig{
				ConfiguredPath:     "vikunja.db",
				RootPath:           "/opt/vikunja",
				RootPathConfigured: false,
			},
			getUserDataDir: mockGetUserDataDirError,
			expected:       "/opt/vikunja/vikunja.db",
		},

		{
			name: "empty configured path with explicit rootpath",
			cfg: DatabasePathConfig{
				ConfiguredPath:     "",
				RootPath:           "/var/lib/vikunja",
				RootPathConfigured: true,
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/var/lib/vikunja",
		},
		{
			name: "empty configured path with unconfigured rootpath",
			cfg: DatabasePathConfig{
				ConfiguredPath:     "",
				RootPath:           "/opt/vikunja",
				RootPathConfigured: false,
			},
			getUserDataDir: mockGetUserDataDir("/home/user/.local/share/vikunja"),
			expected:       "/home/user/.local/share/vikunja",
		},
		{
			name: "path with dots normalized",
			cfg: DatabasePathConfig{
				ConfiguredPath:     "/var/lib/vikunja/../vikunja/./db.db",
				RootPath:           "/opt/vikunja",
				RootPathConfigured: false,
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
	t.Run("with explicitly configured rootpath", func(t *testing.T) {
		cfg := DatabasePathConfig{
			ConfiguredPath:     "vikunja.db",
			RootPath:           "/custom/path",
			RootPathConfigured: true,
		}

		result, err := resolveDatabasePath(cfg, getUserDataDir)
		require.NoError(t, err)

		expected := filepath.Join("/custom/path", "vikunja.db")
		assert.Equal(t, expected, result)
	})

	t.Run("with unconfigured rootpath uses user data directory", func(t *testing.T) {
		execPath, err := os.Executable()
		require.NoError(t, err)
		execDir := filepath.Dir(execPath)

		cfg := DatabasePathConfig{
			ConfiguredPath:     "vikunja.db",
			RootPath:           execDir,
			RootPathConfigured: false,
		}

		result, err := resolveDatabasePath(cfg, getUserDataDir)
		require.NoError(t, err)

		assert.NotEqual(t, filepath.Join(execDir, "vikunja.db"), result)
		assert.Contains(t, result, "vikunja.db")

		switch runtime.GOOS {
		case "windows":
			assert.Contains(t, result, "Vikunja")
		case "darwin":
			assert.Contains(t, result, "Library")
			assert.Contains(t, result, "Application Support")
		default:
			assert.NotEqual(t,
				filepath.Dir(result),
				execDir,
				"Database should not be in executable directory",
			)
		}
	})

	t.Run("with subdirectory path", func(t *testing.T) {
		cfg := DatabasePathConfig{
			ConfiguredPath:     "data/vikunja.db",
			RootPath:           "/custom/path",
			RootPathConfigured: true,
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
				ConfiguredPath:     "C:\\ProgramData\\Vikunja\\vikunja.db",
				RootPath:           "C:\\Program Files\\Vikunja",
				RootPathConfigured: false,
			},
			getUserDataDir: mockGetUserDataDir("C:\\Users\\test\\AppData\\Local\\Vikunja"),
			expected:       "C:\\ProgramData\\Vikunja\\vikunja.db",
		},
		{
			name: "windows relative path with explicit rootpath",
			cfg: DatabasePathConfig{
				ConfiguredPath:     "vikunja.db",
				RootPath:           "C:\\ProgramData\\Vikunja",
				RootPathConfigured: true,
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

func Test_databasePathConfig(t *testing.T) {
	// Mirrors the env wiring InitConfig sets up so VIKUNJA_* reaches viper.
	setup := func(t *testing.T) {
		t.Helper()
		viper.Reset()
		t.Cleanup(viper.Reset)
		config.InitDefaultConfig()
		viper.SetEnvPrefix("vikunja")
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()
	}

	execDir := func(t *testing.T) string {
		t.Helper()
		execPath, err := os.Executable()
		require.NoError(t, err)
		return filepath.Dir(execPath)
	}

	t.Run("database.path defaults to a relative path", func(t *testing.T) {
		setup(t)

		assert.Equal(t, "vikunja.db", databasePathConfig().ConfiguredPath)
	})

	t.Run("an explicit rootpath moves the default database", func(t *testing.T) {
		setup(t)
		root := t.TempDir()
		t.Setenv("VIKUNJA_SERVICE_ROOTPATH", root)

		path, err := resolveDatabasePath(databasePathConfig(), getUserDataDir)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(root, "vikunja.db"), path)
	})

	t.Run("a rootpath explicitly configured to the executable directory is honoured", func(t *testing.T) {
		setup(t)
		root := execDir(t)
		t.Setenv("VIKUNJA_SERVICE_ROOTPATH", root)
		dataDir := t.TempDir()

		path, err := resolveDatabasePath(databasePathConfig(), func() (string, error) { return dataDir, nil })
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(root, "vikunja.db"), path)
	})

	t.Run("an unconfigured rootpath falls back to the user data dir", func(t *testing.T) {
		setup(t)
		dataDir := t.TempDir()

		path, err := resolveDatabasePath(databasePathConfig(), func() (string, error) { return dataDir, nil })
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(dataDir, "vikunja.db"), path)
	})

	t.Run("an absolute database.path ignores the rootpath", func(t *testing.T) {
		setup(t)
		t.Setenv("VIKUNJA_SERVICE_ROOTPATH", t.TempDir())
		t.Setenv("VIKUNJA_DATABASE_PATH", "/var/lib/vikunja/custom.db")

		path, err := resolveDatabasePath(databasePathConfig(), getUserDataDir)
		require.NoError(t, err)
		assert.Equal(t, "/var/lib/vikunja/custom.db", path)
	})
}

func Test_strandedLegacyDatabasePath(t *testing.T) {
	setup := func(t *testing.T) string {
		t.Helper()
		viper.Reset()
		t.Cleanup(viper.Reset)
		config.InitDefaultConfig()

		t.Chdir(t.TempDir())
		cwd, err := os.Getwd()
		require.NoError(t, err)
		return cwd
	}

	writeDB := func(t *testing.T, path string) string {
		t.Helper()
		require.NoError(t, os.WriteFile(path, []byte("db"), 0600))
		return path
	}

	t.Run("warns when the database was left behind in the working directory", func(t *testing.T) {
		cwd := setup(t)
		legacy := writeDB(t, filepath.Join(cwd, "vikunja.db"))

		assert.Equal(t, legacy, strandedLegacyDatabasePath(filepath.Join(t.TempDir(), "vikunja.db")))
	})

	t.Run("stays quiet when the resolved database exists", func(t *testing.T) {
		cwd := setup(t)
		writeDB(t, filepath.Join(cwd, "vikunja.db"))
		resolved := writeDB(t, filepath.Join(t.TempDir(), "vikunja.db"))

		assert.Empty(t, strandedLegacyDatabasePath(resolved))
	})

	t.Run("stays quiet when there is no database in the working directory", func(t *testing.T) {
		setup(t)

		assert.Empty(t, strandedLegacyDatabasePath(filepath.Join(t.TempDir(), "vikunja.db")))
	})

	t.Run("stays quiet when database.path was configured explicitly", func(t *testing.T) {
		cwd := setup(t)
		writeDB(t, filepath.Join(cwd, "vikunja.db"))
		t.Setenv("VIKUNJA_DATABASE_PATH", "/var/lib/vikunja/vikunja.db")

		assert.Empty(t, strandedLegacyDatabasePath(filepath.Join(t.TempDir(), "vikunja.db")))
	})

	t.Run("stays quiet when the database did not move", func(t *testing.T) {
		cwd := setup(t)
		legacy := writeDB(t, filepath.Join(cwd, "vikunja.db"))

		assert.Empty(t, strandedLegacyDatabasePath(legacy))
	})
}
