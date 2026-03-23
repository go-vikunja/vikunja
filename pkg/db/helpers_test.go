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

package db

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

// isParadeDB returns true when the test engine is running with ParadeDB.
// It is safe to call even when no database engine is initialized (x == nil),
// in which case it returns false.
func isParadeDB() bool {
	return x != nil && ParadeDBAvailable()
}

func TestMultiFieldSearchSingleField(t *testing.T) {
	if x == nil {
		t.Skip("requires initialized database engine")
	}

	cond := MultiFieldSearchWithTableAlias([]string{"title"}, "landing", "")

	w := builder.NewWriter()
	err := cond.WriteTo(w)
	require.NoError(t, err)

	if isParadeDB() {
		assert.Equal(t, "title ||| ?::pdb.fuzzy(1, t)", w.String())
		assert.Equal(t, []interface{}{"landing"}, w.Args())
	} else {
		assert.Contains(t, w.String(), "title")
		assert.Contains(t, w.String(), "LIKE")
		assert.Equal(t, []interface{}{"%landing%"}, w.Args())
	}
}

func TestMultiFieldSearchMultiField(t *testing.T) {
	if x == nil {
		t.Skip("requires initialized database engine")
	}

	cond := MultiFieldSearchWithTableAlias([]string{"title", "description"}, "landing", "")

	w := builder.NewWriter()
	err := cond.WriteTo(w)
	require.NoError(t, err)

	if isParadeDB() {
		assert.Equal(t, "(title ||| ?::pdb.fuzzy(1, t)) OR (description ||| ?::pdb.fuzzy(1, t))", w.String())
		assert.Equal(t, []interface{}{"landing", "landing"}, w.Args())
	} else {
		assert.Contains(t, w.String(), "title")
		assert.Contains(t, w.String(), "description")
		assert.Contains(t, w.String(), "LIKE")
		assert.Equal(t, []interface{}{"%landing%", "%landing%"}, w.Args())
	}
}

func TestMultiFieldSearchWithTableAlias(t *testing.T) {
	if x == nil {
		t.Skip("requires initialized database engine")
	}

	cond := MultiFieldSearchWithTableAlias([]string{"title"}, "test", "tasks")

	w := builder.NewWriter()
	err := cond.WriteTo(w)
	require.NoError(t, err)

	if isParadeDB() {
		assert.Equal(t, "tasks.title ||| ?::pdb.fuzzy(1, t)", w.String())
		assert.Equal(t, []interface{}{"test"}, w.Args())
	} else {
		assert.Contains(t, w.String(), "tasks.title")
		assert.Contains(t, w.String(), "LIKE")
		assert.Equal(t, []interface{}{"%test%"}, w.Args())
	}
}

func TestMultiFieldSearchMultiFieldWithTableAlias(t *testing.T) {
	if x == nil {
		t.Skip("requires initialized database engine")
	}

	cond := MultiFieldSearchWithTableAlias([]string{"title", "description"}, "test", "tasks")

	w := builder.NewWriter()
	err := cond.WriteTo(w)
	require.NoError(t, err)

	if isParadeDB() {
		assert.Equal(t, "(tasks.title ||| ?::pdb.fuzzy(1, t)) OR (tasks.description ||| ?::pdb.fuzzy(1, t))", w.String())
		assert.Equal(t, []interface{}{"test", "test"}, w.Args())
	} else {
		assert.Contains(t, w.String(), "tasks.title")
		assert.Contains(t, w.String(), "tasks.description")
		assert.Contains(t, w.String(), "LIKE")
		assert.Equal(t, []interface{}{"%test%", "%test%"}, w.Args())
	}
}

