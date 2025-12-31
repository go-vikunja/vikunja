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
)

func TestStripBOM(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "with BOM",
			input:    []byte{0xEF, 0xBB, 0xBF, 'H', 'e', 'l', 'l', 'o'},
			expected: []byte("Hello"),
		},
		{
			name:     "without BOM",
			input:    []byte("Hello"),
			expected: []byte("Hello"),
		},
		{
			name:     "empty",
			input:    []byte{},
			expected: []byte{},
		},
		{
			name:     "only BOM",
			input:    []byte{0xEF, 0xBB, 0xBF},
			expected: []byte{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := stripBOM(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDetectDelimiter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "comma separated",
			input:    "name,email,phone\nJohn,john@test.com,123\nJane,jane@test.com,456",
			expected: ",",
		},
		{
			name:     "semicolon separated",
			input:    "name;email;phone\nJohn;john@test.com;123\nJane;jane@test.com;456",
			expected: ";",
		},
		{
			name:     "tab separated",
			input:    "name\temail\tphone\nJohn\tjohn@test.com\t123\nJane\tjane@test.com\t456",
			expected: "\t",
		},
		{
			name:     "pipe separated",
			input:    "name|email|phone\nJohn|john@test.com|123\nJane|jane@test.com|456",
			expected: "|",
		},
		{
			name:     "single line defaults to comma",
			input:    "just a single line",
			expected: ",",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detectDelimiter([]byte(tc.input))
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDetectQuoteChar(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "double quotes",
			input:    `"name","email"\n"John","john@test.com"`,
			expected: "\"",
		},
		{
			name:     "single quotes",
			input:    `'name','email'\n'John','john@test.com'`,
			expected: "'",
		},
		{
			name:     "no quotes defaults to double",
			input:    "name,email\nJohn,john@test.com",
			expected: "\"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detectQuoteChar([]byte(tc.input))
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDetectDateFormat(t *testing.T) {
	tests := []struct {
		name        string
		sampleDates []string
		expected    string
	}{
		{
			name:        "ISO date",
			sampleDates: []string{"2024-01-15", "2024-02-20", "2024-03-25"},
			expected:    "2006-01-02",
		},
		{
			name:        "ISO datetime",
			sampleDates: []string{"2024-01-15T10:30:00", "2024-02-20T14:45:00"},
			expected:    "2006-01-02T15:04:05",
		},
		{
			name:        "European format",
			sampleDates: []string{"15.01.2024", "20.02.2024", "25.03.2024"},
			expected:    "02.01.2006",
		},
		{
			name:        "empty defaults to ISO",
			sampleDates: []string{},
			expected:    "2006-01-02",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := detectDateFormat(tc.sampleDates)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSuggestMapping(t *testing.T) {
	tests := []struct {
		name     string
		columns  []string
		expected map[int]TaskAttribute
	}{
		{
			name:    "standard column names",
			columns: []string{"Title", "Description", "Due Date", "Priority", "Labels"},
			expected: map[int]TaskAttribute{
				0: AttrTitle,
				1: AttrDescription,
				2: AttrDueDate,
				3: AttrPriority,
				4: AttrLabels,
			},
		},
		{
			name:    "alternative column names",
			columns: []string{"Task Name", "Notes", "Deadline", "Tags", "Project"},
			expected: map[int]TaskAttribute{
				0: AttrTitle,
				1: AttrDescription,
				2: AttrDueDate,
				3: AttrLabels,
				4: AttrProject,
			},
		},
		{
			name:    "unknown columns",
			columns: []string{"ID", "Random Column", "Unknown"},
			expected: map[int]TaskAttribute{
				0: AttrIgnore,
				1: AttrIgnore,
				2: AttrIgnore,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mappings := suggestMapping(tc.columns)
			require.Len(t, mappings, len(tc.columns))

			for idx, expectedAttr := range tc.expected {
				assert.Equal(t, expectedAttr, mappings[idx].Attribute, "Column %d (%s)", idx, tc.columns[idx])
			}
		})
	}
}

func TestParseCSV(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		delimiter     string
		quoteChar     string
		expectedCols  []string
		expectedRows  int
		expectedError bool
	}{
		{
			name:         "simple comma CSV",
			input:        "name,email,phone\nJohn,john@test.com,123\nJane,jane@test.com,456",
			delimiter:    ",",
			quoteChar:    "\"",
			expectedCols: []string{"name", "email", "phone"},
			expectedRows: 2,
		},
		{
			name:         "semicolon CSV",
			input:        "name;email;phone\nJohn;john@test.com;123",
			delimiter:    ";",
			quoteChar:    "\"",
			expectedCols: []string{"name", "email", "phone"},
			expectedRows: 1,
		},
		{
			name:         "quoted fields",
			input:        "name,description\n\"John Doe\",\"A long, complicated description\"\nJane,Simple",
			delimiter:    ",",
			quoteChar:    "\"",
			expectedCols: []string{"name", "description"},
			expectedRows: 2,
		},
		{
			name:         "with BOM",
			input:        "\xEF\xBB\xBFname,email\nJohn,john@test.com",
			delimiter:    ",",
			quoteChar:    "\"",
			expectedCols: []string{"name", "email"},
			expectedRows: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			headers, rows, err := parseCSV([]byte(tc.input), tc.delimiter, tc.quoteChar)

			if tc.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedCols, headers)
			assert.Len(t, rows, tc.expectedRows)
		})
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"yes", true},
		{"Yes", true},
		{"1", true},
		{"done", true},
		{"completed", true},
		{"false", false},
		{"no", false},
		{"0", false},
		{"", false},
		{"random", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := parseBool(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"0", 0},
		{"1", 1},
		{"3", 3},
		{"5", 5},
		{"10", 5}, // capped at 5
		{"-1", 0}, // minimum 0
		{"low", 2},
		{"medium", 3},
		{"high", 4},
		{"urgent", 5},
		{"highest", 5},
		{"lowest", 1},
		{"normal", 3},
		{"random", 0},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := parsePriority(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestParseLabels(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"work, personal, urgent", []string{"work", "personal", "urgent"}},
		{"single", []string{"single"}},
		{"  spaced  ,  labels  ", []string{"spaced", "labels"}},
		{"", []string{}},
		{",,,", []string{}},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := parseLabels(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDetectCSVStructure(t *testing.T) {
	csvContent := `Title,Description,Due Date,Priority,Labels
Task 1,Description 1,2024-01-15,high,work
Task 2,Description 2,2024-01-20,low,"personal, urgent"
Task 3,Description 3,2024-01-25,medium,home`

	reader := bytes.NewReader([]byte(csvContent))

	result, err := DetectCSVStructure(reader, int64(len(csvContent)))
	require.NoError(t, err)

	assert.Equal(t, []string{"Title", "Description", "Due Date", "Priority", "Labels"}, result.Columns)
	assert.Equal(t, ",", result.Delimiter)
	assert.Len(t, result.SuggestedMapping, 5)
	assert.Len(t, result.PreviewRows, 3)

	// Check suggested mappings
	titleMapping := result.SuggestedMapping[0]
	assert.Equal(t, AttrTitle, titleMapping.Attribute)
	assert.Equal(t, "Title", titleMapping.ColumnName)

	descMapping := result.SuggestedMapping[1]
	assert.Equal(t, AttrDescription, descMapping.Attribute)

	dueDateMapping := result.SuggestedMapping[2]
	assert.Equal(t, AttrDueDate, dueDateMapping.Attribute)
}

func TestPreviewImport(t *testing.T) {
	csvContent := `Title,Description,Done,Priority
Task 1,Description 1,true,high
Task 2,Description 2,false,low
Task 3,Description 3,yes,medium
Task 4,Description 4,no,urgent
Task 5,Description 5,1,normal
Task 6,Description 6,0,low`

	config := ImportConfig{
		Delimiter:  ",",
		QuoteChar:  "\"",
		DateFormat: "2006-01-02",
		Mapping: []ColumnMapping{
			{ColumnIndex: 0, ColumnName: "Title", Attribute: AttrTitle},
			{ColumnIndex: 1, ColumnName: "Description", Attribute: AttrDescription},
			{ColumnIndex: 2, ColumnName: "Done", Attribute: AttrDone},
			{ColumnIndex: 3, ColumnName: "Priority", Attribute: AttrPriority},
		},
	}

	reader := bytes.NewReader([]byte(csvContent))

	result, err := PreviewImport(reader, int64(len(csvContent)), config)
	require.NoError(t, err)

	assert.Equal(t, 6, result.TotalRows)
	assert.Len(t, result.Tasks, 5) // Preview limited to 5

	// Check first task
	assert.Equal(t, "Task 1", result.Tasks[0].Title)
	assert.Equal(t, "Description 1", result.Tasks[0].Description)
	assert.True(t, result.Tasks[0].Done)
	assert.Equal(t, 4, result.Tasks[0].Priority) // "high" -> 4

	// Check second task
	assert.Equal(t, "Task 2", result.Tasks[1].Title)
	assert.False(t, result.Tasks[1].Done)
	assert.Equal(t, 2, result.Tasks[1].Priority) // "low" -> 2
}

func TestConvertToVikunja(t *testing.T) {
	rows := [][]string{
		{"Task 1", "Description 1", "Project A"},
		{"Task 2", "Description 2", "Project A"},
		{"Task 3", "Description 3", "Project B"},
		{"Task 4", "Description 4", ""}, // No project -> default
	}

	config := ImportConfig{
		Delimiter:  ",",
		QuoteChar:  "\"",
		DateFormat: "2006-01-02",
		Mapping: []ColumnMapping{
			{ColumnIndex: 0, Attribute: AttrTitle},
			{ColumnIndex: 1, Attribute: AttrDescription},
			{ColumnIndex: 2, Attribute: AttrProject},
		},
	}

	result := convertToVikunja(rows, config)

	// Should have parent project + child projects
	require.GreaterOrEqual(t, len(result), 2)

	// First project should be the parent "Imported from CSV"
	assert.Equal(t, "Imported from CSV", result[0].Title)

	// Find Project A
	var projectA, projectB, tasksProject *struct {
		title    string
		numTasks int
	}
	for _, p := range result[1:] {
		switch p.Title {
		case "Project A":
			projectA = &struct {
				title    string
				numTasks int
			}{p.Title, len(p.Tasks)}
		case "Project B":
			projectB = &struct {
				title    string
				numTasks int
			}{p.Title, len(p.Tasks)}
		case "Tasks":
			tasksProject = &struct {
				title    string
				numTasks int
			}{p.Title, len(p.Tasks)}
		}
	}

	assert.NotNil(t, projectA, "Project A should exist")
	assert.Equal(t, 2, projectA.numTasks, "Project A should have 2 tasks")

	assert.NotNil(t, projectB, "Project B should exist")
	assert.Equal(t, 1, projectB.numTasks, "Project B should have 1 task")

	assert.NotNil(t, tasksProject, "Tasks project should exist for tasks without project")
	assert.Equal(t, 1, tasksProject.numTasks, "Tasks project should have 1 task")
}

func TestRowToTask(t *testing.T) {
	row := []string{"My Task", "Task description", "2024-01-15", "high", "work, urgent"}

	config := ImportConfig{
		DateFormat: "2006-01-02",
		Mapping: []ColumnMapping{
			{ColumnIndex: 0, Attribute: AttrTitle},
			{ColumnIndex: 1, Attribute: AttrDescription},
			{ColumnIndex: 2, Attribute: AttrDueDate},
			{ColumnIndex: 3, Attribute: AttrPriority},
			{ColumnIndex: 4, Attribute: AttrLabels},
		},
	}

	task := rowToTask(row, config, 1)

	assert.Equal(t, "My Task", task.Title)
	assert.Equal(t, "Task description", task.Description)
	assert.Equal(t, 2024, task.DueDate.Year())
	assert.Equal(t, 1, int(task.DueDate.Month()))
	assert.Equal(t, 15, task.DueDate.Day())
	assert.Equal(t, int64(4), task.Priority) // "high" -> 4
	require.Len(t, task.Labels, 2)
	assert.Equal(t, "work", task.Labels[0].Title)
	assert.Equal(t, "urgent", task.Labels[1].Title)
}

func TestMigratorName(t *testing.T) {
	m := &Migrator{}
	assert.Equal(t, "csv", m.Name())
}

func TestEmptyFile(t *testing.T) {
	reader := bytes.NewReader([]byte{})

	_, err := DetectCSVStructure(reader, 0)
	require.Error(t, err)
}

func TestRowToTaskWithMissingColumns(t *testing.T) {
	// Row with fewer columns than expected
	row := []string{"My Task"}

	config := ImportConfig{
		Mapping: []ColumnMapping{
			{ColumnIndex: 0, Attribute: AttrTitle},
			{ColumnIndex: 1, Attribute: AttrDescription}, // Index 1 doesn't exist
			{ColumnIndex: 2, Attribute: AttrDueDate},     // Index 2 doesn't exist
		},
	}

	task := rowToTask(row, config, 1)

	// Should still work with available columns
	assert.Equal(t, "My Task", task.Title)
	assert.Empty(t, task.Description)
	assert.True(t, task.DueDate.IsZero())
}

func TestRowToTaskWithEmptyTitle(t *testing.T) {
	row := []string{"", "Some description"}

	config := ImportConfig{
		Mapping: []ColumnMapping{
			{ColumnIndex: 0, Attribute: AttrTitle},
			{ColumnIndex: 1, Attribute: AttrDescription},
		},
	}

	task := rowToTask(row, config, 1)

	// Should have default title
	assert.Equal(t, "Untitled Task", task.Title)
	assert.Equal(t, "Some description", task.Description)
}

func TestDoneTask(t *testing.T) {
	row := []string{"Done Task", "completed"}

	config := ImportConfig{
		Mapping: []ColumnMapping{
			{ColumnIndex: 0, Attribute: AttrTitle},
			{ColumnIndex: 1, Attribute: AttrDone},
		},
	}

	task := rowToTask(row, config, 1)

	assert.Equal(t, "Done Task", task.Title)
	assert.True(t, task.Done)
	assert.False(t, task.DoneAt.IsZero()) // DoneAt should be set
}
