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

package dump

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GHSA-rp64-xg8g-9mpr
func TestCreateDumpFile(t *testing.T) {
	t.Run("is only readable by the owner", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("unix permissions")
		}
		filename := filepath.Join(t.TempDir(), "dump.zip")

		f, err := createDumpFile(filename)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		info, err := os.Stat(filename)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	})
	t.Run("does not overwrite an existing file", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), "dump.zip")
		require.NoError(t, os.WriteFile(filename, []byte("existing"), 0600))

		_, err := createDumpFile(filename)
		require.ErrorIs(t, err, os.ErrExist)

		content, err := os.ReadFile(filename)
		require.NoError(t, err)
		assert.Equal(t, "existing", string(content))
	})
}
