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
	"strings"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

const (
	taskRelationsTable20261124192627      = "task_relations"
	taskRelationTaskKindIndexName20261124 = "IDX_task_relations_task_kind"
)

type taskRelationTaskKindIndex20261124192627 struct {
	TaskID       int64               `xorm:"bigint not null index(task_kind)"`
	RelationKind models.RelationKind `xorm:"varchar(50) not null index(task_kind)"`
}

func (taskRelationTaskKindIndex20261124192627) TableName() string {
	return taskRelationsTable20261124192627
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20261124192627",
		Description: "Add an index on task_relations (task_id, relation_kind) for subtask expansion",
		Migrate:     addTaskRelationTaskKindIndex,
		Rollback: func(tx *xorm.Engine) error {
			// No-op: the migration may have reused a pre-existing user-created index,
			// so dropping it on rollback could unintentionally remove user indexes.
			return nil
		},
	})
}

func addTaskRelationTaskKindIndex(tx *xorm.Engine) error {
	dbType := tx.Dialect().URI().DBType

	// Check if a matching index already exists or if a name conflict prevents creation.
	reused, err := checkExistingTaskRelationIndexes(tx, dbType)
	if err != nil {
		return err
	}
	if reused {
		return nil
	}

	// On MySQL/MariaDB, verify the column type can be indexed without a prefix length.
	if dbType == schemas.MYSQL {
		canIndex, err := canCreateMySQLTaskRelationIndex(tx)
		if err != nil {
			return err
		}
		if !canIndex {
			return nil
		}
	}

	if err := tx.CreateIndexes(taskRelationTaskKindIndex20261124192627{}); err != nil {
		return fmt.Errorf("could not create index %s: %w", taskRelationTaskKindIndexName20261124, err)
	}
	return nil
}

// checkExistingTaskRelationIndexes inspects current indexes on task_relations.
// It returns (true, nil) if an equivalent index already exists and can be reused,
// or an error if an index with the target name exists with a different definition.
func checkExistingTaskRelationIndexes(tx *xorm.Engine, dbType schemas.DBType) (bool, error) {
	indexes, err := inspectTaskRelationIndexes(tx, dbType)
	if err != nil {
		return false, fmt.Errorf("could not inspect %s indexes: %w", taskRelationsTable20261124192627, err)
	}

	hasEquivalent := false
	for name, isEquivalent := range indexes {
		// PostgreSQL treats unquoted identifiers as lowercase while quoted ones are case-sensitive;
		// MySQL and SQLite treat index names case-insensitively.
		isSameName := name == taskRelationTaskKindIndexName20261124 ||
			(dbType != schemas.POSTGRES && strings.EqualFold(name, taskRelationTaskKindIndexName20261124))

		if isSameName && !isEquivalent {
			return false, fmt.Errorf(
				"index %s on %s already exists with a different definition; rename it before upgrading",
				name,
				taskRelationsTable20261124192627,
			)
		}
		if isEquivalent {
			hasEquivalent = true
		}
	}

	return hasEquivalent, nil
}

// canCreateMySQLTaskRelationIndex ensures relation_kind is not an unindexable TEXT/BLOB column.
// Legacy databases migrated from SQLite may still store relation_kind as TEXT, which MySQL
// cannot fully index without specifying a key length (See #4374).
func canCreateMySQLTaskRelationIndex(tx *xorm.Engine) (bool, error) {
	rows, err := tx.QueryString(`
		SELECT DATA_TYPE AS data_type
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		  AND TABLE_NAME = 'task_relations'
		  AND COLUMN_NAME = 'relation_kind'
	`)
	if err != nil {
		return false, fmt.Errorf("could not read the type of task_relations.relation_kind: %w", err)
	}

	if len(rows) > 0 && mysqlUnindexableTypes20260720120000[strings.ToUpper(rows[0]["data_type"])] {
		log.Warningf(
			"Could not create index %s: MySQL cannot index task_relations.relation_kind without a key length. "+
				"Change it to VARCHAR(50), then create the index on (task_id, relation_kind).",
			taskRelationTaskKindIndexName20261124,
		)
		return false, nil
	}

	return true, nil
}

