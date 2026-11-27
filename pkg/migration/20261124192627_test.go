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
	"fmt"
	"slices"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"

	"github.com/stretchr/testify/require"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

// TestTaskRelationCompositeIndex verifies that the current TaskRelation model schema
// defines the expected composite indexes via xorm struct tags.
func TestTaskRelationCompositeIndex(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	table := &models.TaskRelation{}
	require.NoError(t, x.DropTables(table))
	t.Cleanup(func() { require.NoError(t, x.DropTables(table)) })

	require.NoError(t, x.Sync2(table))

	indexes := taskRelationsTable20261103165926(t, x).Indexes
	expectedIndexes := map[string][]string{
		"task_kind":       {"task_id", "relation_kind"},
		"other_task_kind": {"other_task_id", "relation_kind"},
	}

	for name, cols := range expectedIndexes {
		require.Contains(t, indexes, name)
		require.Equal(t, schemas.IndexType, indexes[name].Type)
		require.Equal(t, cols, indexes[name].Cols)
	}
}

// TestTaskRelationIndexMigration verifies equivalent indexes are reused and conflicting
// indexes are rejected without losing existing indexes or task relations.
func TestTaskRelationIndexMigration(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	dbType := x.Dialect().URI().DBType
	quotedDefaultIndex := x.Quote(taskRelationTaskKindIndexName20261124)

	type testCase struct {
		name          string
		setupSQLs     []string
		expectedIndex string
		expectErr     string
	}

	tests := []testCase{
		{
			name:          "creates default index when none exists",
			expectedIndex: taskRelationTaskKindIndexName20261124,
		},
		{
			name:          "reuses existing equivalent custom index",
			setupSQLs:     []string{"CREATE INDEX custom_task_kind ON task_relations (task_id, relation_kind)"},
			expectedIndex: "custom_task_kind",
		},
		{
			name:          "ignores index with reversed column order and creates default",
			setupSQLs:     []string{"CREATE INDEX custom_task_kind ON task_relations (relation_kind, task_id)"},
			expectedIndex: taskRelationTaskKindIndexName20261124,
		},
		{
			name:          "ignores unique index and creates default",
			setupSQLs:     []string{"CREATE UNIQUE INDEX custom_task_kind ON task_relations (task_id, relation_kind)"},
			expectedIndex: taskRelationTaskKindIndexName20261124,
		},
		{
			name: "fails on name conflict even when another equivalent index exists",
			setupSQLs: []string{
				"CREATE INDEX custom_task_kind ON task_relations (task_id, relation_kind)",
				fmt.Sprintf("CREATE INDEX %s ON task_relations (relation_kind, task_id)", quotedDefaultIndex),
			},
			expectErr: "IDX_task_relations_task_kind on task_relations already exists with a different definition",
		},
	}

	switch dbType {
	case schemas.SQLITE, schemas.POSTGRES:
		tests = append(tests, testCase{
			name:          "ignores partial index and creates default",
			setupSQLs:     []string{"CREATE INDEX custom_task_kind ON task_relations (task_id, relation_kind) WHERE task_id = 1"},
			expectedIndex: taskRelationTaskKindIndexName20261124,
		})
	case schemas.MYSQL:
		tests = append(tests, testCase{
			name:          "ignores prefix index and creates default",
			setupSQLs:     []string{"CREATE INDEX custom_task_kind ON task_relations (task_id, relation_kind(4))"},
			expectedIndex: taskRelationTaskKindIndexName20261124,
		})
	}

	migrationIdx := slices.IndexFunc(migrations, func(m *xormigrate.Migration) bool {
		return m.ID == "20261124192627"
	})
	require.NotEqual(t, -1, migrationIdx)
	targetMigration := migrations[migrationIdx]

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupLegacyTaskRelationsTable(t, x, true)
			require.NoError(t, addTaskRelationsOtherTaskKindIndex20261103165926(x))

			// Seed a row to ensure index creation and rollback preserve existing data.
			seed := taskRelationsBefore20261103165926{
				TaskID:       1,
				OtherTaskID:  2,
				RelationKind: "subtask",
				CreatedByID:  3,
			}
			_, err := x.Insert(&seed)
			require.NoError(t, err)

			for _, ddl := range tt.setupSQLs {
				_, err := x.Exec(ddl)
				require.NoError(t, err)
			}

			beforeIndexes, err := inspectTaskRelationIndexes(x, dbType)
			require.NoError(t, err)

			m := xormigrate.New(x, []*xormigrate.Migration{targetMigration})
			err = m.Migrate()

			if tt.expectErr != "" {
				require.ErrorContains(t, err, tt.expectErr)
				afterIndexes, err := inspectTaskRelationIndexes(x, dbType)
				require.NoError(t, err)
				require.Equal(t, beforeIndexes, afterIndexes)
				return
			}

			// Verify migration is idempotent and rollback intentionally leaves the index intact.
			require.NoError(t, err)
			require.NoError(t, m.Migrate())
			require.NoError(t, m.RollbackLast())
			require.NoError(t, m.Migrate())

			beforeIndexes[tt.expectedIndex] = true
			afterIndexes, err := inspectTaskRelationIndexes(x, dbType)
			require.NoError(t, err)
			require.Equal(t, beforeIndexes, afterIndexes)

			// Verify seeded data remains intact.
			var got taskRelationsBefore20261103165926
			found, err := x.ID(seed.ID).Get(&got)
			require.NoError(t, err)
			require.True(t, found)
			require.WithinDuration(t, seed.Created, got.Created, time.Second)
			got.Created = seed.Created
			require.Equal(t, seed, got)
		})
	}
}

// TestTaskRelationIndexMigrationMySQLText verifies that legacy MySQL tables with a TEXT
// relation_kind column log a warning and skip index creation without blocking startup.
func TestTaskRelationIndexMigrationMySQLText(t *testing.T) {
	x, err := db.CreateTestEngine()
	require.NoError(t, err)

	dbType := x.Dialect().URI().DBType
	if dbType != schemas.MYSQL {
		t.Skip("MySQL-specific key length restriction")
	}

	setupLegacyTaskRelationsTable(t, x, false)
	_, err = x.Exec("ALTER TABLE task_relations MODIFY relation_kind TEXT NOT NULL")
	require.NoError(t, err)

	require.NoError(t, addTaskRelationTaskKindIndex(x))

	indexes, err := inspectTaskRelationIndexes(x, dbType)
	require.NoError(t, err)
	require.NotContains(t, indexes, taskRelationTaskKindIndexName20261124)
}

// setupLegacyTaskRelationsTable recreates the legacy table before adding either composite index.
func setupLegacyTaskRelationsTable(t *testing.T, x *xorm.Engine, withMigrationTable bool) {
	t.Helper()

	table := taskRelationsBefore20261103165926{}
	tables := []any{table}
	if withMigrationTable {
		tables = append(tables, &xormigrate.Migration{})
	}

	require.NoError(t, x.DropTables(tables...))
	t.Cleanup(func() { require.NoError(t, x.DropTables(tables...)) })

	require.NoError(t, x.Sync2(table))
}
