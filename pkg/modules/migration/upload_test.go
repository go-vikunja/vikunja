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

package migration

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpoolUpload(t *testing.T) {
	t.Run("round trip", func(t *testing.T) {
		name, size, err := SpoolUpload(strings.NewReader("some export"))
		require.NoError(t, err)
		defer RemoveSpooledUpload(name)

		assert.Equal(t, int64(len("some export")), size)
		assert.NotContains(t, name, string(os.PathSeparator))

		f, err := OpenSpooledUpload(name)
		require.NoError(t, err)
		defer f.Close()

		content, err := io.ReadAll(f)
		require.NoError(t, err)
		assert.Equal(t, "some export", string(content))
	})

	t.Run("remove deletes the file", func(t *testing.T) {
		name, _, err := SpoolUpload(strings.NewReader("gone soon"))
		require.NoError(t, err)

		RemoveSpooledUpload(name)

		_, err = OpenSpooledUpload(name)
		assert.True(t, os.IsNotExist(err), "expected the spooled upload to be gone, got %v", err)
	})

	t.Run("a name from the queue cannot escape the spool directory", func(t *testing.T) {
		dir, err := spoolDir()
		require.NoError(t, err)
		outside := filepath.Join(filepath.Dir(dir), "vikunja-spool-escape-target")
		require.NoError(t, os.WriteFile(outside, []byte("secret"), 0600))
		defer os.Remove(outside)

		_, err = OpenSpooledUpload("../" + filepath.Base(outside))
		assert.True(t, os.IsNotExist(err), "traversal must not resolve to a file outside the spool dir, got %v", err)

		RemoveSpooledUpload("../" + filepath.Base(outside))
		assert.FileExists(t, outside, "traversal must not delete a file outside the spool dir")
	})

	t.Run("an empty name is rejected", func(t *testing.T) {
		_, err := OpenSpooledUpload("")
		require.ErrorIs(t, err, errEmptySpoolName)
	})

	t.Run("cleanup removes orphaned uploads", func(t *testing.T) {
		name, _, err := SpoolUpload(strings.NewReader("orphan"))
		require.NoError(t, err)

		CleanupSpooledUploads()

		_, err = OpenSpooledUpload(name)
		assert.True(t, os.IsNotExist(err), "expected the orphaned upload to be cleaned up, got %v", err)
	})
}