// inspectTaskRelationIndexes queries database-specific system catalogs to determine
// whether each index on task_relations is structurally equivalent to a non-unique,
// full-table composite index on (task_id, relation_kind).
// Note: We query system catalogs directly because xorm compares index columns as an
// unordered set and omits crucial metadata such as partial predicates or column order.
func inspectTaskRelationIndexes(tx *xorm.Engine, dbType schemas.DBType) (map[string]bool, error) {
	var query string

	switch dbType {
	case schemas.SQLITE:
		// Verify: non-unique, non-partial, exact 2-column match in order with default BINARY collation.
		query = `
			SELECT
				il.name,
				CASE WHEN il."unique" = 0
					AND il.partial = 0
					AND COUNT(*) = 2
					AND SUM(CASE
						WHEN ii.coll = 'BINARY' AND (
							(ii.seqno = 0 AND ii.name = 'task_id') OR
							(ii.seqno = 1 AND ii.name = 'relation_kind')
						) THEN 1 ELSE 0
					END) = 2
				THEN 1 ELSE 0 END AS equivalent
			FROM pragma_index_list('task_relations') il
			JOIN pragma_index_xinfo(il.name) ii ON ii."key" = 1
			GROUP BY il.name, il."unique", il.partial
		`
	case schemas.POSTGRES:
		// Verify: B-tree, non-unique, valid/ready/live (not a failed concurrent build),
		// no partial WHERE predicate, and exact 2 columns in (task_id, relation_kind) order.
		query = `
			SELECT
				idx.relname AS name,
				CASE WHEN am.amname = 'btree'
					AND NOT i.indisunique
					AND i.indisvalid
					AND i.indisready
					AND i.indislive
					AND i.indpred IS NULL
					AND i.indnatts = 2
					AND pg_get_indexdef(i.indexrelid, 1, true) = 'task_id'
					AND pg_get_indexdef(i.indexrelid, 2, true) = 'relation_kind'
				THEN 1 ELSE 0 END AS equivalent
			FROM pg_index i
			JOIN pg_class idx ON idx.oid = i.indexrelid
			JOIN pg_am am ON am.oid = idx.relam
			WHERE i.indrelid = 'task_relations'::regclass
		`
	case schemas.MYSQL:
		return inspectMySQLTaskRelationIndexes(tx)
	default:
		return nil, fmt.Errorf("unsupported database type %s", dbType)
	}

	rows, err := tx.QueryString(query)
	if err != nil {
		return nil, err
	}

	indexes := make(map[string]bool, len(rows))
	for _, row := range rows {
		indexes[row["name"]] = row["equivalent"] == "1"
	}
	return indexes, nil
}

type mysqlIndexColumn20261124192627 struct {
	seq        string
	columnName string
	valid      bool
}

// inspectMySQLTaskRelationIndexes parses SHOW INDEX output to detect equivalent composite indexes.
func inspectMySQLTaskRelationIndexes(tx *xorm.Engine) (map[string]bool, error) {
	rows, err := tx.QueryString("SHOW INDEX FROM task_relations")
	if err != nil {
		return nil, err
	}

	columnsByIndex := make(map[string][]mysqlIndexColumn20261124192627)
	for _, row := range rows {
		// MySQL reports index visibility in "Visible", whereas MariaDB reports it in "Ignored".
		isValid := row["Non_unique"] == "1" &&
			row["Index_type"] == "BTREE" &&
			row["Sub_part"] == "" &&
			row["Visible"] != "NO" &&
			row["Ignored"] != "YES"

		name := row["Key_name"]
		columnsByIndex[name] = append(columnsByIndex[name], mysqlIndexColumn20261124192627{
			seq:        row["Seq_in_index"],
			columnName: row["Column_name"],
			valid:      isValid,
		})
	}

	indexes := make(map[string]bool, len(columnsByIndex))
	for name, cols := range columnsByIndex {
		indexes[name] = isEquivalentMySQLCompositeIndex(cols)
	}
	return indexes, nil
}

// isEquivalentMySQLCompositeIndex checks whether the index consists of exactly
// two valid columns: (1: task_id, 2: relation_kind).
func isEquivalentMySQLCompositeIndex(cols []mysqlIndexColumn20261124192627) bool {
	if len(cols) != 2 {
		return false
	}

	hasTaskIDFirst := false
	hasRelationKindSecond := false

	for _, col := range cols {
		if !col.valid {
			return false
		}
		if col.seq == "1" && col.columnName == "task_id" {
			hasTaskIDFirst = true
		}
		if col.seq == "2" && col.columnName == "relation_kind" {
			hasRelationKindSecond = true
		}
	}

	return hasTaskIDFirst && hasRelationKindSecond
}
