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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// P10: Gregorian CSV dates stay Gregorian through import to preview.
// Frontend export-preview Jalali rendering is covered by central display tests (formatDate.test.ts).
func TestCSVGregorianDateRoundTrip(t *testing.T) {
	csvContent := "Title,Due Date\nTask 1,2024-01-15\n"
	reader := bytes.NewReader([]byte(csvContent))
	detected, err := DetectCSVStructure(reader, int64(len(csvContent)))
	require.NoError(t, err)
	assert.Equal(t, "2006-01-02", detected.DateFormat)

	config := ImportConfig{
		Delimiter:  ",",
		QuoteChar:  "\"",
		DateFormat: "2006-01-02",
		Mapping: []ColumnMapping{
			{ColumnIndex: 0, ColumnName: "Title", Attribute: AttrTitle},
			{ColumnIndex: 1, ColumnName: "Due Date", Attribute: AttrDueDate},
		},
	}
	preview, err := PreviewImport(bytes.NewReader([]byte(csvContent)), int64(len(csvContent)), &config)
	require.NoError(t, err)
	require.Len(t, preview.Tasks, 1)
	assert.Equal(t, "2024-01-15", preview.Tasks[0].DueDate)

	task := rowToTask([]string{"Task 1", "2024-01-15"}, &config, 1)
	assert.Equal(t, 2024, task.DueDate.Year())
	assert.Equal(t, 1, int(task.DueDate.Month()))
	assert.Equal(t, 15, task.DueDate.Day())
	assert.Equal(t, "2024-01-15", task.DueDate.Format("2006-01-02"))
	assert.Equal(t, preview.Tasks[0].DueDate, task.DueDate.Format("2006-01-02"))

	for _, s := range []string{preview.Tasks[0].DueDate, task.DueDate.Format("2006-01-02")} {
		assert.NotContains(t, s, "۱۴۰")
		assert.True(t, strings.IndexFunc(s, func(r rune) bool { return r >= '۰' && r <= '۹' }) == -1)
	}
}
