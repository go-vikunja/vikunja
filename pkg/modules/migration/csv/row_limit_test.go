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

package csv

import (
	"bytes"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/migration"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildCSV(count int) string {
	var sb strings.Builder
	sb.WriteString("Title,Description\n")
	for i := 0; i < count; i++ {
		sb.WriteString("Task,Description\n")
	}
	return sb.String()
}

// TestCSVRowLimit covers the typed row limit across CSV operations (GHSA-pqf9-h8g4-8gmh).
func TestCSVRowLimit(t *testing.T) {
	config.MigrationMaxCSVRows.Set("100")
	defer config.MigrationMaxCSVRows.Set("100000")

	importConfig := &ImportConfig{
		Delimiter: ",",
		Mapping: []ColumnMapping{
			{ColumnIndex: 0, ColumnName: "Title", Attribute: AttrTitle},
			{ColumnIndex: 1, ColumnName: "Description", Attribute: AttrDescription},
		},
	}

	assertLimitErr := func(t *testing.T, err error) {
		t.Helper()
		require.Error(t, err)
		var limitErr *migration.ErrImportRowLimitExceeded
		require.ErrorAs(t, err, &limitErr, "expected ErrImportRowLimitExceeded, got %v", err)
	}

	t.Run("exactly the limit passes detection and preview", func(t *testing.T) {
		content := buildCSV(100)
		reader := bytes.NewReader([]byte(content))

		_, err := DetectCSVStructure(reader, int64(len(content)))
		require.NoError(t, err)

		reader = bytes.NewReader([]byte(content))
		_, err = PreviewImport(reader, int64(len(content)), importConfig)
		require.NoError(t, err)
	})

	t.Run("limit+1 rows fail detect", func(t *testing.T) {
		content := buildCSV(101)
		_, err := DetectCSVStructure(bytes.NewReader([]byte(content)), int64(len(content)))
		assertLimitErr(t, err)
	})

	t.Run("limit+1 rows fail preview", func(t *testing.T) {
		content := buildCSV(101)
		_, err := PreviewImport(bytes.NewReader([]byte(content)), int64(len(content)), importConfig)
		assertLimitErr(t, err)
	})

	t.Run("limit+1 rows fail migrate", func(t *testing.T) {
		content := buildCSV(101)
		err := MigrateWithConfig(nil, bytes.NewReader([]byte(content)), int64(len(content)), importConfig)
		assertLimitErr(t, err)
	})

	t.Run("quoted multiline rows still parse", func(t *testing.T) {
		content := "Title,Description\nTask,\"multi\nline, content\"\n"
		headers, rows, err := parseCSV([]byte(content), ",")
		require.NoError(t, err)
		assert.Equal(t, []string{"Title", "Description"}, headers)
		require.Len(t, rows, 1)
		assert.Equal(t, "multi\nline, content", rows[0][1])
	})

	t.Run("BOM handling still works", func(t *testing.T) {
		content := append([]byte{0xEF, 0xBB, 0xBF}, []byte("Title,Description\nTask,Description\n")...)
		headers, _, err := parseCSV(content, ",")
		require.NoError(t, err)
		assert.Equal(t, []string{"Title", "Description"}, headers)
	})
}
