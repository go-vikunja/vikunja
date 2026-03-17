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

package models

import (
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTaskExpandParameterFormats tests the backward-compatible expand parameter handling
// This test verifies that both CSV format (?expand=comments,reactions) and array format (?expand[]=comments&expand[]=reactions) work correctly
func TestTaskExpandParameterFormats(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	// Test data setup
	u := &user.User{ID: 1}

	tests := []struct {
		name        string
		csvExpand   string
		arrayExpand []TaskCollectionExpandable
		description string
	}{
		{
			name:        "CSV format single value",
			csvExpand:   "comments",
			arrayExpand: []TaskCollectionExpandable{},
			description: "Test CSV format with single expand value (as per API documentation)",
		},
		{
			name:        "CSV format multiple values",
			csvExpand:   "comments,reactions",
			arrayExpand: []TaskCollectionExpandable{},
			description: "Test CSV format with multiple expand values (as per API documentation)",
		},
		{
			name:        "array format single value",
			csvExpand:   "",
			arrayExpand: []TaskCollectionExpandable{TaskCollectionExpandComments},
			description: "Test array format with single expand value (backward compatibility)",
		},
		{
			name:        "array format multiple values",
			csvExpand:   "",
			arrayExpand: []TaskCollectionExpandable{TaskCollectionExpandComments, TaskCollectionExpandReactions},
			description: "Test array format with multiple expand values (backward compatibility)",
		},
		{
			name:        "mixed format",
			csvExpand:   "comments",
			arrayExpand: []TaskCollectionExpandable{TaskCollectionExpandReactions},
			description: "Test mixed format - CSV and array parameters combined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create TaskCollection with different expand parameter formats
			tc := &TaskCollection{
				ProjectID: 1,
				ExpandCSV: tt.csvExpand,
				Expand:    tt.arrayExpand,
			}

			// Test that ReadAll processes expand parameters correctly
			result, _, _, err := tc.ReadAll(s, u, "", 1, 50)
			require.NoError(t, err, "ReadAll should not fail for %s", tt.description)
			assert.NotNil(t, result, "Result should not be nil for %s", tt.description)

			// Verify that tasks are returned (actual expansion testing would require specific fixture setup)
			tasks, ok := result.([]*Task)
			require.True(t, ok, "Result should be a slice of tasks for %s", tt.description)
			assert.NotEmpty(t, tasks, "Should return tasks for %s", tt.description)

			// The actual expansion verification would depend on specific fixtures with comments/reactions
			// This test primarily verifies that the parameter parsing doesn't cause errors
		})
	}
}

// TestTaskReadOneExpandParameterFormats tests the expand parameter handling for individual task retrieval
func TestTaskReadOneExpandParameterFormats(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	u := &user.User{ID: 1}

	tests := []struct {
		name        string
		taskID      int64
		csvExpand   string
		arrayExpand []TaskCollectionExpandable
		description string
	}{
		{
			name:        "CSV format comments",
			taskID:      1,
			csvExpand:   "comments",
			arrayExpand: []TaskCollectionExpandable{},
			description: "Test CSV format expand for task ReadOne",
		},
		{
			name:        "array format comments",
			taskID:      1,
			csvExpand:   "",
			arrayExpand: []TaskCollectionExpandable{TaskCollectionExpandComments},
			description: "Test array format expand for task ReadOne (backward compatibility)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Task with different expand parameter formats
			task := &Task{
				ID:        tt.taskID,
				ExpandCSV: tt.csvExpand,
				Expand:    tt.arrayExpand,
			}

			// Test that ReadOne processes expand parameters correctly
			err := task.ReadOne(s, u)
			require.NoError(t, err, "ReadOne should not fail for %s", tt.description)
			assert.NotZero(t, task.ID, "Task should have valid ID for %s", tt.description)
			assert.NotEmpty(t, task.Title, "Task should have title for %s", tt.description)

			// The actual expansion verification (checking if comments are loaded) would require
			// specific fixtures with comments. This test primarily verifies parameter parsing works.
		})
	}
}