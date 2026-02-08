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

package log

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigureStandardLoggerWithPath(t *testing.T) {
	tempDir := t.TempDir()

	ConfigureStandardLogger(true, "file", tempDir, "INFO", "text")

	// Verify log file was created in the correct location
	expectedPath := filepath.Join(tempDir, "standard.log")
	_, err := os.Stat(expectedPath)
	require.NoError(t, err, "Log file should be created in configured path")

	// Verify NO log file was created in current directory
	_, err = os.Stat("./standard.log")
	assert.True(t, os.IsNotExist(err), "Log file should NOT be created in current directory")
}

func TestMakeLogHandlerCreatesCorrectLogFile(t *testing.T) {
	tempDir := t.TempDir()
	logPath = tempDir

	tests := []struct {
		name     string
		logfile  string
		expected string
	}{
		{"standard", "standard", "standard.log"},
		{"database", "database", "database.log"},
		{"http", "http", "http.log"},
		{"events", "events", "events.log"},
		{"mail", "mail", "mail.log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = makeLogHandler(true, "file", tt.logfile, "INFO", "text")

			expectedPath := filepath.Join(tempDir, tt.expected)
			_, err := os.Stat(expectedPath)
			require.NoError(t, err, "Log file %s should be created", tt.expected)
		})
	}
}

func TestConfigureStandardLoggerSetsPathBeforeHandler(t *testing.T) {
	tempDir := t.TempDir()

	// This test verifies the fix for issue #2020
	// The path must be set BEFORE makeLogHandler is called
	// so that file output goes to the correct location
	ConfigureStandardLogger(true, "file", tempDir, "INFO", "text")

	// If the fix is correct, the file will be in tempDir
	// If the fix is wrong, the file would be in "." (the old logPath value)
	expectedPath := filepath.Join(tempDir, "standard.log")
	_, err := os.Stat(expectedPath)
	require.NoError(t, err, "Log file must be created in the configured path, not the default")
}
