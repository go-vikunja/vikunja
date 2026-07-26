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

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initConfigFromYAML runs InitConfig against a config file in a temporary directory.
// cors.enable is turned off because InitConfig aborts when it is on without a publicurl.
func initConfigFromYAML(t *testing.T, config string) {
	t.Helper()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yml"), []byte("cors:\n  enable: false\n"+config), 0600))

	cwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(cwd))
		viper.Reset()
	})

	viper.Reset()
	InitConfig()
}

func TestServiceSecret(t *testing.T) {
	t.Run("service.jwtsecret is migrated to service.secret", func(t *testing.T) {
		initConfigFromYAML(t, "service:\n  jwtsecret: legacy-secret\n")

		assert.Equal(t, "legacy-secret", ServiceSecret.GetString())
	})
	t.Run("service.jwtsecret is migrated from the environment", func(t *testing.T) {
		t.Setenv("VIKUNJA_SERVICE_JWTSECRET", "legacy-env-secret")
		initConfigFromYAML(t, "")

		assert.Equal(t, "legacy-env-secret", ServiceSecret.GetString())
	})
	t.Run("service.secret wins over service.jwtsecret", func(t *testing.T) {
		initConfigFromYAML(t, "service:\n  secret: new-secret\n  jwtsecret: legacy-secret\n")

		assert.Equal(t, "new-secret", ServiceSecret.GetString())
	})
	t.Run("a secret is generated when none is configured", func(t *testing.T) {
		initConfigFromYAML(t, "")

		assert.Len(t, ServiceSecret.GetString(), 64)
	})
	t.Run("InitDefaultConfig generates a secret on its own", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)
		InitDefaultConfig()

		assert.Len(t, ServiceSecret.GetString(), 64)
	})
}

func TestGetRootpathLocation(t *testing.T) {
	// The function should return the current working directory
	expected, err := os.Getwd()
	require.NoError(t, err)

	result := getRootpathLocation()
	assert.Equal(t, expected, result)
}

func TestResolvePath(t *testing.T) {
	// Save and restore rootpath
	original := ServiceRootpath.GetString()
	defer ServiceRootpath.Set(original)
	ServiceRootpath.Set("/var/lib/vikunja")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "absolute path returned as-is",
			input:    "/etc/vikunja/config.yml",
			expected: "/etc/vikunja/config.yml",
		},
		{
			name:     "relative path joined with rootpath",
			input:    "files",
			expected: "/var/lib/vikunja/files",
		},
		{
			name:     "relative subdir path joined with rootpath",
			input:    "data/vikunja.db",
			expected: "/var/lib/vikunja/data/vikunja.db",
		},
		{
			name:     "empty string returns rootpath",
			input:    "",
			expected: "/var/lib/vikunja",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolvePath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