func TestIsMySQLDuplicateEntryError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		constraintName string
		expected       bool
	}{
		{
			name:           "nil error",
			err:            nil,
			constraintName: "UQE_task_buckets_task_project_view",
			expected:       false,
		},
		{
			name:           "matching MySQL duplicate entry error",
			err:            errors.New("Error 1062 (23000): Duplicate entry '424-557' for key 'UQE_task_buckets_task_project_view'"),
			constraintName: "UQE_task_buckets_task_project_view",
			expected:       true,
		},
		{
			name:           "MySQL duplicate entry error with different constraint",
			err:            errors.New("Error 1062 (23000): Duplicate entry '424-557' for key 'some_other_constraint'"),
			constraintName: "UQE_task_buckets_task_project_view",
			expected:       false,
		},
		{
			name:           "non-MySQL error",
			err:            errors.New("some other database error"),
			constraintName: "UQE_task_buckets_task_project_view",
			expected:       false,
		},
		{
			name:           "case insensitive matching",
			err:            errors.New("ERROR 1062 (23000): DUPLICATE ENTRY '424-557' FOR KEY 'UQE_TASK_BUCKETS_TASK_PROJECT_VIEW'"),
			constraintName: "uqe_task_buckets_task_project_view",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMySQLDuplicateEntryError(tt.err, tt.constraintName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsUniqueConstraintError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		constraintName string
		expected       bool
	}{
		{
			name:           "nil error",
			err:            nil,
			constraintName: "UQE_task_buckets_task_project_view",
			expected:       false,
		},
		// MySQL tests
		{
			name:           "MySQL duplicate entry error",
			err:            errors.New("Error 1062 (23000): Duplicate entry '424-557' for key 'UQE_task_buckets_task_project_view'"),
			constraintName: "UQE_task_buckets_task_project_view",
			expected:       true,
		},
		{
			name:           "MySQL duplicate entry error with different constraint",
			err:            errors.New("Error 1062 (23000): Duplicate entry '424-557' for key 'some_other_constraint'"),
			constraintName: "UQE_task_buckets_task_project_view",
			expected:       false,
		},
		// PostgreSQL tests
		{
			name:           "PostgreSQL duplicate key error",
			err:            errors.New(`duplicate key value violates unique constraint "UQE_task_buckets_task_project_view"`),
			constraintName: "UQE_task_buckets_task_project_view",
			expected:       true,
		},
		{
			name:           "PostgreSQL duplicate key error with different constraint",
			err:            errors.New(`duplicate key value violates unique constraint "some_other_constraint"`),
			constraintName: "UQE_task_buckets_task_project_view",
			expected:       false,
		},
		// SQLite tests
		{
			name:           "SQLite unique constraint failed",
			err:            errors.New("UNIQUE constraint failed: task_buckets.task_id, task_buckets.project_view_id"),
			constraintName: "UQE_task_buckets_task_project_view",
			expected:       true,
		},
		{
			name:           "SQLite constraint failed with unique",
			err:            errors.New("constraint failed: UNIQUE constraint failed: task_buckets.task_id"),
			constraintName: "UQE_task_buckets_task_project_view",
			expected:       true, // Should match on task_buckets table pattern
		},
		// General tests
		{
			name:           "non-constraint error",
			err:            errors.New("some other database error"),
			constraintName: "UQE_task_buckets_task_project_view",
			expected:       false,
		},
		{
			name:           "case insensitive matching - MySQL",
			err:            errors.New("ERROR 1062 (23000): DUPLICATE ENTRY '424-557' FOR KEY 'UQE_TASK_BUCKETS_TASK_PROJECT_VIEW'"),
			constraintName: "uqe_task_buckets_task_project_view",
			expected:       true,
		},
		{
			name:           "case insensitive matching - PostgreSQL",
			err:            errors.New(`DUPLICATE KEY VALUE VIOLATES UNIQUE CONSTRAINT "UQE_TASK_BUCKETS_TASK_PROJECT_VIEW"`),
			constraintName: "uqe_task_buckets_task_project_view",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUniqueConstraintError(tt.err, tt.constraintName)
			assert.Equal(t, tt.expected, result)
		})
	}
}
