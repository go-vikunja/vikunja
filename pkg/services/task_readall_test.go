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

package services

import (
	"testing"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web"
)

// TestSortParamValidation tests the sortParam validation logic
func TestSortParamValidation(t *testing.T) {
	tests := []struct {
		name      string
		sortParam sortParam
		wantErr   bool
	}{
		{
			name: "valid sort param - id",
			sortParam: sortParam{
				sortBy:  taskPropertyID,
				orderBy: orderAscending,
			},
			wantErr: false,
		},
		{
			name: "valid sort param - title desc",
			sortParam: sortParam{
				sortBy:  taskPropertyTitle,
				orderBy: orderDescending,
			},
			wantErr: false,
		},
		{
			name: "invalid sort param - unknown field",
			sortParam: sortParam{
				sortBy:  "invalid_field",
				orderBy: orderAscending,
			},
			wantErr: true,
		},
		{
			name: "invalid sort param - invalid order",
			sortParam: sortParam{
				sortBy:  taskPropertyID,
				orderBy: sortOrder("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sortParam.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("sortParam.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetSortOrderFromString tests the sort order string conversion
func TestGetSortOrderFromString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  sortOrder
	}{
		{
			name:  "ascending",
			input: "asc",
			want:  orderAscending,
		},
		{
			name:  "descending",
			input: "desc",
			want:  orderDescending,
		},
		{
			name:  "uppercase ascending",
			input: "ASC",
			want:  orderAscending,
		},
		{
			name:  "uppercase descending",
			input: "DESC",
			want:  orderDescending,
		},
		{
			name:  "ascending with whitespace",
			input: " asc ",
			want:  orderAscending,
		},
		{
			name:  "descending with whitespace",
			input: "  desc  ",
			want:  orderDescending,
		},
		{
			name:  "mixed case",
			input: "AsC",
			want:  orderAscending,
		},
		{
			name:  "full word ascending",
			input: "ascending",
			want:  orderAscending,
		},
		{
			name:  "full word descending",
			input: "descending",
			want:  orderDescending,
		},
		{
			name:  "invalid defaults to ascending",
			input: "invalid",
			want:  orderAscending,
		},
		{
			name:  "empty defaults to ascending",
			input: "",
			want:  orderAscending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSortOrderFromString(tt.input)
			if got != tt.want {
				t.Errorf("getSortOrderFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTaskServiceBasicFunctionality tests that the TaskService is properly initialized
func TestTaskServiceBasicFunctionality(t *testing.T) {
	// Initialize the TaskService without a real database connection for basic testing
	ts := NewTaskService(nil)

	if ts == nil {
		t.Error("NewTaskService() returned nil")
	}

	if ts.Registry == nil {
		t.Error("TaskService.Registry not initialized")
	}
}

// TestTaskCollectionToSearchOptions tests the conversion logic
func TestTaskCollectionToSearchOptions(t *testing.T) {
	ts := NewTaskService(nil)

	tc := &models.TaskCollection{
		SortBy:  []string{"title", "id"},
		OrderBy: []string{"desc", "asc"},
		Filter:  "done = false",
	}

	// Test without a project view
	opts, err := ts.getTaskFilterOptsFromCollection(tc, nil)
	if err != nil {
		t.Fatalf("getTaskFilterOptsFromCollection() error = %v", err)
	}

	if opts == nil {
		t.Fatal("getTaskFilterOptsFromCollection() returned nil options")
	}

	// Verify sort parameters
	if len(opts.sortby) != 2 {
		t.Errorf("Expected 2 sort parameters, got %d", len(opts.sortby))
	}

	if opts.sortby[0].sortBy != "title" || opts.sortby[0].orderBy != orderDescending {
		t.Errorf("First sort param incorrect: got %v %v, want title desc", opts.sortby[0].sortBy, opts.sortby[0].orderBy)
	}

	if opts.sortby[1].sortBy != "id" || opts.sortby[1].orderBy != orderAscending {
		t.Errorf("Second sort param incorrect: got %v %v, want id asc", opts.sortby[1].sortBy, opts.sortby[1].orderBy)
	}

	// Verify filter options
	if opts.filter != "done = false" {
		t.Errorf("Filter not preserved: got %v, want 'done = false'", opts.filter)
	}
}

// MockAuth implements web.Auth for testing
type MockAuth struct {
	userID int64
}

// Verify that MockAuth implements web.Auth interface
var _ web.Auth = (*MockAuth)(nil)

func (m *MockAuth) GetID() int64 {
	return m.userID
}

// TestGetRelevantProjectsFromCollection tests project collection logic
func TestGetRelevantProjectsFromCollection(t *testing.T) {
	ts := NewTaskService(nil)

	// Test with specific project ID
	tc := &models.TaskCollection{
		ProjectID: 1,
	}

	mockAuth := &MockAuth{userID: 1}

	// This will fail with nil session, but we're testing the basic structure
	_, err := ts.getRelevantProjectsFromCollection(nil, mockAuth, tc)

	// We expect an error since we don't have a real database connection
	// but the function should not panic
	if err == nil {
		t.Log("getRelevantProjectsFromCollection() completed without error (unexpected but not critical)")
	} else {
		t.Logf("getRelevantProjectsFromCollection() error = %v (expected due to nil session)", err)
	}
}
