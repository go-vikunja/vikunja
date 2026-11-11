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

package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRotateLogFileKeepsExistingRotations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")
	now := time.Date(2026, 1, 2, 3, 4, 5, 6_000_000, time.UTC)

	for _, content := range []string{"first", "second", "third"} {
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
		require.NoError(t, rotateLogFile(path, now))
		assert.NoFileExists(t, path)
	}

	rotated, err := filepath.Glob(filepath.Join(dir, "audit-*.log"))
	require.NoError(t, err)
	require.Len(t, rotated, 3)

	var contents []string
	for _, f := range rotated {
		b, err := os.ReadFile(f)
		require.NoError(t, err)
		contents = append(contents, string(b))
	}
	assert.ElementsMatch(t, []string{"first", "second", "third"}, contents)
}
