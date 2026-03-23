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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
