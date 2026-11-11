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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/unicode"
)

const encodingTestCSV = "Title\tDescription\tLabels\nTäsk 1\tDescription 1\twork\n"

var encodingTestConfig = &ImportConfig{
	Delimiter: "\t",
	Mapping: []ColumnMapping{
		{ColumnIndex: 0, ColumnName: "Title", Attribute: AttrTitle},
		{ColumnIndex: 1, ColumnName: "Description", Attribute: AttrDescription},
		{ColumnIndex: 2, ColumnName: "Labels", Attribute: AttrLabels},
	},
}

func encodeUTF16(t *testing.T, endianness unicode.Endianness, s string) []byte {
	t.Helper()
	encoded, err := unicode.UTF16(endianness, unicode.UseBOM).NewEncoder().Bytes([]byte(s))
	require.NoError(t, err)
	return encoded
}

func TestCSVEncoding(t *testing.T) {
	inputs := map[string][]byte{
		"UTF-16LE with BOM": encodeUTF16(t, unicode.LittleEndian, encodingTestCSV),
		"UTF-16BE with BOM": encodeUTF16(t, unicode.BigEndian, encodingTestCSV),
		"NUL bytes":         []byte("Title\tDescr\x00iption\tLabels\nTäsk\x00 1\tDescription\x00 1\twork\x00\n"),
	}

	for name, content := range inputs {
		t.Run(name+" detect", func(t *testing.T) {
			result, err := DetectCSVStructure(bytes.NewReader(content), int64(len(content)))
			require.NoError(t, err)
			assert.Equal(t, "\t", result.Delimiter)
			assert.Equal(t, []string{"Title", "Description", "Labels"}, result.Columns)
			assert.Equal(t, [][]string{{"Täsk 1", "Description 1", "work"}}, result.PreviewRows)
		})

		t.Run(name+" preview", func(t *testing.T) {
			result, err := PreviewImport(bytes.NewReader(content), int64(len(content)), encodingTestConfig)
			require.NoError(t, err)
			require.Len(t, result.Tasks, 1)
			assert.Equal(t, "Täsk 1", result.Tasks[0].Title)
			assert.Equal(t, "Description 1", result.Tasks[0].Description)
			assert.Equal(t, []string{"work"}, result.Tasks[0].Labels)
		})

		t.Run(name+" migrate", func(t *testing.T) {
			projects, err := projectsFromCSV(bytes.NewReader(content), int64(len(content)), encodingTestConfig)
			require.NoError(t, err)
			require.Len(t, projects, 1)
			require.Len(t, projects[0].Tasks, 1)
			task := projects[0].Tasks[0]
			assert.Equal(t, "Täsk 1", task.Title)
			assert.Equal(t, "Description 1", task.Description)
			require.Len(t, task.Labels, 1)
			assert.Equal(t, "work", task.Labels[0].Title)
		})
	}

	t.Run("invalid UTF-8 is replaced", func(t *testing.T) {
		content := []byte("Title\nT\xe4sk\n")
		projects, err := projectsFromCSV(bytes.NewReader(content), int64(len(content)), &ImportConfig{
			Mapping: []ColumnMapping{{ColumnIndex: 0, ColumnName: "Title", Attribute: AttrTitle}},
		})
		require.NoError(t, err)
		require.Len(t, projects[0].Tasks, 1)
		assert.Equal(t, "T�sk", projects[0].Tasks[0].Title)
	})
}
